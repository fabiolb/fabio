// (C) Copyright 2012, Jeramey Crawford <jeramey@antihe.ro>. All
// rights reserved. Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package sha512_crypt implements Ulrich Drepper's SHA512-crypt password
// hashing algorithm.
//
// The specification for this algorithm can be found here:
// http://www.akkadia.org/drepper/SHA-crypt.txt
package sha512_crypt

import (
	"bytes"
	"crypto/sha512"
	"crypto/subtle"
	"strconv"

	"github.com/GehirnInc/crypt"
	"github.com/GehirnInc/crypt/common"
	"github.com/GehirnInc/crypt/internal"
)

func init() {
	crypt.RegisterCrypt(crypt.SHA512, New, MagicPrefix)
}

const (
	MagicPrefix   = "$6$"
	SaltLenMin    = 1
	SaltLenMax    = 16
	RoundsMin     = 1000
	RoundsMax     = 999999999
	RoundsDefault = 5000
)

var _rounds = []byte("rounds=")

type crypter struct{ Salt common.Salt }

// New returns a new crypt.Crypter computing the SHA512-crypt password hashing.
func New() crypt.Crypter {
	return &crypter{
		common.Salt{
			MagicPrefix:   []byte(MagicPrefix),
			SaltLenMin:    SaltLenMin,
			SaltLenMax:    SaltLenMax,
			RoundsDefault: RoundsDefault,
			RoundsMin:     RoundsMin,
			RoundsMax:     RoundsMax,
		},
	}
}

func (c *crypter) Generate(key, salt []byte) (string, error) {
	if len(salt) == 0 {
		salt = c.Salt.GenerateWRounds(SaltLenMax, RoundsDefault)
	}
	salt, rounds, isRoundsDef, _, err := c.Salt.Decode(salt)
	if err != nil {
		return "", err
	}

	keyLen := len(key)
	saltLen := len(salt)
	h := sha512.New()

	// compute sumB
	// step 4-8
	h.Write(key)
	h.Write(salt)
	h.Write(key)
	sumB := h.Sum(nil)

	// Compute sumA
	// step 1-3, 9-12
	h.Reset()
	h.Write(key)
	h.Write(salt)
	h.Write(internal.RepeatByteSequence(sumB, keyLen))
	for i := keyLen; i > 0; i >>= 1 {
		if i%2 == 0 {
			h.Write(key)
		} else {
			h.Write(sumB)
		}
	}
	sumA := h.Sum(nil)
	internal.CleanSensitiveData(sumB)

	// Compute seqP
	// step 13-16
	h.Reset()
	for i := 0; i < keyLen; i++ {
		h.Write(key)
	}
	seqP := internal.RepeatByteSequence(h.Sum(nil), keyLen)

	// Compute seqS
	// step 17-20
	h.Reset()
	for i := 0; i < 16+int(sumA[0]); i++ {
		h.Write(salt)
	}
	seqS := internal.RepeatByteSequence(h.Sum(nil), saltLen)

	// step 21
	for i := 0; i < rounds; i++ {
		h.Reset()

		if i&1 != 0 {
			h.Write(seqP)
		} else {
			h.Write(sumA)
		}
		if i%3 != 0 {
			h.Write(seqS)
		}
		if i%7 != 0 {
			h.Write(seqP)
		}
		if i&1 != 0 {
			h.Write(sumA)
		} else {
			h.Write(seqP)
		}
		copy(sumA, h.Sum(nil))
	}
	internal.CleanSensitiveData(seqP)
	internal.CleanSensitiveData(seqS)

	// make output
	buf := bytes.Buffer{}
	buf.Grow(len(c.Salt.MagicPrefix) + len(_rounds) + 9 + 1 + len(salt) + 1 + 86)
	buf.Write(c.Salt.MagicPrefix)
	if isRoundsDef {
		buf.Write(_rounds)
		buf.WriteString(strconv.Itoa(rounds))
		buf.WriteByte('$')
	}
	buf.Write(salt)
	buf.WriteByte('$')
	buf.Write(common.Base64_24Bit([]byte{
		sumA[42], sumA[21], sumA[0],
		sumA[1], sumA[43], sumA[22],
		sumA[23], sumA[2], sumA[44],
		sumA[45], sumA[24], sumA[3],
		sumA[4], sumA[46], sumA[25],
		sumA[26], sumA[5], sumA[47],
		sumA[48], sumA[27], sumA[6],
		sumA[7], sumA[49], sumA[28],
		sumA[29], sumA[8], sumA[50],
		sumA[51], sumA[30], sumA[9],
		sumA[10], sumA[52], sumA[31],
		sumA[32], sumA[11], sumA[53],
		sumA[54], sumA[33], sumA[12],
		sumA[13], sumA[55], sumA[34],
		sumA[35], sumA[14], sumA[56],
		sumA[57], sumA[36], sumA[15],
		sumA[16], sumA[58], sumA[37],
		sumA[38], sumA[17], sumA[59],
		sumA[60], sumA[39], sumA[18],
		sumA[19], sumA[61], sumA[40],
		sumA[41], sumA[20], sumA[62],
		sumA[63],
	}))
	return buf.String(), nil
}

func (c *crypter) Verify(hashedKey string, key []byte) error {
	newHash, err := c.Generate(key, []byte(hashedKey))
	if err != nil {
		return err
	}
	if subtle.ConstantTimeCompare([]byte(newHash), []byte(hashedKey)) != 1 {
		return crypt.ErrKeyMismatch
	}
	return nil
}

func (c *crypter) Cost(hashedKey string) (int, error) {
	_, rounds, _, _, err := c.Salt.Decode([]byte(hashedKey))
	if err != nil {
		return 0, err
	}
	return rounds, nil
}

func (c *crypter) SetSalt(salt common.Salt) { c.Salt = salt }
