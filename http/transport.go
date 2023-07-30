package http

import (
	"context"
	"net"
	"net/http"

	"github.com/muktihari/netlimit"
)

// NewTransport creates new *http.Transport based on given t and wraps the underlying net.Conn
// to limit the network transfer rate. It might not precisely limit the transfer rate on exact time tick
// but should be close in wider/overall view.
//
// If t is nil, http.DefaultTransport will be used.
// readLimiter is limit to read as writeLimiter is to write, nil means no limit of each.
func NewTransport(t *http.Transport, readLimiter, writeLimiter netlimit.Limiter) *http.Transport {
	if t == nil {
		t = http.DefaultTransport.(*http.Transport)
	}

	t2 := t.Clone()
	t2.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		conn, err := t.DialContext(ctx, network, addr)
		if err != nil {
			return nil, err
		}

		return &netlimit.Conn{
			Conn:         conn,
			ReadLimiter:  readLimiter,
			WriteLimiter: writeLimiter,
		}, nil
	}

	return t2
}
