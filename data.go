package qrn

import (
	"bufio"
	"io"
	"math/rand"
	"os"
	"time"

	"github.com/valyala/fastjson"
)

type Data struct {
	Path   string
	Key    string
	Loop   bool
	Random bool
	Rate   int
}

func (data *Data) EachLine(block func(string) (bool, error)) error {
	file, err := os.OpenFile(data.Path, os.O_RDONLY, 0)

	if err != nil {
		return err
	}

	defer file.Close()

	if data.Random {
		fileinfo, err := file.Stat()

		if err != nil {
			return err
		}

		size := fileinfo.Size()
		offset := rand.Int63n(size)
		_, err = file.Seek(offset, io.SeekStart)

		if err != nil {
			return err
		}
	}

	var parser fastjson.Parser
	limit := time.Duration(0)

	if data.Rate > 0 {
		limit = time.Second / time.Duration(data.Rate+1)
	}

	scanner := bufio.NewScanner(file)

	if data.Random {
		scanner.Scan()
	}

	start := time.Now()

	for {
		for scanner.Scan() {
			line := scanner.Bytes()
			json, err := parser.ParseBytes(line)

			if err != nil {
				return err
			}

			rawQuery := json.GetStringBytes(data.Key)
			query := string(rawQuery)
			cont, err := block(query)

			if !cont || err != nil {
				return err
			}

			end := time.Now()
			time.Sleep(limit - end.Sub(start))
			start = time.Now()
		}

		if !data.Loop {
			break
		}

		_, err := file.Seek(0, io.SeekStart)

		if err != nil {
			return err
		}

		scanner = bufio.NewScanner(file)
	}

	return nil
}
