package main

import (
	"bufio"
	"fmt"
	"io/ioutil"

	jsoniter "github.com/json-iterator/go"
	"github.com/robertkrimen/otto"
)

type JSONLine struct {
	Query string `json:"query"`
}

func evalScript(path string) (string, error) {
	script, err := ioutil.ReadFile(path)

	if err != nil {
		return "", err
	}

	tmpfile, err := ioutil.TempFile("", "qrn.*.jsonl")

	if err != nil {
		return tmpfile.Name(), err
	}

	defer tmpfile.Close()
	writer := bufio.NewWriter(tmpfile)

	vm := otto.New()

	err = vm.Set("query", func(call otto.FunctionCall) otto.Value {
		query := call.Argument(0).String()
		rawLine := JSONLine{Query: query}
		line, _ := jsoniter.MarshalToString(rawLine)
		fmt.Fprintln(writer, line)
		return otto.Value{}
	})

	if err != nil {
		return tmpfile.Name(), err
	}

	_, err = vm.Run(script)

	if err != nil {
		return tmpfile.Name(), err
	}

	if err != nil {
		return tmpfile.Name(), err
	}

	writer.Flush()

	return tmpfile.Name(), nil
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
