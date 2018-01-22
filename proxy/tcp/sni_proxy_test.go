package tcp

import (
	"testing"

	"github.com/fabiolb/fabio/assert"
)

func TestCreateClientHelloBufferNotTLS(t *testing.T) {
	assertEqual := assert.Equal(t)

	testCases := [][]byte{
		// not enough data
		{0x16, 0x03, 0x01, 0x00, 0x00, 0x01, 0x00, 0x05},

		// not tls record
		{0x15, 0x03, 0x01, 0x01, 0xF4, 0x01, 0x00, 0x01, 0xeb},

		// too large record
		//                |---------|
		{0x16, 0x03, 0x01, 0x40, 0x01, 0x01, 0x00, 0x01, 0xec},

		// zero record length
		//                |----------|
		{0x16, 0x03, 0x01, 0x00, 0x00, 0x01, 0x00, 0x01, 0xec},

		// not client hello
		//                            |----|
		{0x16, 0x03, 0x01, 0x01, 0xF4, 0x02, 0x00, 0x01, 0xeb},

		// bad handshake length
		//                                  |----- 0 --------|
		{0x16, 0x03, 0x01, 0x00, 0xaa, 0x01, 0x00, 0x00, 0x00},

		// Fragmentation (handshake larger than record)
		//                |-  500 ---|      |----- 497 ------|
		{0x16, 0x03, 0x01, 0x01, 0xF4, 0x01, 0x00, 0x01, 0xf1},
	}

	for i := 0; i < len(testCases); i++ {
		_, err := createClientHelloBuffer(testCases[i])
		if err == nil {
			t.Logf("Case idx %d did not return an error", i)
		}
		assertEqual(err != nil, true)
	}
}

func TestCreateClientHelloBufferOk(t *testing.T) {
	assertEqual := assert.Equal(t)
	// Largest possible client hello message
	//                               |- 16384 -|      |----- 16380 ----|
	data := []byte{0x16, 0x03, 0x01, 0x40, 0x00, 0x01, 0x00, 0x3f, 0xfc}
	buffer, err := createClientHelloBuffer(data)
	assertEqual(err, nil)
	assertEqual(buffer != nil, true)
	assertEqual(len(buffer), 16384+5) // record length + record header
}
