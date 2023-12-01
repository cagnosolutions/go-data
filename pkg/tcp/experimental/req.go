package experimental

import (
	"bytes"
	"encoding/binary"
	"io"
)

type Request struct {
	Status uint32
	Header string
	Body   io.ReadCloser
}

func (r *Request) MarshalBinary() (data []byte, err error) {
	buf := new(bytes.Buffer)
	err = binary.Write(buf, binary.BigEndian, r.Status)
	if err != nil {
		return nil, err
	}
	err = binary.Write(buf, binary.BigEndian, r.Header)
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	err = binary.Write(buf, binary.BigEndian, body)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (r *Request) UnmarshalBinary(data []byte) error {
	buf := bytes.NewReader(data)
	err := binary.Read(buf, binary.BigEndian, r.Status)
	if err != nil {
		return err
	}
	err = binary.Read(buf, binary.BigEndian, r.Header)
	if err != nil {
		return err
	}
	err = binary.Read(buf, binary.BigEndian, r.Body)
	if err != nil {
		return err
	}
	return nil
}
