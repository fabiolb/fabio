package uuid

import (
	"github.com/rogpeppe/fastuuid"
)

var generator = fastuuid.MustNewGenerator()

// NewUUID return UUID in string fromat
func NewUUID() string {
	return ToString(generator.Next())
}
