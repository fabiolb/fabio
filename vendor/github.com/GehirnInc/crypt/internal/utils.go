// Copyright (c) 2015 Kohei YOSHIDA. All rights reserved.
// This software is licensed under the 3-Clause BSD License
// that can be found in LICENSE file.
package internal

const (
	cleanBytesLen = 64
)

var (
	cleanBytes = make([]byte, cleanBytesLen)
)

func CleanSensitiveData(b []byte) {
	l := len(b)

	for ; l > cleanBytesLen; l -= cleanBytesLen {
		copy(b[l-cleanBytesLen:l], cleanBytes)
	}

	if l > 0 {
		copy(b[0:l], cleanBytes[0:l])
	}
}

func RepeatByteSequence(input []byte, length int) []byte {
	var (
		sequence = make([]byte, length)
		unit     = len(input)
	)

	j := length / unit * unit
	for i := 0; i < j; i += unit {
		copy(sequence[i:length], input)
	}
	if j < length {
		copy(sequence[j:length], input[0:length-j])
	}

	return sequence
}
