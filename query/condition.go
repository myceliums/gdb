package query

import (
	"fmt"
	"io"
)

type Condition interface {
	Cdn(wr io.Writer, arg1 Field, arg2 ...interface{})
}

type primitiveCondition struct {
	c string
}

func (x primitiveCondition) Cdn(wr io.Writer, col Field, args ...interface{}) {
	fmt.Fprintf(wr, "")
}
