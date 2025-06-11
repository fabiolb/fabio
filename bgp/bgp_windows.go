package bgp

import (
	"errors"
	"github.com/fabiolb/fabio/config"
)

type BGPHandler struct{}

var ErrNoWindows = errors.New("cannot run bgp on windows")

func NewBGPHandler(config *config.BGP) (*BGPHandler, error) {
	return nil, ErrNoWindows
}

func (bgph *BGPHandler) Start() error {
	return ErrNoWindows
}

func ValidateConfig(config *config.BGP) error {
	return ErrNoWindows
}
