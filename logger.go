package qrn

import (
	"fmt"
	"io"
	"time"

	jsoniter "github.com/json-iterator/go"
)

type QueryLog struct {
	Query     string        `json:"query"`
	Time      time.Duration `json:"time"`
	Timestamp time.Time     `json:"timestamp"`
}

type Logger struct {
	Channel chan QueryLog
	Null    bool
}

type ClosableDiscard struct{}

func (_ *ClosableDiscard) Write(p []byte) (n int, err error) {
	return 0, nil
}

func (_ *ClosableDiscard) Close() error {
	return nil
}

func NewLogger(out io.WriteCloser, threshold time.Duration) *Logger {
	switch out.(type) {
	case *ClosableDiscard:
		return &Logger{Null: true}
	}

	ch := make(chan QueryLog)

	logger := &Logger{
		Channel: ch,
		Null:    false,
	}

	go func() {
		for ql := range ch {
			if ql.Time >= threshold {
				log, _ := jsoniter.MarshalToString(ql)
				fmt.Fprintln(out, log)
			}
		}

		out.Close()
	}()

	return logger
}

func (logger *Logger) Log(query string, time time.Duration, ts time.Time) {
	if logger.Null {
		return
	}

	ql := QueryLog{
		Query:     query,
		Time:      time,
		Timestamp: ts,
	}

	logger.Channel <- ql
}

func (logger *Logger) Close() {
	close(logger.Channel)
}
