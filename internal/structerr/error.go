package structerr

import (
	"errors"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Error is a structured error that contains additional details.
type Error struct {
	err  error
	code codes.Code
}

func newError(code codes.Code, format string, a ...any) Error {
	for i, arg := range a {
		err, ok := arg.(error)
		if !ok {
			continue
		}

		if errors.As(err, &(Error{})) {
			// We need to explicitly handle this, otherwise `status.FromError()` would
			// return these because we implement `GRPCStatus()`.
			continue
		}

		// Convert any gRPC status we see here to an `Error`.
		if st, ok := status.FromError(err); ok {
			a[i] = Error{
				err:  errors.New(st.Message()),
				code: st.Code(),
			}
		}
	}

	formattedErr := fmt.Errorf(format, a...)

	// When we wrap an Error, we retain its error code. The intent of this is to retain the most
	// specific error code we have in the general case.
	var wrappedErr Error
	if errors.As(formattedErr, &wrappedErr) {
		code = wrappedErr.code
	}

	return Error{
		err:  formattedErr,
		code: code,
	}
}

// New returns a new Error with the default error code, which is Internal. When this function is
// used to wrap another Error, then the error code of that wrapped Error will be retained. The
// intent of this is to always retain the most specific error code in the general case.
func New(format string, a ...any) Error {
	return newError(codes.Internal, format, a...)
}

// NewAborted constructs a new error code with the Aborted error code. Please refer to New for
// further details.
func NewAborted(format string, a ...any) Error {
	return newError(codes.Aborted, format, a...)
}

// NewAlreadyExists constructs a new error code with the AlreadyExists error code. Please refer to
// New for further details.
func NewAlreadyExists(format string, a ...any) Error {
	return newError(codes.AlreadyExists, format, a...)
}

// NewCanceled constructs a new error code with the Canceled error code. Please refer to New for
// further details.
func NewCanceled(format string, a ...any) Error {
	return newError(codes.Canceled, format, a...)
}

// NewDataLoss constructs a new error code with the DataLoss error code. Please refer to New for
// further details.
func NewDataLoss(format string, a ...any) Error {
	return newError(codes.DataLoss, format, a...)
}

// NewDeadlineExceeded constructs a new error code with the DeadlineExceeded error code. Please
// refer to New for further details.
func NewDeadlineExceeded(format string, a ...any) Error {
	return newError(codes.DeadlineExceeded, format, a...)
}

// NewFailedPrecondition constructs a new error code with the FailedPrecondition error code. Please
// refer to New for further details.
func NewFailedPrecondition(format string, a ...any) Error {
	return newError(codes.FailedPrecondition, format, a...)
}

// NewInternal constructs a new error code with the Internal error code. Please refer to New for
// further details.
func NewInternal(format string, a ...any) Error {
	return newError(codes.Internal, format, a...)
}

// NewInvalidArgument constructs a new error code with the InvalidArgument error code. Please refer
// to New for further details.
func NewInvalidArgument(format string, a ...any) Error {
	return newError(codes.InvalidArgument, format, a...)
}

// NewNotFound constructs a new error code with the NotFound error code. Please refer to New for
// further details.
func NewNotFound(format string, a ...any) Error {
	return newError(codes.NotFound, format, a...)
}

// NewPermissionDenied constructs a new error code with the PermissionDenied error code. Please
// refer to New for further details.
func NewPermissionDenied(format string, a ...any) Error {
	return newError(codes.PermissionDenied, format, a...)
}

// NewResourceExhausted constructs a new error code with the ResourceExhausted error code. Please
// refer to New for further details.
func NewResourceExhausted(format string, a ...any) Error {
	return newError(codes.ResourceExhausted, format, a...)
}

// NewUnavailable constructs a new error code with the Unavailable error code. Please refer to New
// for further details.
func NewUnavailable(format string, a ...any) Error {
	return newError(codes.Unavailable, format, a...)
}

// NewUnauthenticated constructs a new error code with the Unauthenticated error code. Please refer
// to New for further details.
func NewUnauthenticated(format string, a ...any) Error {
	return newError(codes.Unauthenticated, format, a...)
}

// NewUnimplemented constructs a new error code with the Unimplemented error code. Please refer to
// New for further details.
func NewUnimplemented(format string, a ...any) Error {
	return newError(codes.Unimplemented, format, a...)
}

// NewUnknown constructs a new error code with the Unknown error code. Please refer to New for
// further details.
func NewUnknown(format string, a ...any) Error {
	return newError(codes.Unknown, format, a...)
}

// Error returns the error message of the Error.
func (e Error) Error() string {
	return e.err.Error()
}

// Unwrap returns the wrapped error if any, otherwise it returns nil.
func (e Error) Unwrap() error {
	return errors.Unwrap(e.err)
}

// Code returns the error code of the Error.
func (e Error) Code() codes.Code {
	return e.code
}

// GRPCStatus returns the gRPC status of this error.
func (e Error) GRPCStatus() *status.Status {
	return status.New(e.Code(), e.Error())
}
