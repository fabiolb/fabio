// (C) Copyright 2012, Jeramey Crawford <jeramey@antihe.ro>. All
// rights reserved. Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package common

const (
	alphabet = "./0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
)

// Base64_24Bit is a variant of Base64 encoding, commonly used with password
// hashing algorithms to encode the result of their checksum output.
//
// The algorithm operates on up to 3 bytes at a time, encoding the following
// 6-bit sequences into up to 4 hash64 ASCII bytes.
//
//   1. Bottom 6 bits of the first byte
//   2. Top 2 bits of the first byte, and bottom 4 bits of the second byte.
//   3. Top 4 bits of the second byte, and bottom 2 bits of the third byte.
//   4. Top 6 bits of the third byte.
//
// This encoding method does not emit padding bytes as Base64 does.
func Base64_24Bit(src []byte) []byte {
	if len(src) == 0 {
		return []byte{} // TODO: return nil
	}

	dstlen := (len(src)*8 + 5) / 6
	dst := make([]byte, dstlen)

	di, si := 0, 0
	n := len(src) / 3 * 3
	for si < n {
		val := uint(src[si+2])<<16 | uint(src[si+1])<<8 | uint(src[si])
		dst[di+0] = alphabet[val&0x3f]
		dst[di+1] = alphabet[val>>6&0x3f]
		dst[di+2] = alphabet[val>>12&0x3f]
		dst[di+3] = alphabet[val>>18]
		di += 4
		si += 3
	}

	rem := len(src) - si
	if rem == 0 {
		return dst
	}

	val := uint(src[si+0])
	if rem == 2 {
		val |= uint(src[si+1]) << 8
	}

	dst[di+0] = alphabet[val&0x3f]
	dst[di+1] = alphabet[val>>6&0x3f]
	if rem == 2 {
		dst[di+2] = alphabet[val>>12]
	}
	return dst
}
