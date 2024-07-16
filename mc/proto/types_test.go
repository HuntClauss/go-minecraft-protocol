package proto

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"
)

func TestWriteVarInt(t *testing.T) {
	tests := []struct {
		Input int
		Want  []byte
	}{
		{0, []byte{0x00}},
		{1, []byte{0x01}},
		{2, []byte{0x02}},
		{127, []byte{0x7f}},
		{128, []byte{0x80, 0x01}},
		{255, []byte{0xff, 0x01}},
		{25565, []byte{0xdd, 0xc7, 0x01}},
		{2097151, []byte{0xff, 0xff, 0x7f}},
		{2147483647, []byte{0xff, 0xff, 0xff, 0xff, 0x07}},
		{-1, []byte{0xff, 0xff, 0xff, 0xff, 0x0f}},
		{-2147483648, []byte{0x80, 0x80, 0x80, 0x80, 0x08}},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("VarInt-Write-%d-%x", tt.Input, tt.Want), func(t *testing.T) {
			buf := bytes.NewBuffer(nil)
			v := VarInt(tt.Input)
			_, _ = v.Write(buf)
			if !reflect.DeepEqual(tt.Want, buf.Bytes()) {
				t.Errorf("Want: %v, Got: %v", tt.Want, buf.Bytes())
				return
			}
		})
	}
}

func TestReadVarInt(t *testing.T) {
	tests := []struct {
		Want  int
		Input []byte
	}{
		{0, []byte{0x00}},
		{1, []byte{0x01}},
		{2, []byte{0x02}},
		{127, []byte{0x7f}},
		{128, []byte{0x80, 0x01}},
		{255, []byte{0xff, 0x01}},
		{25565, []byte{0xdd, 0xc7, 0x01}},
		{2097151, []byte{0xff, 0xff, 0x7f}},
		{2147483647, []byte{0xff, 0xff, 0xff, 0xff, 0x07}},
		{-1, []byte{0xff, 0xff, 0xff, 0xff, 0x0f}},
		{-2147483648, []byte{0x80, 0x80, 0x80, 0x80, 0x08}},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("VarInt-Write-%d-%x", tt.Input, tt.Want), func(t *testing.T) {
			buf := bytes.NewBuffer(tt.Input)
			var v VarInt
			_, _ = v.Read(buf)
			if int32(tt.Want) != int32(v) {
				t.Errorf("Want: %v, Got: %v", tt.Want, v)
				return
			}
		})
	}
}
