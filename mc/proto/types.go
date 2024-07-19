package proto

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"math"
	"strings"
)

type (
	VarInt int32
	Short  int16
	UShort uint16
	String string
	Bool   bool
	Byte   int8
	UByte  uint8
	Uuid   [2]uint64
	Long   int64
	Float  float32
	Double float64
)

func NewVarInt(v int) *VarInt {
	def := VarInt(v)
	return &def
}

func NewShort(v int) *Short {
	def := Short(v)
	return &def
}

func NewUShort(v int) *UShort {
	def := UShort(v)
	return &def
}

func NewString(v string) *String {
	def := String(v)
	return &def
}

func NewBool(v bool) *Bool {
	def := Bool(v)
	return &def
}

func NewByte(v byte) *Byte {
	def := Byte(v)
	return &def
}

func NewUuidFromStr(v string) *Uuid {
	out := Uuid{}
	v = strings.Replace(v, "-", "", -1)
	data, err := hex.DecodeString(v)
	if err != nil {
		panic(fmt.Sprintf("invalid uuid string: %s", err))
	}

	buf := bytes.NewBuffer(data)
	_, _ = out.ReadFrom(buf)
	return &out
}

func NewUuidFromBytes(v []byte) *Uuid {
	out := &Uuid{}
	buf := bytes.NewBuffer(v)
	_, _ = out.ReadFrom(buf)
	return out
}

func int64Wrap(n int, err error) (int64, error) {
	return int64(n), err
}

func (v *VarInt) WriteTo(w io.Writer) (int64, error) {
	const SegmentBit, ContinueBit uint32 = 0x7F, 0x80
	val := uint32(*v)

	buf := make([]byte, 0, 5)
	for {
		if (val & ^SegmentBit) == 0 {
			buf = append(buf, uint8(val))
			break
		}

		buf = append(buf, uint8((val&SegmentBit)|ContinueBit))
		val >>= 7
	}
	return int64Wrap(w.Write(buf))
}

func (v *VarInt) ReadFrom(r io.Reader) (int64, error) {
	const SegmentBit, ContinueBit uint8 = 0x7F, 0x80

	val, pos, size := uint32(0), 0, int64(0)
	buf := make([]byte, 1)

	for {
		n, err := int64Wrap(r.Read(buf))
		size += n
		if err != nil {
			return size, err
		}

		val |= uint32(buf[0]&SegmentBit) << pos

		if (buf[0] & ContinueBit) == 0 {
			break
		}
		pos += 7

		if pos >= 32 {
			return size, ErrVarIntTooBig
		}
	}
	*v = VarInt(val)
	return size, nil
}

func (s *Short) WriteTo(w io.Writer) (int64, error) {
	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf, uint16(*s))
	return int64Wrap(w.Write(buf))
}

func (s *Short) ReadFrom(r io.Reader) (int64, error) {
	buf := make([]byte, 2)
	if n, err := int64Wrap(r.Read(buf)); err != nil {
		return n, err
	}

	*s = Short(binary.BigEndian.Uint16(buf))
	return 2, nil
}

func (s *UShort) WriteTo(w io.Writer) (int64, error) {
	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf, uint16(*s))
	return int64Wrap(w.Write(buf))
}

func (s *UShort) ReadFrom(r io.Reader) (int64, error) {
	buf := make([]byte, 2)
	if n, err := int64Wrap(r.Read(buf)); err != nil {
		return n, err
	}

	*s = UShort(binary.BigEndian.Uint16(buf))
	return 2, nil
}

func (s *String) WriteTo(w io.Writer) (int64, error) {
	buf := bytes.NewBuffer(make([]byte, 0, 3+len(*s)))
	l := VarInt(len(*s))
	if n, err := l.WriteTo(buf); err != nil {
		return n, err
	}

	buf.WriteString(string(*s))
	return int64Wrap(w.Write(buf.Bytes()))
}

