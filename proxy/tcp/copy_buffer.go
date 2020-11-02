package tcp

import (
	"io"

	gkm "github.com/go-kit/kit/metrics"
)

// copyBuffer is an adapted version of io.copyBuffer which updates a
// counter instead of returning the total bytes written.
func copyBuffer(dst io.Writer, src io.Reader, c gkm.Counter) (err error) {
	buf := make([]byte, 32*1024)
	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			nw, ew := dst.Write(buf[0:nr])
			if nw > 0 {
				if c != nil {
					c.Add(float64(nw))
				}
			}
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
		}
		if er != nil {
			if er != io.EOF {
				err = er
			}
			break
		}
	}
	return err
}
