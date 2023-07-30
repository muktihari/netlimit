package netlimit_test

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/muktihari/netlimit"
	"golang.org/x/time/rate"
)

type connStub struct {
	readFunc  func(b []byte) (int, error)
	writeFunc func(p []byte) (int, error)
}

var _ net.Conn = connStub{} // should satisfy net.Conn interface

func (cs connStub) Read(b []byte) (int, error) { return cs.readFunc(b) }

func (cs connStub) Write(p []byte) (int, error) { return cs.writeFunc(p) }

// remaining implementations that unused in this context
func (cs connStub) Close() error                       { return nil }
func (cs connStub) LocalAddr() net.Addr                { return nil }
func (cs connStub) RemoteAddr() net.Addr               { return nil }
func (cs connStub) SetDeadline(t time.Time) error      { return nil }
func (cs connStub) SetReadDeadline(t time.Time) error  { return nil }
func (cs connStub) SetWriteDeadline(t time.Time) error { return nil }

type limiterStub struct {
	limit     rate.Limit
	waitNFunc func(ctx context.Context, n int) error
}

var _ netlimit.Limiter = limiterStub{} // should satisfy Limiter interface

func (ls limiterStub) Limit() rate.Limit { return ls.limit }

func (ls limiterStub) WaitN(ctx context.Context, n int) error { return ls.waitNFunc(ctx, n) }

func TestRead(t *testing.T) {
	tt := []struct {
		name          string
		size          int
		limit         int
		connErr       error
		limiterErr    error
		expectedN     int
		expectedParts []int
	}{
		{
			name:          "without limiter, read success",
			size:          555,
			expectedN:     555,
			expectedParts: []int{},
		},
		{
			name:          "size > limit, read success - expected result",
			size:          555,
			limit:         100,
			expectedN:     555,
			expectedParts: []int{100, 100, 100, 100, 100, 55},
		},
		{
			name:          "size < limit, read success - expected result",
			size:          55,
			limit:         100,
			expectedN:     55,
			expectedParts: []int{55},
		},
		{
			name:          "read fail - connection closed",
			size:          555,
			limit:         100,
			connErr:       net.ErrClosed,
			expectedN:     0,
			expectedParts: []int{},
		},
		{
			name:          "read fail - limit exceed token",
			size:          555,
			limit:         100,
			limiterErr:    errors.New("limit exceed token"),
			expectedN:     100,
			expectedParts: []int{},
		},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var (
				content = make([]byte, tc.size)
				parts   = make([]int, 0)
			)

			for i := 0; i < len(content); i++ { // generate number in order
				content[i] = byte(i)
			}

			pos := 0
			conn := &netlimit.Conn{
				Conn: connStub{
					readFunc: func(b []byte) (n int, err error) {
						err = tc.connErr
						if err != nil {
							return 0, err
						}
						pos += copy(b, content[pos:pos+len(b)])
						return len(b), nil
					},
				},
				ReadLimiter: limiterStub{
					limit: rate.Limit(tc.limit),
					waitNFunc: func(ctx context.Context, n int) (err error) {
						err = tc.limiterErr
						if err != nil {
							return
						}
						parts = append(parts, n)
						return
					},
				},
			}

			if tc.limit <= 0 {
				conn.ReadLimiter = nil
			}

			expectedErr := tc.connErr
			if tc.connErr == nil {
				expectedErr = tc.limiterErr
			}

			b := make([]byte, len(content))
			n, err := conn.Read(b)
			if err != expectedErr {
				t.Fatalf("expected: %v, got: %v", expectedErr, err)
			}

			if n != tc.expectedN {
				t.Fatalf("expected n: %d; got: %d", tc.expectedN, n)
			}

			if diff := cmp.Diff(tc.expectedParts, parts); diff != "" {
				t.Fatalf("expected result: %v; got: %v", tc.expectedParts, parts)
			}

			expectedContent := make([]byte, len(content))
			for i := 0; i < pos; i++ {
				expectedContent[i] = content[i]
			}

			if diff := cmp.Diff(expectedContent, b); diff != "" {
				t.Fatalf("expected content of b: %v, got: %v", expectedContent, b)
			}
		})
	}
}

func TestWrite(t *testing.T) {
	tt := []struct {
		name          string
		size          int
		limit         int
		connErr       error
		limiterErr    error
		expectedN     int
		expectedParts []int
	}{
		{
			name:          "without limiter, write success",
			size:          555,
			limit:         0,
			expectedN:     555,
			expectedParts: []int{},
		},
		{
			name:          "size > limit, write success - expected result",
			size:          555,
			limit:         100,
			expectedN:     555,
			expectedParts: []int{100, 100, 100, 100, 100, 55},
		},
		{
			name:          "size < limit, write success - expected result",
			size:          55,
			limit:         100,
			expectedN:     55,
			expectedParts: []int{55},
		},
		{
			name:          "write fail - connection closed",
			size:          555,
			limit:         100,
			connErr:       net.ErrClosed,
			expectedN:     0,
			expectedParts: []int{},
		},
		{
			name:          "write fail - limit exceed token",
			size:          555,
			limit:         100,
			limiterErr:    errors.New("limit exceed token"),
			expectedN:     100,
			expectedParts: []int{},
		},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var (
				content = make([]byte, tc.size)
				parts   = make([]int, 0)
			)

			for i := 0; i < len(content); i++ { // generate number in order
				content[i] = byte(i)
			}

			pos := 0
			conn := &netlimit.Conn{
				Conn: connStub{
					writeFunc: func(p []byte) (n int, err error) {
						err = tc.connErr
						if err != nil {
							return 0, err
						}
						pos += copy(p, content[pos:pos+len(p)])
						return len(p), nil
					},
				},
				WriteLimiter: limiterStub{
					limit: rate.Limit(tc.limit),
					waitNFunc: func(ctx context.Context, n int) (err error) {
						err = tc.limiterErr
						if err != nil {
							return
						}
						parts = append(parts, n)
						return
					},
				},
			}

			if tc.limit <= 0 {
				conn.WriteLimiter = nil
			}

			expectedErr := tc.connErr
			if tc.connErr == nil {
				expectedErr = tc.limiterErr
			}

			p := make([]byte, len(content))
			n, err := conn.Write(p)
			if err != expectedErr {
				t.Fatalf("expected: %v, got: %v", expectedErr, err)
			}

			if n != tc.expectedN {
				t.Fatalf("expected n: %d; got: %d", tc.expectedN, n)
			}

			if diff := cmp.Diff(tc.expectedParts, parts); diff != "" {
				t.Fatalf("expected result: %v; got: %v", tc.expectedParts, parts)
			}

			expectedContent := make([]byte, len(content))
			for i := 0; i < pos; i++ {
				expectedContent[i] = content[i]
			}

			if diff := cmp.Diff(expectedContent, p); diff != "" {
				t.Fatalf("expected content of p: %v, got: %v", expectedContent, p)
			}
		})
	}
}
