// (C) Copyright 2013, Jonas mg. All rights reserved.
// Use of this source code is governed by a BSD-style license
// that can be found in the LICENSE file.

// Package crypt provides interface for password crypt functions and collects
// common constants.
package crypt

import (
	"errors"
	"strings"

	"github.com/GehirnInc/crypt/common"
)

var ErrKeyMismatch = errors.New("hashed value is not the hash of the given password")

// Crypter is the common interface implemented by all crypt functions.
type Crypter interface {
	// Generate performs the hashing algorithm, returning a full hash suitable
	// for storage and later password verification.
	//
	// If the salt is empty, a randomly-generated salt will be generated with a
	// length of SaltLenMax and number RoundsDefault of rounds.
	//
	// Any error only can be got when the salt argument is not empty.
	Generate(key, salt []byte) (string, error)

	// Verify compares a hashed key with its possible key equivalent.
	// Returns nil on success, or an error on failure; if the hashed key is
	// diffrent, the error is "ErrKeyMismatch".
	Verify(hashedKey string, key []byte) error

	// Cost returns the hashing cost (in rounds) used to create the given hashed
	// key.
	//
	// When, in the future, the hashing cost of a key needs to be increased in
	// order to adjust for greater computational power, this function allows one
	// to establish which keys need to be updated.
	//
	// The algorithms based in MD5-crypt use a fixed value of rounds.
	Cost(hashedKey string) (int, error)

	// SetSalt sets a different salt. It is used to easily create derivated
	// algorithms, i.e. "apr1_crypt" from "md5_crypt".
	SetSalt(salt common.Salt)
}

// Crypt identifies a crypt function that is implemented in another package.
type Crypt uint

const (
	APR1   Crypt = 1 + iota // import github.com/GehirnInc/crypt/apr1_crypt
	MD5                     // import github.com/GehirnInc/crypt/md5_crypt
	SHA256                  // import github.com/GehirnInc/crypt/sha256_crypt
	SHA512                  // import github.com/GehirnInc/crypt/sha512_crypt
	maxCrypt
)

var crypts = make([]func() Crypter, maxCrypt)

// New returns new Crypter making the Crypt c.
// New panics if the Crypt c is unavailable.
func (c Crypt) New() Crypter {
	if c > 0 && c < maxCrypt {
		f := crypts[c]
		if f != nil {
			return f()
		}
	}
	panic("crypt: requested crypt function is unavailable")
}

// Available reports whether the Crypt c is available.
func (c Crypt) Available() bool {
	return c > 0 && c < maxCrypt && crypts[c] != nil
}

var cryptPrefixes = make([]string, maxCrypt)

// RegisterCrypt registers a function that returns a new instance of the given
// crypt function. This is intended to be called from the init function in
// packages that implement crypt functions.
func RegisterCrypt(c Crypt, f func() Crypter, prefix string) {
	if c >= maxCrypt {
		panic("crypt: RegisterHash of unknown crypt function")
	}
	crypts[c] = f
	cryptPrefixes[c] = prefix
}

// New returns a new crypter.
func New(c Crypt) Crypter {
	return c.New()
}

// IsHashSupported returns true if hashedKey has a supported prefix.
// NewFromHash will not panic for this hashedKey
func IsHashSupported(hashedKey string) bool {
	for i := range cryptPrefixes {
		prefix := cryptPrefixes[i]
		if crypts[i] != nil && strings.HasPrefix(hashedKey, prefix) {
			return true
		}
	}

	return false
}

// NewFromHash returns a new Crypter using the prefix in the given hashed key.
func NewFromHash(hashedKey string) Crypter {
	for i := range cryptPrefixes {
		prefix := cryptPrefixes[i]
		if crypts[i] != nil && strings.HasPrefix(hashedKey, prefix) {
			crypt := Crypt(uint(i))
			return crypt.New()
		}
	}

	panic("crypt: unknown crypt function")
}
