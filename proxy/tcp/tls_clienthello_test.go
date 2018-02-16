package tcp

import (
	"encoding/hex"
	"testing"
)

func TestClientHelloBufferSize(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		size int
		fail bool
	}{
		{
			name: "valid data",
			// Largest possible client hello message
			//                            |- 16384 -|      |----- 16380 ----|
			data: []byte{0x16, 0x03, 0x01, 0x40, 0x00, 0x01, 0x00, 0x3f, 0xfc},
			size: 16384 + 5, // max record length + record header
			fail: false,
		},
		{
			name: "not enough data",
			data: []byte{0x16, 0x03, 0x01, 0x40, 0x00, 0x01, 0x00, 0x3f},
			size: 0,
			fail: true,
		},
		{
			name: "not a TLS record",
			data: []byte{0x15, 0x03, 0x01, 0x01, 0xF4, 0x01, 0x00, 0x01, 0xeb},
			size: 0,
			fail: true,
		},

		{
			name: "TLS record too large",
			//                             | max + 1 |
			data: []byte{0x16, 0x03, 0x01, 0x40, 0x01, 0x01, 0x00, 0x3f, 0xfc},
			size: 0,
			fail: true,
		},

		{
			name: "TLS record length zero",
			//                            |----------|
			data: []byte{0x16, 0x03, 0x01, 0x00, 0x00, 0x01, 0x00, 0x3f, 0xfc},
			size: 0,
			fail: true,
		},

		{
			name: "Not a client hello",
			//                                        |----|
			data: []byte{0x16, 0x03, 0x01, 0x40, 0x00, 0x02, 0x00, 0x3f, 0xfc},
			size: 0,
			fail: true,
		},

		{
			name: "Invalid handshake message record length",
			//                                              |----- 0 --------|
			data: []byte{0x16, 0x03, 0x01, 0x40, 0x00, 0x01, 0x00, 0x00, 0x00},
			size: 0,
			fail: true,
		},

		{
			name: "Fragmentation (handshake message larger than record)",
			//                            |-  500 ---|      |----- 497 ------|
			data: []byte{0x16, 0x03, 0x01, 0x01, 0xF4, 0x01, 0x00, 0x01, 0xf1},
			size: 0,
			fail: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := clientHelloBufferSize(tt.data)

			if tt.fail && err == nil {
				t.Fatal("expected error, got nil")
			} else if !tt.fail && err != nil {
				t.Fatalf("expected error to be nil, got %s", err)
			}

			if want := tt.size; got != want {
				t.Fatalf("want size %d, got %d", want, got)
			}
		})
	}
}

func TestReadServerName(t *testing.T) {
	tests := []struct {
		name       string
		servername string
		ok         bool
		data       string //Hex string, decoded by test
	}{
		{
			// Client hello from:
			// openssl s_client -connect google.com:443 -servername google.com
			name:       "valid client hello with server name",
			servername: "google.com",
			ok:         true,
			data: "0100014803032657cacce41598fa82e5b75061050bc31c5affdba106b8e7431852" +
				"24af0fa1aa000098cc14cc13cc15c030c02cc028c024c014c00a00a3009f00" +
				"6b006a00390038ff8500c400c3008800870081c032c02ec02ac026c00fc005" +
				"009d003d003500c00084c02fc02bc027c023c013c00900a2009e0067004000" +
				"33003200be00bd00450044c031c02dc029c025c00ec004009c003c002f00ba" +
				"0041c011c007c00cc00200050004c012c00800160013c00dc003000a001500" +
				"12000900ff010000870000000f000d00000a676f6f676c652e636f6d000b00" +
				"0403000102000a003a0038000e000d0019001c000b000c001b00180009000a" +
				"001a0016001700080006000700140015000400050012001300010002000300" +
				"0f0010001100230000000d00260024060106020603efef0501050205030401" +
				"04020403eeeeeded030103020303020102020203",
		},
		{
			// Client hello from:
			// openssl s_client -connect google.com:443
			name:       "valid client hello but no server name extension",
			servername: "",
			ok:         true,
			data: "0100013503036dfb09de7b16503dd1bb304dcbe54079913b65abf53de997f73b26c99e" +
				"67ba28000098cc14cc13cc15c030c02cc028c024c014c00a00a3009f006b006a00" +
				"390038ff8500c400c3008800870081c032c02ec02ac026c00fc005009d003d0035" +
				"00c00084c02fc02bc027c023c013c00900a2009e006700400033003200be00bd00" +
				"450044c031c02dc029c025c00ec004009c003c002f00ba0041c011c007c00cc002" +
				"00050004c012c00800160013c00dc003000a00150012000900ff01000074000b00" +
				"0403000102000a003a0038000e000d0019001c000b000c001b00180009000a001a" +
				"00160017000800060007001400150004000500120013000100020003000f001000" +
				"1100230000000d00260024060106020603efef050105020503040104020403eeee" +
				"eded030103020303020102020203",
		},
		{
			name:       "invalid client hello",
			servername: "",
			ok:         false,
			data: "0100014c5768656e2070656f706c652073617920746f206d653a20776f756c6420796f" +
				"75207261746865722062652074686f75676874206f6620617320612066756e6e79" +
				"206d616e206f72206120677265617420626f73733f204d7920616e737765722773" +
				"20616c77617973207468652073616d652c20746f206d652c207468657927726520" +
				"6e6f74206d757475616c6c79206578636c75736976652e2d204461766964204272" +
				"656e74",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clientHelloMsg, _ := hex.DecodeString(tt.data)
			servername, ok := readServerName(clientHelloMsg)
			if got, want := servername, tt.servername; got != want {
				t.Fatalf("%s: got servername \"%s\" want \"%s\"", tt.name, got, want)
			}

			if got, want := ok, tt.ok; got != want {
				t.Fatalf("%s: got ok %t want %t", tt.name, got, want)
			}
		})
	}
}
