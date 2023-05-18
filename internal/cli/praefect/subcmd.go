package praefect

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/sirupsen/logrus"
	gitalyauth "gitlab.com/gitlab-org/gitaly/v16/auth"
	"gitlab.com/gitlab-org/gitaly/v16/client"
	internalclient "gitlab.com/gitlab-org/gitaly/v16/internal/gitaly/client"
	"gitlab.com/gitlab-org/gitaly/v16/internal/praefect/config"
	"gitlab.com/gitlab-org/gitaly/v16/internal/praefect/datastore/glsql"
	"google.golang.org/grpc"
)

type subcmd interface {
	FlagSet() *flag.FlagSet
	Exec(flags *flag.FlagSet, config config.Config) error
}

const (
	defaultDialTimeout        = 10 * time.Second
	paramVirtualStorage       = "virtual-storage"
	paramRelativePath         = "repository"
	paramAuthoritativeStorage = "authoritative-storage"
)

func subcommands(logger *logrus.Entry) map[string]subcmd {
	return map[string]subcmd{
		sqlPingCmdName:              &sqlPingSubcommand{},
		sqlMigrateDownCmdName:       &sqlMigrateDownSubcommand{},
		sqlMigrateStatusCmdName:     &sqlMigrateStatusSubcommand{},
		setReplicationFactorCmdName: newSetReplicatioFactorSubcommand(os.Stdout),
		trackRepositoriesCmdName:    newTrackRepositories(logger, os.Stdout),
	}
}

// subCommand returns an exit code, to be fed into os.Exit.
func subCommand(conf config.Config, logger *logrus.Entry, arg0 string, argRest []string) int {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	go func() {
		<-interrupt
		os.Exit(130) // indicates program was interrupted
	}()

	subcmd, ok := subcommands(logger)[arg0]
	if !ok {
		printfErr("%s: unknown subcommand: %q\n", progname, arg0)
		return 1
	}

	flags := subcmd.FlagSet()

	if err := flags.Parse(argRest); err != nil {
		printfErr("%s\n", err)
		return 1
	}

	if err := subcmd.Exec(flags, conf); err != nil {
		printfErr("%s\n", err)
		return 1
	}

	return 0
}

func getNodeAddress(cfg config.Config) (string, error) {
	switch {
	case cfg.SocketPath != "":
		return "unix:" + cfg.SocketPath, nil
	case cfg.ListenAddr != "":
		return "tcp://" + cfg.ListenAddr, nil
	case cfg.TLSListenAddr != "":
		return "tls://" + cfg.TLSListenAddr, nil
	default:
		return "", errors.New("no Praefect address configured")
	}
}

func openDB(conf config.DB) (*sql.DB, func(), error) {
	ctx := context.Background()

	openDBCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	db, err := glsql.OpenDB(openDBCtx, conf)
	if err != nil {
		return nil, nil, fmt.Errorf("sql open: %w", err)
	}

	clean := func() {
		if err := db.Close(); err != nil {
			printfErr("sql close: %v\n", err)
		}
	}

	return db, clean, nil
}

func printfErr(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, format, a...)
}

func subCmdDial(ctx context.Context, addr, token string, timeout time.Duration, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	opts = append(opts,
		grpc.WithBlock(),
		internalclient.UnaryInterceptor(),
		internalclient.StreamInterceptor(),
	)

	if len(token) > 0 {
		opts = append(opts,
			grpc.WithPerRPCCredentials(
				gitalyauth.RPCCredentialsV2(token),
			),
		)
	}

	return client.DialContext(ctx, addr, opts)
}

type requiredParameterError string

func (p requiredParameterError) Error() string {
	return fmt.Sprintf("%q is a required parameter", string(p))
}

type unexpectedPositionalArgsError struct{ Command string }

func (err unexpectedPositionalArgsError) Error() string {
	return fmt.Sprintf("%s doesn't accept positional arguments", err.Command)
}
