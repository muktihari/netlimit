package netlimit

import (
	"context"
	"net"
)

// Conn is net.Conn embed-wrapper to rate limit data stream over the network in connection level.
type Conn struct {
	net.Conn

	ReadLimiter  Limiter
	WriteLimiter Limiter
}

// Read reads data from the connection and will wait if total n byte read exceed the limit.
func (lc *Conn) Read(b []byte) (n int, err error) {
	if lc.ReadLimiter == nil {
		return lc.Conn.Read(b)
	}

	limit := int64(lc.ReadLimiter.Limit())
	nb := int64(len(b))
	if nb < limit {
		limit = nb
	}

	for i := int64(0); i < nb; i += limit {
		end := i + limit
		if end > nb {
			end = nb
		}

		nr, err := lc.Conn.Read(b[i:end])
		n += nr
		if err != nil {
			return n, err
		}

		if err := lc.ReadLimiter.WaitN(context.Background(), nr); err != nil {
			return n, err
		}
	}

	return
}

// Write writes data to the connection and will wait if total n byte written exceed the limit.
func (lc *Conn) Write(p []byte) (n int, err error) {
	if lc.WriteLimiter == nil {
		return lc.Conn.Write(p)
	}

	limit := int64(lc.WriteLimiter.Limit())
	np := int64(len(p))
	if np < limit {
		limit = np
	}

	for i := int64(0); i < np; i += limit {
		end := i + limit
		if end > np {
			end = np
		}

		nw, err := lc.Conn.Write(p[i:end])
		n += nw
		if err != nil {
			return n, err
		}

		if err := lc.WriteLimiter.WaitN(context.Background(), nw); err != nil {
			return n, err
		}
	}

	return
}
