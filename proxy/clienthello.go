package proxy

// record types
const (
	handshakeRecord = 0x16
	clientHelloType = 0x01
)

// readServerName returns the server name from a TLS ClientHello message which
// has the server_name extension (SNI). ok is set to true if the ClientHello
// message was parsed successfully. If the server_name extension was not set
// and empty string is returned as serverName.
func readServerName(data []byte) (serverName string, ok bool) {
	if m, ok := readClientHello(data); ok {
		return m.serverName, true
	}
	return "", false
}

// readClientHello
func readClientHello(data []byte) (m *clientHelloMsg, ok bool) {
	if len(data) < 9 {
		// println("buf too short")
		return nil, false
	}

	// TLS record header
	// -----------------
	// byte   0: rec type (should be 0x16 == Handshake)
	// byte 1-2: version (should be 0x3000 < v < 0x3003)
	// byte 3-4: rec len
	recType := data[0]
	if recType != handshakeRecord {
		// println("no handshake ")
		return nil, false
	}

	recLen := int(data[3])<<8 | int(data[4])
	if recLen == 0 || recLen > len(data)-5 {
		// println("rec too short")
		return nil, false
	}

	// Handshake record header
	// -----------------------
	// byte   5: hs msg type (should be 0x01 == client_hello)
	// byte 6-8: hs msg len
	hsType := data[5]
	if hsType != clientHelloType {
		// println("no client_hello")
		return nil, false
	}

	hsLen := int(data[6])<<16 | int(data[7])<<8 | int(data[8])
	if hsLen == 0 || hsLen > len(data)-9 {
		// println("handshake rec too short")
		return nil, false
	}

	// byte 9- : client hello msg
	//
	// m.unmarshal parses the entire handshake message and
	// not just the client hello. Therefore, we need to pass
	// data from byte 5 instead of byte 9. (see comment below)
	m = new(clientHelloMsg)
	if !m.unmarshal(data[5:]) {
		// println("client_hello unmarshal failed")
		return nil, false
	}
	return m, true
}

// The code below is a verbatim copy from go1.7/src/crypto/tls/handshake_messages.go
// with some parts commented out. It does enough work to parse a TLS client hello
// message and extract the server name extension since this is all we care about.
//
// Copyright (c) 2016 The Go Authors

// TLS extension numbers
const (
	extensionServerName uint16 = 0
	// extensionStatusRequest       uint16 = 5
	// extensionSupportedCurves     uint16 = 10
	// extensionSupportedPoints     uint16 = 11
	// extensionSignatureAlgorithms uint16 = 13
	// extensionALPN                uint16 = 16
	// extensionSCT                 uint16 = 18 // https://tools.ietf.org/html/rfc6962#section-6
	// extensionSessionTicket       uint16 = 35
	// extensionNextProtoNeg        uint16 = 13172 // not IANA assigned
	// extensionRenegotiationInfo   uint16 = 0xff01
)

type clientHelloMsg struct {
	raw                []byte
	vers               uint16
	random             []byte
	sessionId          []byte
	cipherSuites       []uint16
	compressionMethods []uint8
	nextProtoNeg       bool
	serverName         string
	ocspStapling       bool
	scts               bool
	// supportedCurves              []CurveID
	supportedPoints []uint8
	ticketSupported bool
	sessionTicket   []uint8
	//signatureAndHashes           []signatureAndHash
	secureRenegotiation          []byte
	secureRenegotiationSupported bool
	alpnProtocols                []string
}

