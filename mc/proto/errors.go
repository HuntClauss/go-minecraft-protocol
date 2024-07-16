package proto

import "errors"

var (
	ErrVarIntTooBig = errors.New("varint is too big")
)
