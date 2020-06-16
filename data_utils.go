package qrn

import (
	"bufio"
)

var ReadLineBufSize = 4096

func LongReadLine(reader *bufio.Reader) ([]byte, error) {
	buf := make([]byte, 0, ReadLineBufSize)
	var err error

	for {
		line, isPrefix, e := reader.ReadLine()
		err = e

		if len(line) > 0 {
			buf = append(buf, line...)
		}

		if !isPrefix || err != nil {
			break
		}
	}

	return buf, err
}
