package clock

import (
	"fmt"
)

type ErrNotRenewed struct {
}

func (e *ErrNotRenewed) Error() string {
	return "leader lease not renewed"
}

type ErrElectionError struct {
	err error
}

func (e *ErrElectionError) Error() string {
	return fmt.Sprintf("Election error: %s", e.err)
}
