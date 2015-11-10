package proxy

import "sync/atomic"

var shutdown int32

func Shutdown() {
	atomic.StoreInt32(&shutdown, 1)
}

func ShuttingDown() bool {
	return atomic.LoadInt32(&shutdown) != 0
}
