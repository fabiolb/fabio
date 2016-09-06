package proxy

import (
	"net"
	"time"
)

// Dialer can dial ip connections from different source ip
// addresses to overcome the limitation on source ports.
type Dialer struct {
	LocalAddrs []net.Addr
	Timeout    time.Duration
	KeepAlive  time.Duration
}

func (d *Dialer) Dial(network, address string) (c net.Conn, err error) {
	// create a new dialer
	nd := &net.Dialer{KeepAlive: d.KeepAlive}

	// use a deadline to ensure the timeout is
	// for setting up the connection and does
	// not become cumulativ per attempt
	if d.Timeout > 0 {
		nd.Deadline = time.Now().Add(d.Timeout)
	}

	n := len(d.LocalAddrs)

	// use default source address
	if n == 0 {
		return nd.Dial(network, address)
	}

	// only one address, no work to do
	if n == 1 {
		nd.LocalAddr = d.LocalAddrs[0]
		return nd.Dial(network, address)
	}

	// start with a random address and continue
	// until we establish a connection.
	// TODO(fs): can we tell "no more ports" from other errors? Does it matter?
	start := int(time.Now().UnixNano()) % n
	for i := 0; i < n; i++ {
		nd.LocalAddr = d.LocalAddrs[(start+i)%n]
		if c, err = nd.Dial(network, address); err == nil {
			return
		}
	}
	return
}
