package main

import "bytes"

type LineWriter struct {
	bytes.Buffer
}

func NewLineWriter() *LineWriter {
	return &LineWriter{}
}

func (w *LineWriter) WriteLine(line string) {
	w.WriteString(line + "\n")
}
