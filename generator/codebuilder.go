package main

import (
	"fmt"
)

type CodeBuilder struct {
	text string
}

func (c *CodeBuilder) AddLine(args ...interface{}) {
	if len(args) > 0 {
		c.text += fmt.Sprintf(args[0].(string), args[1:]...)
	}
	c.text += "\n"
}

func (c *CodeBuilder) Text() string {
	return c.text
}
