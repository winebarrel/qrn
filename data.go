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
	Path       string
	Key        string
	Loop       bool
	Force      bool
	Random     bool
	Rate       int
	MaxCount   int64
	CommitRate int64
}

func (data *Data) EachLine(block func(string) (bool, error)) (int64, error) {
	file, err := os.OpenFile(data.Path, os.O_RDONLY, 0)

	if err != nil {
		return 0, err
	}

	defer file.Close()

	if data.Random {
		fileinfo, err := file.Stat()

		if err != nil {
			return 0, err
		}

		size := fileinfo.Size()
		offset := rand.Int63n(size)
		_, err = file.Seek(offset, io.SeekStart)

		if err != nil {
			return 0, err
		}
	}

	var parser fastjson.Parser
	originLimit := time.Duration(0)

	if data.Rate > 0 {
		originLimit = time.Second / time.Duration(data.Rate+1)
	}

	reader := bufio.NewReader(file)

	if data.Random {
		_, err := LongReadLine(reader)

		if err != nil {
			return 0, err
		}
	}

	ticker := time.NewTicker(ThrottleInterrupt)
	defer ticker.Stop()
	start := time.Now()
	limit := originLimit
	var tx, totalTx, loopCount int64
	throttleStart := time.Now()
	commitRate := data.CommitRate
	var nextQuery string

	if commitRate > 0 {
		commitRate += 2
		nextQuery = "BEGIN"
	}

	for {
		for {
			line := "/* query inserted by qrn */"
			var query string

			if nextQuery != "" {
				query = nextQuery
				nextQuery = ""
			} else if commitRate > 0 && totalTx%commitRate == 0 {
				query = "COMMIT"
				nextQuery = "BEGIN"
			} else {
				rawLine, err := LongReadLine(reader)

				if err == io.EOF {
					break
				} else if err != nil {
					return loopCount, fmt.Errorf("%w: key=%s, json=%s", err, data.Key, string(line))
				}

				json, err := parser.ParseBytes(rawLine)

				if err != nil {
					return loopCount, fmt.Errorf("%w: key=%s, json=%s", err, data.Key, string(line))
				}

				rawQuery := json.GetStringBytes(data.Key)
				line = string(rawLine)
				query = string(rawQuery)
			}

			cont, err := block(query)

			if !cont || err != nil {
				if err != nil {
					errmsg := fmt.Sprintf("key=%s, json=%s", data.Key, string(line))

					if data.Force {
						fmt.Fprintf(os.Stderr, "%s: %s", err, errmsg)
						start = time.Now()
						continue
					} else {
						err = fmt.Errorf("%w: %s", err, errmsg)
					}
				}

				return loopCount, err
			}

			tx++
			totalTx++

			if data.MaxCount > 0 && totalTx >= data.MaxCount {
				return loopCount, nil
			}

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
			return loopCount, err
		}

		reader = bufio.NewReader(file)
		loopCount++
	}

	return loopCount, nil
}
