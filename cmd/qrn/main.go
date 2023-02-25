package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"qrn"
	"strings"
	"time"

	"golang.org/x/term"
)

const ReportPeriod = 1
const HTMLReportName = "qrn-%d.html"

func init() {
	log.SetFlags(log.LstdFlags)
}

func main() {
	flags := parseFlags()

	if flags.Query != "" {
		path, err := queryToFile(flags.Query)

		if path != "" {
			defer os.Remove(path)
		}

		if err != nil {
			log.Fatalf("query error: %s", err)
		}

		flags.TaskOptions.Files = qrn.Strings{path}
	}

	task := qrn.NewTask(flags.TaskOptions)

	err := task.Prepare()

	if err != nil {
		log.Fatalf("task prepare error: %s", err)
	}

	recorder, err := task.Run(flags.Time, ReportPeriod*time.Second, withProgress(func(count int, qps float64, width int, elapsed time.Duration, running int) {
		d := elapsed.Round(time.Second)
		m := d / time.Minute
		s := (d - m*time.Minute) / time.Second
		status := fmt.Sprintf("%02d:%02d | %d agents / run %d queries (%.0f qps)", m, s, running, count, qps)
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

func withProgress(block func(int, float64, int, time.Duration, int)) func(*qrn.Recorder, int) {
	start := time.Now()
	prev := 0

	return func(r *qrn.Recorder, running int) {
		count := r.Count()
		qps := float64(count-prev) / ReportPeriod
		elapsed := time.Since(start)
		width, _, _ := term.GetSize(0)
		block(count, qps, width, elapsed, running)
		prev = count
	}
}

func showResult(flags *Flags, recorder *qrn.Recorder) error {
	report := recorder.Report()
	rawJSON, _ := json.MarshalIndent(report, "", "  ")

	w, _, err := term.GetSize(0)

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
