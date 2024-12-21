package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"slices"

	"github.com/sirupsen/logrus"
)

const RECURSION_FLAG uint16 = 256

type header struct {
	ID      uint16
	Flags   uint16
	QdCount uint16
	ANCount uint16
	NSCount uint16
	ARCount uint16
}

func (h *header) ToBytes() []byte {
	encodedHeader := new(bytes.Buffer)
	binary.Write(encodedHeader, binary.BigEndian, h.ID)
	binary.Write(encodedHeader, binary.BigEndian, h.Flags)
	binary.Write(encodedHeader, binary.BigEndian, h.QdCount)
	binary.Write(encodedHeader, binary.BigEndian, h.ANCount)
	binary.Write(encodedHeader, binary.BigEndian, h.NSCount)
	binary.Write(encodedHeader, binary.BigEndian, h.ARCount)
	return encodedHeader.Bytes()
}

type question struct {
	QName  uint16
	QType  uint16
	QClass uint16
}

func (q *question) ToBytes() []byte {
	encodedQuestion := new(bytes.Buffer)
	binary.Write(encodedQuestion, binary.BigEndian, q.QName)
	binary.Write(encodedQuestion, binary.BigEndian, q.QType)
	binary.Write(encodedQuestion, binary.BigEndian, q.QClass)

	return encodedQuestion.Bytes()
}

func encodeDnsName(qname []byte) []byte {
	var encoded []byte
	parts := bytes.Split([]byte(qname), []byte{'.'})
	for _, part := range parts {
		encoded = append(encoded, byte(len(part)))
		encoded = append(encoded, part...)
	}
	return append(encoded, 0x00)
}

func comparyQueryID(query, response []byte) bool {
	return slices.Equal(query[:2], response[:2])
}

func NewQuery(h *header, q *question) []byte {
	var query []byte

	query = append(query, h.ToBytes()...)
	query = append(query, q.ToBytes()...)

	return query
}

type client struct {
	serverAddress string
	port          int
}

func NewClient(address string, port int) *client {
	return &client{
		serverAddress: address,
		port:          port,
	}
}

func (c *client) SendQuery(q []byte) []byte {
	conn, err := net.Dial("udp", fmt.Sprintf("%s:%d", c.serverAddress, c.port))
	if err != nil {
		logrus.Fatalf("Dial err %v", err)
	}
	defer conn.Close()

	if _, err = conn.Write(q); err != nil {
		logrus.Fatalf("Write err %v", err)
	}

	response := make([]byte, 1024)
	l, err := conn.Read(response)
	if err != nil {
		logrus.Fatal("read error ", err)
	}

	if !comparyQueryID(q, response) {
		logrus.Fatal("poisoned")
	}

	return response[:l]

}

func ParseHeader(reader *bytes.Reader) (*header, error) {
	var h header
	binary.Read(reader, binary.BigEndian, &h.ID)
	binary.Read(reader, binary.BigEndian, &h.Flags)

	switch h.Flags & 0b111 {
	case 1:
		return nil, errors.New("error with query")

	case 2:
		return nil, errors.New("error with the server")

	case 3:
		return nil, errors.New("the domain doesn't exist")

	}

	binary.Read(reader, binary.BigEndian, &h.QdCount)
	binary.Read(reader, binary.BigEndian, &h.ANCount)
	binary.Read(reader, binary.BigEndian, &h.NSCount)
	binary.Read(reader, binary.BigEndian, &h.ARCount)

	return &h, nil
}

func main() {

}
