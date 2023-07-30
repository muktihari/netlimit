package http_test

import (
	"context"
	"errors"
	"net"
	"net/http"
	"testing"

	"github.com/muktihari/netlimit"
	httplimit "github.com/muktihari/netlimit/http"
	"golang.org/x/time/rate"
)

func TestNewTransport(t *testing.T) {
	var (
		limit        = 1024
		readLimiter  = rate.NewLimiter(rate.Limit(limit), limit)
		writeLimiter = rate.NewLimiter(rate.Limit(limit), limit)
		transport    = httplimit.NewTransport(&http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return &net.TCPConn{}, nil // dial context stub
			},
		}, readLimiter, writeLimiter)
	)

	conn, err := transport.DialContext(nil, "", "") // invoke to get the conn
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	switch c := conn.(type) {
	case *netlimit.Conn:
		if readLimiter != c.ReadLimiter {
			t.Fatalf("expected: %v, got: %v", readLimiter, c.ReadLimiter)
		}
		if writeLimiter != c.WriteLimiter {
			t.Fatalf("expected: %v, got: %v", writeLimiter, c.WriteLimiter)
		}
	default:
		t.Fatalf("expected conn type is *netlimit.Conn, got: %T", conn)
	}
}

func TestNewTransportNil(t *testing.T) {
	transport := httplimit.NewTransport(nil, nil, nil)
	if transport == nil {
		t.Fatalf("expected %T, got: %T", http.DefaultTransport, transport)
	}
}

func TestNewTransportDialFail(t *testing.T) {
	expectedError := errors.New("any dial error")
	transport := httplimit.NewTransport(&http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return nil, expectedError
		},
	}, nil, nil)

	_, err := transport.DialContext(nil, "", "")
	if err != expectedError {
		t.Fatalf("expected %v, got: %v", expectedError, err)
	}
}
