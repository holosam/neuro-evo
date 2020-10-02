package env

import (
	"testing"
)

func TestAdder(t *testing.T) {
	econf := DefaultEnvConfig()
	econf.NumPlaygrounds = 5
	Adder(econf)
}
