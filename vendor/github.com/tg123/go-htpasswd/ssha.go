package htpasswd

import (
	"crypto/sha1"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"
)

type sshaPassword struct {
	hashed []byte
	salt   []byte
}

//AcceptSsha accepts any valid password encoded using bcrypt.
func AcceptSsha(src string) (EncodedPasswd, error) {
	if !strings.HasPrefix(src, "{SSHA}") {
		return nil, nil
	}

	b64 := strings.TrimPrefix(src, "{SSHA}")
	hashed, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return nil, fmt.Errorf("Malformed ssha(%s): %s", src, err.Error())
	}

	//ssha appends the length onto the end of the SHA, so the length can't be less than sha1.Size.
	if len(hashed) < sha1.Size {
		return nil, fmt.Errorf("Malformed ssha(%s): wrong length", src)
	}

	hash := hashed[:sha1.Size]
	salt := hashed[sha1.Size:]
	return &sshaPassword{hash, salt}, nil
}

//RejectSsha rejects any password encoded using SSHA1.
func RejectSsha(src string) (EncodedPasswd, error) {
	if !strings.HasPrefix(src, "{SSHA}") {
		return nil, nil
	}
	return nil, fmt.Errorf("ssha passwords are not accepted: %s", src)
}

func (s *sshaPassword) MatchesPassword(password string) bool {
	//SSHA appends the salt onto the password before computing the hash.
	sha := append([]byte(password), s.salt[:]...)
	hash := sha1.Sum(sha)
	return subtle.ConstantTimeCompare(hash[:], s.hashed) == 1
}