func (s *String) ReadFrom(r io.Reader) (int64, error) {
	size := VarInt(0)
	nn, err := size.ReadFrom(r)
	if err != nil {
		return nn, err
	}

	buf := make([]byte, size)
	n, err := int64Wrap(r.Read(buf))
	nn += n
	if err != nil {
		return nn, err
	}

	*s = String(buf)
	return nn, err
}

func (b *Bool) WriteTo(w io.Writer) (int64, error) {
	buf := make([]byte, 1)
	buf[0] = 0
	if *b {
		buf[0] = 1
	}

	return int64Wrap(w.Write(buf))
}

func (b *Bool) ReadFrom(r io.Reader) (int64, error) {
	buf := make([]byte, 1)
	if n, err := int64Wrap(r.Read(buf)); err != nil {
		return n, err
	}

	*b = false
	if buf[0] == 1 {
		*b = true
	}
	return 1, nil
}

func (b *Byte) WriteTo(w io.Writer) (int64, error) {
	return int64Wrap(w.Write([]byte{byte(int8(*b))}))
}

func (b *Byte) ReadFrom(r io.Reader) (int64, error) {
	buf := make([]byte, 1)
	if n, err := int64Wrap(r.Read(buf)); err != nil {
		return n, err
	}

	*b = Byte(buf[0])
	return 1, nil
}

func (b *UByte) WriteTo(w io.Writer) (int64, error) {
	return int64Wrap(w.Write([]byte{byte(*b)}))
}

func (b *UByte) ReadFrom(r io.Reader) (int64, error) {
	buf := make([]byte, 1)
	if n, err := int64Wrap(r.Read(buf)); err != nil {
		return n, err
	}

	*b = UByte(buf[0])
	return 1, nil
}

func (u *Uuid) WriteTo(w io.Writer) (int64, error) {
	buf := make([]byte, 16)
	binary.BigEndian.PutUint64(buf[:8], u[0])
	binary.LittleEndian.PutUint64(buf[8:], u[1])
	return int64Wrap(w.Write(buf))
}

func (u *Uuid) ReadFrom(r io.Reader) (int64, error) {
	buf := make([]byte, 16)
	if n, err := int64Wrap(r.Read(buf)); err != nil {
		return n, err
	}

	u[0] = binary.BigEndian.Uint64(buf[:8])
	u[1] = binary.LittleEndian.Uint64(buf[8:])
	return 16, nil
}

// 69359037-9599-48e7-b8f2-48393c019135
func (u *Uuid) String() string {
	w := bytes.NewBuffer(make([]byte, 0, 16))
	_, _ = u.WriteTo(w)
	buf := w.Bytes()

	return fmt.Sprintf("%x-%x-%x-%x-%x", buf[:4], buf[4:6], buf[6:8], buf[8:10], buf[10:16])
}

func (l *Long) WriteTo(w io.Writer) (int64, error) {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(*l))
	return int64Wrap(w.Write(buf))
}

func (l *Long) ReadFrom(r io.Reader) (int64, error) {
	buf := make([]byte, 8)
	if n, err := int64Wrap(r.Read(buf)); err != nil {
		return n, err
	}

	*l = Long(binary.BigEndian.Uint64(buf))
	return 8, nil
}

func (f *Float) WriteTo(w io.Writer) (int64, error) {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, math.Float32bits(float32(*f)))
	return int64Wrap(w.Write(buf))
}

func (f *Float) ReadFrom(r io.Reader) (int64, error) {
	buf := make([]byte, 4)
	if n, err := int64Wrap(r.Read(buf)); err != nil {
		return n, err
	}

	*f = Float(math.Float32frombits(binary.BigEndian.Uint32(buf)))
	return 4, nil
}

func (d *Double) WriteTo(w io.Writer) (int64, error) {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, math.Float64bits(float64(*d)))
	return int64Wrap(w.Write(buf))
}

func (d *Double) ReadFrom(r io.Reader) (int64, error) {
	buf := make([]byte, 8)
	if n, err := int64Wrap(r.Read(buf)); err != nil {
		return n, err
	}

	*d = Double(math.Float64frombits(binary.BigEndian.Uint64(buf)))
	return 8, nil
}
