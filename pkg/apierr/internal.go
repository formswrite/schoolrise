package apierr

import (
	"encore.dev/beta/errs"
	"encore.dev/rlog"
)

const internalMessage = "internal error"

func WrapInternal(op string, err error) error {
	if err == nil {
		return nil
	}
	rlog.Error("internal error", "op", op, "err", err)
	return &errs.Error{Code: errs.Internal, Message: internalMessage}
}

func Internal(err error) error {
	return WrapInternal("", err)
}