func (m *clientHelloMsg) unmarshal(data []byte) bool {
	if len(data) < 42 {
		return false
	}
	m.raw = data
	m.vers = uint16(data[4])<<8 | uint16(data[5])
	m.random = data[6:38]
	sessionIdLen := int(data[38])
	if sessionIdLen > 32 || len(data) < 39+sessionIdLen {
		return false
	}
	m.sessionId = data[39 : 39+sessionIdLen]
	data = data[39+sessionIdLen:]
	if len(data) < 2 {
		return false
	}
	// cipherSuiteLen is the number of bytes of cipher suite numbers. Since
	// they are uint16s, the number must be even.
	cipherSuiteLen := int(data[0])<<8 | int(data[1])
	if cipherSuiteLen%2 == 1 || len(data) < 2+cipherSuiteLen {
		return false
	}
	// numCipherSuites := cipherSuiteLen / 2
	// m.cipherSuites = make([]uint16, numCipherSuites)
	// for i := 0; i < numCipherSuites; i++ {
	// 	m.cipherSuites[i] = uint16(data[2+2*i])<<8 | uint16(data[3+2*i])
	// 	if m.cipherSuites[i] == scsvRenegotiation {
	// 		m.secureRenegotiationSupported = true
	// 	}
	// }
	data = data[2+cipherSuiteLen:]
	if len(data) < 1 {
		return false
	}
	compressionMethodsLen := int(data[0])
	if len(data) < 1+compressionMethodsLen {
		return false
	}
	m.compressionMethods = data[1 : 1+compressionMethodsLen]

	data = data[1+compressionMethodsLen:]

	m.nextProtoNeg = false
	m.serverName = ""
	m.ocspStapling = false
	m.ticketSupported = false
	m.sessionTicket = nil
	// m.signatureAndHashes = nil
	m.alpnProtocols = nil
	m.scts = false

	if len(data) == 0 {
		// ClientHello is optionally followed by extension data
		return true
	}
	if len(data) < 2 {
		return false
	}

	extensionsLength := int(data[0])<<8 | int(data[1])
	data = data[2:]
	if extensionsLength != len(data) {
		return false
	}

	for len(data) != 0 {
		if len(data) < 4 {
			return false
		}
		extension := uint16(data[0])<<8 | uint16(data[1])
		length := int(data[2])<<8 | int(data[3])
		data = data[4:]
		if len(data) < length {
			return false
		}

		switch extension {
		case extensionServerName:
			d := data[:length]
			if len(d) < 2 {
				return false
			}
			namesLen := int(d[0])<<8 | int(d[1])
			d = d[2:]
			if len(d) != namesLen {
				return false
			}
			for len(d) > 0 {
				if len(d) < 3 {
					return false
				}
				nameType := d[0]
				nameLen := int(d[1])<<8 | int(d[2])
				d = d[3:]
				if len(d) < nameLen {
					return false
				}
				if nameType == 0 {
					m.serverName = string(d[:nameLen])
					break
				}
				d = d[nameLen:]
			}
			// case extensionNextProtoNeg:
			// 	if length > 0 {
			// 		return false
			// 	}
			// 	m.nextProtoNeg = true
			// case extensionStatusRequest:
			// 	m.ocspStapling = length > 0 && data[0] == statusTypeOCSP
			// case extensionSupportedCurves:
			// 	// http://tools.ietf.org/html/rfc4492#section-5.5.1
			// 	if length < 2 {
			// 		return false
			// 	}
			// 	l := int(data[0])<<8 | int(data[1])
			// 	if l%2 == 1 || length != l+2 {
			// 		return false
			// 	}
			// 	numCurves := l / 2
			// 	m.supportedCurves = make([]CurveID, numCurves)
			// 	d := data[2:]
			// 	for i := 0; i < numCurves; i++ {
			// 		m.supportedCurves[i] = CurveID(d[0])<<8 | CurveID(d[1])
			// 		d = d[2:]
			// 	}
			// case extensionSupportedPoints:
			// 	// http://tools.ietf.org/html/rfc4492#section-5.5.2
			// 	if length < 1 {
			// 		return false
			// 	}
			// 	l := int(data[0])
			// 	if length != l+1 {
			// 		return false
			// 	}
			// 	m.supportedPoints = make([]uint8, l)
			// 	copy(m.supportedPoints, data[1:])
			// case extensionSessionTicket:
			// 	// http://tools.ietf.org/html/rfc5077#section-3.2
			// 	m.ticketSupported = true
			// 	m.sessionTicket = data[:length]
			// case extensionSignatureAlgorithms:
			// 	// https://tools.ietf.org/html/rfc5246#section-7.4.1.4.1
			// 	if length < 2 || length&1 != 0 {
			// 		return false
			// 	}
			// 	l := int(data[0])<<8 | int(data[1])
			// 	if l != length-2 {
			// 		return false
			// 	}
			// 	n := l / 2
			// 	d := data[2:]
			// 	m.signatureAndHashes = make([]signatureAndHash, n)
			// 	for i := range m.signatureAndHashes {
			// 		m.signatureAndHashes[i].hash = d[0]
			// 		m.signatureAndHashes[i].signature = d[1]
			// 		d = d[2:]
			// 	}
			// case extensionRenegotiationInfo:
			// 	if length == 0 {
			// 		return false
			// 	}
			// 	d := data[:length]
			// 	l := int(d[0])
			// 	d = d[1:]
			// 	if l != len(d) {
			// 		return false
			// 	}

			// 	m.secureRenegotiation = d
			// 	m.secureRenegotiationSupported = true
			// case extensionALPN:
			// 	if length < 2 {
			// 		return false
			// 	}
			// 	l := int(data[0])<<8 | int(data[1])
			// 	if l != length-2 {
			// 		return false
			// 	}
			// 	d := data[2:length]
			// 	for len(d) != 0 {
			// 		stringLen := int(d[0])
			// 		d = d[1:]
			// 		if stringLen == 0 || stringLen > len(d) {
			// 			return false
			// 		}
			// 		m.alpnProtocols = append(m.alpnProtocols, string(d[:stringLen]))
			// 		d = d[stringLen:]
			// 	}
			// case extensionSCT:
			// 	m.scts = true
			// 	if length != 0 {
			// 		return false
			// 	}
		}
		data = data[length:]
	}

	return true
}
