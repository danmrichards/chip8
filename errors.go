package main

import "fmt"

// errHandleOpcode is returned by the emulation cycle when unable to handle
// a given opcode.
type errHandleOpcode struct {
	opc uint16
	err error
}

func (e errHandleOpcode) Error() string {
	return fmt.Sprintf("error handling opcode 0x%X: %s", e.opc, e.err)
}
