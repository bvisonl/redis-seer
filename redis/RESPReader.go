/**
 * The RESPReader.go implementation was copied from the following site:
 * https://www.redisgreen.net/blog/reading-and-writing-redis-protocol/
 *
 * I found that for my purposes this utility was enough so I used this one
 * instead of rewriting it.
 */

package redis

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strconv"
	"strings"
)

const (
	SIMPLE_STRING         = '+'
	BULK_STRING           = '$'
	INTEGER               = ':'
	ARRAY                 = '*'
	ERROR                 = '-'
	COMMAND_TARGET_MASTER = "Master"
	COMMAND_TARGET_SLAVE  = "Slave"
)

var (
	ErrInvalidSyntax = errors.New("resp: invalid syntax")
)

type RESPReader struct {
	*bufio.Reader
}

func NewReader(reader io.Reader) *RESPReader {
	return &RESPReader{
		Reader: bufio.NewReaderSize(reader, 32*1024),
	}
}
func (r *RESPReader) ReadObject() ([]byte, error) {
	line, err := r.readLine()
	if err != nil {
		return nil, err
	}

	switch line[0] {
	case SIMPLE_STRING, INTEGER, ERROR:
		return line, nil
	case BULK_STRING:
		return r.readBulkString(line)
	case ARRAY:
		return r.readArray(line)
	default:
		return nil, ErrInvalidSyntax
	}
}

// GetTarget - Should return if a command should go either to a Slave or to a Master
func GetTarget(command []byte) (target string, err error) {

	cmd := strings.Split(string(command), "\r\n")

	switch string(cmd[2]) {
	case "SET":
		return COMMAND_TARGET_MASTER, nil
	default:
		return COMMAND_TARGET_SLAVE, nil
	}

	return "", errors.New("Invalid command")
}

func (r *RESPReader) readLine() (line []byte, err error) {
	line, err = r.ReadBytes('\n')
	if err != nil {
		return nil, err
	}

	if len(line) > 1 && line[len(line)-2] == '\r' {
		return line, nil
	} else {
		// Line was too short or \n wasn't preceded by \r.
		return nil, ErrInvalidSyntax
	}
}

func (r *RESPReader) readBulkString(line []byte) ([]byte, error) {
	count, err := r.getCount(line)
	if err != nil {
		return nil, err
	}
	if count == -1 {
		return line, nil
	}

	buf := make([]byte, len(line)+count+2)
	copy(buf, line)
	_, err = io.ReadFull(r, buf[len(line):])
	if err != nil {
		return nil, err
	}

	return buf, nil
}

func (r *RESPReader) getCount(line []byte) (int, error) {
	end := bytes.IndexByte(line, '\r')
	return strconv.Atoi(string(line[1:end]))
}

func (r *RESPReader) readArray(line []byte) ([]byte, error) {
	// Get number of array elements.
	count, err := r.getCount(line)
	if err != nil {
		return nil, err
	}

	// Read `count` number of RESP objects in the array.
	for i := 0; i < count; i++ {
		buf, err := r.ReadObject()
		if err != nil {
			return nil, err
		}
		line = append(line, buf...)
	}

	return line, nil
}
