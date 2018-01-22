package tcp

import (
	"encoding/hex"
	"testing"

	"github.com/fabiolb/fabio/assert"
)

func TestReadServerNameBadData(t *testing.T) {
	assertEqual := assert.Equal(t)
	clientHelloMsg := []byte{0x16, 0x03, 0x01, 0x45, 0x03, 0x01, 0x2, 0x01}
	serverName, ok := readServerName(clientHelloMsg)
	assertEqual(serverName, "")
	assertEqual(ok, false)
}

func TestReadServerNameNoExtension(t *testing.T) {
	assertEqual := assert.Equal(t)
	// Client hello from:
	// openssl s_client -connect google.com:443
	clientHelloMsg, _ := hex.DecodeString(
		"0100013503036dfb09de7b16503dd1bb304dcbe54079913b65abf53de997f73b26c99e" +
			"67ba28000098cc14cc13cc15c030c02cc028c024c014c00a00a3009f006b006a00" +
			"390038ff8500c400c3008800870081c032c02ec02ac026c00fc005009d003d0035" +
			"00c00084c02fc02bc027c023c013c00900a2009e006700400033003200be00bd00" +
			"450044c031c02dc029c025c00ec004009c003c002f00ba0041c011c007c00cc002" +
			"00050004c012c00800160013c00dc003000a00150012000900ff01000074000b00" +
			"0403000102000a003a0038000e000d0019001c000b000c001b00180009000a001a" +
			"00160017000800060007001400150004000500120013000100020003000f001000" +
			"1100230000000d00260024060106020603efef050105020503040104020403eeee" +
			"eded030103020303020102020203")
	serverName, ok := readServerName(clientHelloMsg)
	assertEqual(serverName, "")
	assertEqual(ok, true)
}

func TestReadServerNameOk(t *testing.T) {
	assertEqual := assert.Equal(t)
	// Client hello from:
	// openssl s_client -connect google.com:443 -servername google.com
	clientHelloMsg, _ := hex.DecodeString(
		"0100014803032657cacce41598fa82e5b75061050bc31c5affdba106b8e7431852" +
			"24af0fa1aa000098cc14cc13cc15c030c02cc028c024c014c00a00a3009f00" +
			"6b006a00390038ff8500c400c3008800870081c032c02ec02ac026c00fc005" +
			"009d003d003500c00084c02fc02bc027c023c013c00900a2009e0067004000" +
			"33003200be00bd00450044c031c02dc029c025c00ec004009c003c002f00ba" +
			"0041c011c007c00cc00200050004c012c00800160013c00dc003000a001500" +
			"12000900ff010000870000000f000d00000a676f6f676c652e636f6d000b00" +
			"0403000102000a003a0038000e000d0019001c000b000c001b00180009000a" +
			"001a0016001700080006000700140015000400050012001300010002000300" +
			"0f0010001100230000000d00260024060106020603efef0501050205030401" +
			"04020403eeeeeded030103020303020102020203")
	serverName, ok := readServerName(clientHelloMsg)
	assertEqual(serverName, "google.com")
	assertEqual(ok, true)
}
