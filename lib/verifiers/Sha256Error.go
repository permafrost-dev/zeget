package verifiers

import "fmt"

type Sha256Error struct {
	Expected []byte
	Got      []byte
}

func (e *Sha256Error) Error() string {
	return fmt.Sprintf("sha256 checksum mismatch:\nexpected: %x\ngot:      %x", e.Expected, e.Got)
}
