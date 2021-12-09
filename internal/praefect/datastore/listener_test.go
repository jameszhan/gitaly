package datastore

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/gitlab-org/gitaly/v14/internal/praefect/config"
	"gitlab.com/gitlab-org/gitaly/v14/internal/praefect/datastore/glsql"
	"gitlab.com/gitlab-org/gitaly/v14/internal/testhelper"
	"gitlab.com/gitlab-org/gitaly/v14/internal/testhelper/testdb"
)

func TestNewListener(t *testing.T) {
	t.Parallel()

	t.Run("bad configuration", func(t *testing.T) {
		_, err := NewListener(config.DB{SSLMode: "invalid"})
		require.Error(t, err)
		require.Regexp(t, "connection config preparation:.*`sslmode=invalid.*`", err.Error())
	})
}

func TestListener_Listen(t *testing.T) {
	t.Parallel()
	db := testdb.New(t)

	lis, err := NewListener(testdb.GetConfig(t, db.Name))
	require.NoError(t, err)

	newChannel := func(i int) func() string {
		return func() string {
			i++
			return fmt.Sprintf("channel_%d", i)
		}
	}(0)

	notifyListener := func(t *testing.T, channels []string, payload string) {
		t.Helper()
		for _, channel := range channels {
			_, err := db.Exec(fmt.Sprintf(`NOTIFY %s, '%s'`, channel, payload))
			assert.NoError(t, err)
		}
	}

	waitFor := func(t *testing.T, c <-chan struct{}, d time.Duration) {
		t.Helper()
		select {
		case <-time.After(d):
			require.FailNow(t, "it takes too long")
		case <-c:
			// proceed
		}
	}

	listenNotify := func(t *testing.T, lis *Listener, channels []string, numNotifiers int, payloads []string) []string {
		t.Helper()

		start := make(chan struct{})
		go func() {
			<-start

			for i := 0; i < numNotifiers; i++ {
				go func() {
					for _, payload := range payloads {
						notifyListener(t, channels, payload)
					}
				}()
			}
		}()

		numResults := len(channels) * len(payloads) * numNotifiers
		result := make([]string, numResults)
		allReceivedChan := make(chan struct{})
		callback := func(idx int) func(n glsql.Notification) {
			return func(n glsql.Notification) {
				idx++
				result[idx] = n.Payload
				if idx+1 == numResults {
					close(allReceivedChan)
				}
			}
		}(-1)

		handler := mockListenHandler{OnNotification: callback, OnConnected: func() { close(start) }}
		ctx, cancel := testhelper.Context()
		defer cancel()
		allDone := make(chan struct{})
		go func() {
			waitFor(t, allReceivedChan, time.Minute)
			cancel()
			close(allDone)
		}()
		err := lis.Listen(ctx, handler, channels...)
		<-allDone
		assert.True(t, errors.Is(err, context.Canceled), err)
		return result
	}

	t.Run("listen on bad channel", func(t *testing.T) {
		ctx, cancel := testhelper.Context()
		defer cancel()
		err := lis.Listen(ctx, mockListenHandler{}, "bad channel")
		require.EqualError(t, err, `listen on channel(s): ERROR: syntax error at or near "channel" (SQLSTATE 42601)`)
	})

	t.Run("single listener and single notifier", func(t *testing.T) {
		channel := newChannel()
		payloads := []string{"this", "is", "a", "payload"}
		result := listenNotify(t, lis, []string{channel}, 1, payloads)
		require.Equal(t, payloads, result)
	})

	t.Run("single listener and multiple notifiers", func(t *testing.T) {
		channel := newChannel()

		const numNotifiers = 10

		payloads := []string{"this", "is", "a", "payload"}
		var expResult []string
		for i := 0; i < numNotifiers; i++ {
			expResult = append(expResult, payloads...)
		}

		result := listenNotify(t, lis, []string{channel}, numNotifiers, payloads)
		require.ElementsMatch(t, expResult, result, "there must be no additional data, only expected")
	})

	t.Run("listen multiple channels", func(t *testing.T) {
		channel1 := newChannel()
		channel2 := newChannel()

		result := listenNotify(t, lis, []string{channel1, channel2}, 1, []string{"payload"})
		require.Equal(t, []string{"payload", "payload"}, result)
	})

	t.Run("sequential Listen calls are allowed", func(t *testing.T) {
		channel1 := newChannel()
		result := listenNotify(t, lis, []string{channel1}, 1, []string{"payload-1"})
		require.Equal(t, []string{"payload-1"}, result)

		channel2 := newChannel()
		result2 := listenNotify(t, lis, []string{channel2}, 1, []string{"payload-2"})
		require.Equal(t, []string{"payload-2"}, result2)
	})

	t.Run("connection interruption", func(t *testing.T) {
		lis, err := NewListener(testdb.GetConfig(t, db.Name))
		require.NoError(t, err)

		channel := newChannel()

		connected := make(chan struct{})
		disconnected := make(chan struct{})
		handler := mockListenHandler{
			OnConnected:  func() { close(connected) },
			OnDisconnect: func(error) { close(disconnected) },
		}

		ctx, cancel := testhelper.Context()
		defer cancel()
		done := make(chan struct{})
		go func() {
			defer close(done)
			err := lis.Listen(ctx, handler, channel)
			var pgErr *pgconn.PgError
			if assert.True(t, errors.As(err, &pgErr)) {
				const adminShutdownCode = "57P01"
				assert.Equal(t, adminShutdownCode, pgErr.Code)
				assert.Equal(t, "FATAL", pgErr.Severity)
			}
		}()

		waitFor(t, connected, time.Minute)
		disconnectListener(t, db, channel)
		waitFor(t, disconnected, time.Minute)
		<-done
	})
}

func disconnectListener(t *testing.T, db testdb.DB, channel string) {
	t.Helper()
	res, err := db.Exec(
		`SELECT PG_TERMINATE_BACKEND(pid) FROM PG_STAT_ACTIVITY WHERE datname = $1 AND query = $2`,
		db.Name,
		"listen "+channel+";",
	)
	require.NoError(t, err)
	affected, err := res.RowsAffected()
	require.NoError(t, err)
	require.EqualValues(t, 1, affected)
}
