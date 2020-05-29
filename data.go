package qrn

import (
	"bufio"
	"fmt"
	"io"
	"math/rand"
	"os"
	"time"

	"github.com/valyala/fastjson"
)

const ThrottleInterrupt = 1 * time.Millisecond

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
	originLimit := time.Duration(0)

	if data.Rate > 0 {
		originLimit = time.Second / time.Duration(data.Rate+1)
	}

	reader := bufio.NewReader(file)

	if data.Random {
		reader.ReadLine()
	}

	ticker := time.NewTicker(ThrottleInterrupt)
	defer ticker.Stop()
	start := time.Now()
	limit := originLimit
	var tx int64 = 0
	throttleStart := time.Now()

	for {
		for {
			line, _, err := reader.ReadLine()

			if err == io.EOF {
				break
			} else if err != nil {
				return fmt.Errorf("%w: key=%s, json=%s", err, data.Key, string(line))
			}

			json, err := parser.ParseBytes(line)

			if err != nil {
				return fmt.Errorf("%w: key=%s, json=%s", err, data.Key, string(line))
			}

			rawQuery := json.GetStringBytes(data.Key)
			query := string(rawQuery)
			cont, err := block(query)

			if !cont || err != nil {
				if err != nil {
					err = fmt.Errorf("%w: key=%s, json=%s", err, data.Key, string(line))
				}

				return err
			}

			tx++

			select {
			case <-ticker.C:
				throttleEnd := time.Now()
				elapsed := throttleEnd.Sub(throttleStart)
				actual := elapsed / time.Duration(tx)
				limit += (originLimit - actual)

				if limit < 0 {
					limit = 0
				}

				throttleStart = throttleEnd
				tx = 0
			default:
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

		reader = bufio.NewReader(file)
	}

	return nil
}
