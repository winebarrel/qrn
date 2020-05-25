package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"qrn"
	"strings"
	"time"

	"golang.org/x/crypto/ssh/terminal"
)

const ReportPeriod = 1
const HTMLReportName = "qrn-%d.html"

func init() {
	log.SetFlags(log.LstdFlags)
}

func main() {
	flags := parseFlags()

	if flags.Script != "" || flags.Query != "" {
		var path string
		var err error

		if flags.Script != "" {
			path, err = evalScript(flags.Script)
		}

		if flags.Query != "" {
			path, err = queryToFile(flags.Query)
		}

		if path != "" {
			defer os.Remove(path)
		}

		if err != nil {
			log.Fatalf("script error: %s", err)
		}

		flags.TaskOptions.File = path
	}

	task, err := qrn.NewTask(flags.TaskOptions)

	if err != nil {
		log.Fatalf("task create error: %s", err)
	}

	err = task.Prepare()

	if err != nil {
		log.Fatalf("task prepare error: %s", err)
	}

	recorder, err := task.Run(flags.Time, ReportPeriod*time.Second, withProgress(func(count int, qps float64, width int, elapsed time.Duration) {
		d := elapsed.Round(time.Second)
		m := d / time.Minute
		s := (d - m*time.Minute) / time.Second
		status := fmt.Sprintf("%02d:%02d run %d queries (%.0f qps)", m, s, count, qps)
		fmt.Fprintf(os.Stderr, "\r%-*s", width, status)
	}))

	fmt.Fprintf(os.Stderr, "\r\n\n")

	if err != nil {
		log.Fatalf("task run error: %s", err)
	}

	err = showResult(flags, recorder)

	if err != nil {
		log.Fatalf("show result error: %s", err)
	}
}

func withProgress(block func(int, float64, int, time.Duration)) func(*qrn.Recorder) {
	start := time.Now()
	prev := 0

	return func(r *qrn.Recorder) {
		count := r.Count()
		qps := float64(count-prev) / ReportPeriod
		elapsed := time.Since(start)
		width, _, _ := terminal.GetSize(0)
		block(count, qps, width, elapsed)
		prev = count
	}
}

func showResult(flags *Flags, recorder *qrn.Recorder) error {
	report := recorder.Report()
	rawJSON, _ := json.MarshalIndent(report, "", "  ")

	w, _, err := terminal.GetSize(0)

	if err != nil {
		return err
	}

	if flags.Histogram {
		fmt.Fprintf(os.Stderr, "%s\n", report.Response.Histogram.String(w/3))
	}

	fmt.Println(string(rawJSON))

	if flags.HTML {
		fname := fmt.Sprintf(HTMLReportName, time.Now().Unix())
		err := recorder.WriteHTMLFile(fname, strings.Join(os.Args, " "))

		if err != nil {
			return err
		}

		fmt.Fprintf(os.Stderr, "\noutput %s\n", fname)
	}

	return nil
}
