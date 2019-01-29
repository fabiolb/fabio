package tcp

import (
	"fmt"
	"net"
)

// WriteProxyHeader extracts remote and local IP address and port
// combinations from incoming connection and writes the PROXY proto
// header to the outgoing connection
func WriteProxyHeader(out, in net.Conn) error {
	clientAddr, clientPort, _ := net.SplitHostPort(in.RemoteAddr().String())
	serverAddr, serverPort, _ := net.SplitHostPort(in.LocalAddr().String())

	var proto string
	if net.ParseIP(clientAddr).To4() != nil {
		proto = "TCP4"
	} else {
		proto = "TCP6"
	}

	header := fmt.Sprintf("PROXY %s %s %s %s %s\r\n", proto, clientAddr, serverAddr, clientPort, serverPort)
	_, err := out.Write([]byte(header))
	return err
}
