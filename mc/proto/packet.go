package proto

import (
	"bytes"
	"compress/zlib"
	"io"
	"reflect"
)

type Packet struct {
	ID   int
	Data []byte
}

type Experim struct {
	ID      VarInt
	Version VarInt
	Addr    String
	Port    UShort
}

func NewPacket(id int) Packet {
	return Packet{
		ID:   id,
		Data: make([]byte, 0, 10240),
	}
}

func NewPacketFromReader(r io.Reader) Packet {
	size, id := VarInt(0), VarInt(0)
	_, _ = size.ReadFrom(r)
	nn, _ := id.ReadFrom(r)

	buf := make([]byte, int64(size)-nn)
	n, _ := r.Read(buf)
	return Packet{
		ID:   int(id),
		Data: buf[:n],
	}
}

func NewCompressPacketFromReader(r io.Reader) (Packet, error) {
	totalLength, dataLength := VarInt(0), VarInt(0)
	if _, err := totalLength.ReadFrom(r); err != nil {
		return Packet{ID: -1}, err
	}

	n, err := dataLength.ReadFrom(r)
	if err != nil {
		return Packet{ID: -1}, err
	}

	buf := make([]byte, int64(totalLength)-n)

	read := 0
	x, err := r.Read(buf)
	if err != nil {
		return Packet{ID: -1}, err
	}

	read += x
	for int(totalLength)-read > 0 && x > 0 {
		x, err = r.Read(buf[read:])
		if err != nil && err != io.EOF {
			return Packet{ID: -1}, err
		}

		read += x
	}
	buf = buf[:read]

	//fmt.Println(int(totalLength)-int(n), len(buf))
	//return Packet{ID: -2}
	//fmt.Println(totalLength, dataLength, buf)

	body := make([]byte, dataLength)
	if dataLength != 0 {
		ss := bytes.NewReader(buf)
		reader, err := zlib.NewReader(ss)
		if err != nil {
			return Packet{ID: -1}, err
		}
		n, _ := reader.Read(body)
		body = body[:n]
	} else {
		body = buf
	}

	b := bytes.NewBuffer(body)
	id := VarInt(-1)
	n, _ = id.ReadFrom(b)
	pk := NewPacket(int(id))
	pk.Data = body[n:]
	return pk, nil
}

func (p *Packet) Append(s ...any) error {
	buf := bytes.NewBuffer(nil)
	input := []reflect.Value{reflect.ValueOf(buf)}

	for _, v := range s {
		r := reflect.ValueOf(v)
		if r.Kind() == reflect.Pointer {
			r = r.Elem()
		}

		if r.Kind() == reflect.Struct {
			for i := 0; i < r.NumField(); i++ {
				field := r.Field(i).Addr()
				method := field.MethodByName("WriteTo")
				method.Call(input)
			}
		} else {
			_, _ = v.(io.WriterTo).WriteTo(buf)
		}
	}

	p.Data = append(p.Data, buf.Bytes()...)
	return nil
}

func (p *Packet) Scan(s ...any) error {
	buf := bytes.NewBuffer(p.Data)
	input := []reflect.Value{reflect.ValueOf(buf)}

	for _, v := range s {
		r := reflect.ValueOf(v)
		if r.Kind() == reflect.Pointer {
			r = r.Elem()
		}

		if r.Kind() == reflect.Struct {
			for i := 0; i < r.NumField(); i++ {
				field := r.Field(i).Addr()
				method := field.MethodByName("ReadFrom")
				method.Call(input)
			}
		} else {
			_, _ = v.(io.ReaderFrom).ReadFrom(buf)
		}
	}
	return nil
}

func (p *Packet) Bytes() []byte {
	buf := bytes.NewBuffer(make([]byte, 0, 10240))
	id := VarInt(p.ID)
	n, _ := id.WriteTo(buf)
	buf.Reset()

	size := VarInt(n + int64(len(p.Data)))
	size.WriteTo(buf)
	id.WriteTo(buf)
	buf.Write(p.Data)
	return buf.Bytes()
}

func (p *Packet) CompressBytes(threshold int) []byte {
	buf := bytes.NewBuffer(make([]byte, 0, 10240))
	id := VarInt(p.ID)
	n, _ := id.WriteTo(buf)
	buf.Reset()
	dataLength := VarInt(n + int64(len(p.Data)))

	data := bytes.NewBuffer(nil)
	if int64(dataLength) >= int64(threshold) {
		w := zlib.NewWriter(data)
		defer w.Close()

		id.WriteTo(w)
		w.Write(p.Data)
	} else {
		dataLength = VarInt(0)
		id.WriteTo(data)
		data.Write(p.Data)
	}

	realDataSize := VarInt(len(data.Bytes()))

	n, _ = dataLength.WriteTo(buf)
	totalSize := VarInt(n) + realDataSize
	totalSize.WriteTo(buf)
	dataLength.WriteTo(buf)
	buf.Write(data.Bytes())
	return buf.Bytes()
}

//func (p *Packet) Scan
