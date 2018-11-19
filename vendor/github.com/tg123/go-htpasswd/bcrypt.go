package htpasswd

import (
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type bcryptPassword struct {
	hashed []byte
}

//AcceptBcrypt accepts any valid password encoded using bcrypt.
func AcceptBcrypt(src string) (EncodedPasswd, error) {
	if !strings.HasPrefix(src, "$2y$") && !strings.HasPrefix(src, "$2a$") {
		return nil, nil
	}

	return &bcryptPassword{hashed: []byte(src)}, nil
}

//RejectBcrypt rejects any password encoded using bcrypt.
func RejectBcrypt(src string) (EncodedPasswd, error) {
	if strings.HasPrefix(src, "$2y$") || strings.HasPrefix(src, "$2a$") {
		return nil, fmt.Errorf("bcrypt passwords are not accepted: %s", src)
	}

	return nil, nil
}

func (b *bcryptPassword) MatchesPassword(password string) bool {
	if err := bcrypt.CompareHashAndPassword(b.hashed, []byte(password)); err != nil {
		return false
	}
	return true
}
