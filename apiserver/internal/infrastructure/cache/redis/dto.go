package redis

import (
	"bytes"
	"encoding/binary"
	"errors"
)

type UserDto struct {
	Id           int64  `redis:"id"`
	Name         string `redis:"name"`
	PasswordHash []byte `redis:"password_hash"`
	Salt         []byte `redis:"salt"`
}

func (u *UserDto) MarshalBinary() ([]byte, error) {
	buf := new(bytes.Buffer)

	if err := binary.Write(buf, binary.BigEndian, u.Id); err != nil {
		return nil, err
	}

	nameBytes := []byte(u.Name)
	if err := binary.Write(buf, binary.BigEndian, uint64(len(nameBytes))); err != nil {
		return nil, err
	}

	if _, err := buf.Write(nameBytes); err != nil {
		return nil, err
	}

	if err := binary.Write(buf, binary.BigEndian, uint64(len(u.PasswordHash))); err != nil {
		return nil, err
	}
	if _, err := buf.Write(u.PasswordHash); err != nil {
		return nil, err
	}

	if err := binary.Write(buf, binary.BigEndian, uint64(len(u.Salt))); err != nil {
		return nil, err
	}
	if _, err := buf.Write(u.Salt); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (u *UserDto) UnmarshalBinary(data []byte) error {
	buf := bytes.NewReader(data)

	if err := binary.Read(buf, binary.BigEndian, &u.Id); err != nil {
		return err
	}

	var nameLen uint64
	if err := binary.Read(buf, binary.BigEndian, &nameLen); err != nil {
		return err
	}
	nameBytes := make([]byte, nameLen)

	if n, err := buf.Read(nameBytes); err != nil || uint64(n) != nameLen {
		return errors.New("invalid name data")
	}

	u.Name = string(nameBytes)

	var hashLen uint64
	if err := binary.Read(buf, binary.BigEndian, &hashLen); err != nil {
		return err
	}

	u.PasswordHash = make([]byte, hashLen)

	if n, err := buf.Read(u.PasswordHash); err != nil || uint64(n) != hashLen {
		return errors.New("invalid password hash data")
	}

	var saltLen uint64
	if err := binary.Read(buf, binary.BigEndian, &saltLen); err != nil {
		return err
	}

	u.Salt = make([]byte, saltLen)
	if n, err := buf.Read(u.Salt); err != nil || uint64(n) != saltLen {
		return errors.New("invalid salt data")
	}

	return nil
}
