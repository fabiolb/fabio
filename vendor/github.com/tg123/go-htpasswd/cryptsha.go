package htpasswd

import (
	"fmt"
	"strings"

	"github.com/GehirnInc/crypt"
	_ "github.com/GehirnInc/crypt/sha256_crypt"
	_ "github.com/GehirnInc/crypt/sha512_crypt"
)

type cryptPassword struct {
	prefix string
	rounds string
	salt   string
	hashed string
}

// Prefixes
const PrefixCryptSha256 = "$5$"
const PrefixCryptSha512 = "$6$"
const Separator = "$"

// Accepts valid passwords
func AcceptCryptSha(src string) (EncodedPasswd, error) {
	if !strings.HasPrefix(src, PrefixCryptSha256) && !strings.HasPrefix(src, PrefixCryptSha512) {
		return nil, nil
	}

	prefix := PrefixCryptSha512
	if strings.HasPrefix(src, PrefixCryptSha256) {
		prefix = PrefixCryptSha256
	}

	rest := strings.TrimPrefix(src, prefix)
	mparts := strings.SplitN(rest, "$", 3)
	if len(mparts) < 2 {
		return nil, fmt.Errorf("malformed crypt-SHA password: %s", src)
	}

	var rounds, salt, hashed string
	// Do we have a "rounds-component"
	if len(mparts) == 3 {
		rounds, salt, hashed = mparts[0], mparts[1], mparts[2]
	} else {
		salt, hashed = mparts[0], mparts[1]
	}

	if len(salt) > 16 {
		salt = salt[0:16]
	}
	return &cryptPassword{prefix, rounds, salt, hashed}, nil
}

// PK04832_45b047bab2bf:$6$rounds=5000$e4fb4910470fd97e$afWSvXIlcC4KnENaYStPG/ELJ.uBAnG7r/rFz8fkNwpkU.salSCchDjtxyh.qA.fftcd5hmIcem7A4oA76HCE0

// RejectCryptSha known indexes
func RejectCryptSha(src string) (EncodedPasswd, error) {
	if !strings.HasPrefix(src, PrefixCryptSha512) && !strings.HasPrefix(src, PrefixCryptSha256) {
		return nil, nil
	}
	return nil, fmt.Errorf("crypt-sha password rejected: %s", src)
}

func shaCrypt(password string, rounds string, salt string, prefix string) string {

	var ret string
	var sb strings.Builder
	sb.WriteString(prefix)
	if len(rounds) > 0 {
		sb.WriteString(rounds)
		sb.WriteString(Separator)
	}
	sb.WriteString(salt)
	totalSalt := sb.String()

	if prefix == PrefixCryptSha512 {
		crypt := crypt.SHA512.New()
		ret, _ = crypt.Generate([]byte(password), []byte(totalSalt))

	} else if prefix == PrefixCryptSha256 {
		crypt := crypt.SHA256.New()
		ret, _ = crypt.Generate([]byte(password), []byte(totalSalt))
	}

	return ret[len(totalSalt)+1:]
}

func (m *cryptPassword) MatchesPassword(pw string) bool {
	hashed := shaCrypt(pw, m.rounds, m.salt, m.prefix)
	return constantTimeEquals(hashed, m.hashed)
}
