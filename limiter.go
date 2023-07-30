package netlimit

import (
	"context"

	"golang.org/x/time/rate"
)

type Limiter interface {
	Limit() rate.Limit
	WaitN(context.Context, int) error
}
