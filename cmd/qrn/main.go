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

var spinner = []string{"|", "/", "-", "\\"}

func init() {
	log.SetFlags(log.LstdFlags)
}

func main() {
	flags := parseFlags()

	if flags.Script != "" {
		path, err := evalScript(flags.Script)

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

	if err != nil {
		log.Fatal(err)
	}

	err = task.Prepare()

	if err != nil {
		log.Fatalf("task prepare error: %s", err)
	}

	recorder, err := task.Run(flags.Time, ReportPeriod*time.Second, withProgress(func(count int, qps float64, width int, spnr string) {
		status := fmt.Sprintf("%s run %d queries (%.0f qps)", spnr, count, qps)
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

func withProgress(block func(int, float64, int, string)) func(*qrn.Recorder) {
	s := spinner[0:]
	prev := 0

	return func(r *qrn.Recorder) {
		count := r.Count()
		qps := float64(count-prev) / ReportPeriod
		spnr := s[0]
		s = append(s[1:], spnr)
		width, _, _ := terminal.GetSize(0)
		block(count, qps, width, spnr)
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
