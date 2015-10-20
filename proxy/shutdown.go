package proxy

import "sync/atomic"

var shutdown int32

// Shutdown sets the shutdown flag which triggers the proxy
// to stop routing new requests.
func Shutdown() {
	atomic.StoreInt32(&shutdown, 1)
}

// ShuttingDown returns whether the shutdown flag has been set.
func ShuttingDown() bool {
	return atomic.LoadInt32(&shutdown) != 0
}
