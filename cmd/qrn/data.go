package main

import (
	"bufio"
	"fmt"
	"io/ioutil"

	jsoniter "github.com/json-iterator/go"
)

type JSONLine struct {
	Query string `json:"query"`
}

func queryToFile(query string) (string, error) {
	tmpfile, err := ioutil.TempFile("", "qrn.*.jsonl")

	if err != nil {
		return tmpfile.Name(), err
	}

	defer tmpfile.Close()

	writer := bufio.NewWriter(tmpfile)

	rawLine := JSONLine{Query: query}
	line, _ := jsoniter.MarshalToString(rawLine)
	fmt.Fprintln(writer, line)
	writer.Flush()

	return tmpfile.Name(), nil
}
