package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"qrn"
	"time"
)

var version string

const DefaultTime = 60
const DefaultJsonKey = "query"
const DefaultHBins = 10

type Flags struct {
	Time        time.Duration
	Histogram   bool
	HTML        bool
	Script      string
	TaskOptions *qrn.TaskOptions
}

func parseFlags() (flags *Flags) {
	flags = &Flags{
		TaskOptions: &qrn.TaskOptions{},
	}

	flag.StringVar(&flags.TaskOptions.DSN, "dsn", "", "data source name")
	flag.IntVar(&flags.TaskOptions.NAgents, "nagents", 1, "number of agents")
	argTime := flag.Int("time", DefaultTime, "test run time (sec)")
	flag.StringVar(&flags.TaskOptions.File, "data", "", "file path of execution queries")
	flag.StringVar(&flags.Script, "script", "", "file path of execution script")
	flag.IntVar(&flags.TaskOptions.Rate, "rate", 0, "rate limit for each agent (qps). zero is unlimited")
	flag.StringVar(&flags.TaskOptions.Key, "key", DefaultJsonKey, "json key of query")
	flag.BoolVar(&flags.TaskOptions.Loop, "loop", true, "input data loop flag")
	flag.BoolVar(&flags.TaskOptions.Random, "random", true, "randomize the start position of input data")
	flag.IntVar(&flags.TaskOptions.HBins, "hbins", DefaultHBins, "histogram bins")
	hinterval := flag.String("hinterval", "0", "histogram interval")
	flag.BoolVar(&flags.Histogram, "histogram", false, "show histogram")
	flag.BoolVar(&flags.HTML, "html", false, "output histogram html")
	argVersion := flag.Bool("version", false, "Print version and exit")
	flag.Parse()

	if flag.NFlag() == 0 {
		printUsageAndExit()
	}

	if *argVersion {
		printVersionAndEixt()
		fmt.Fprintln(os.Stderr, version)
		os.Exit(2)
	}

	if flags.TaskOptions.DSN == "" {
		printErrorAndExit("'-dsn' is required")
	}

	if flags.TaskOptions.NAgents < 1 {
		printErrorAndExit("'-nagents' must be >= 1")
	}

	if *argTime < 1 {
		printErrorAndExit("'-time' must be >= 1")
	}

	flags.Time = time.Duration(*argTime) * time.Second

	if flags.TaskOptions.Rate < 0 {
		printErrorAndExit("'-rate' must be >= 0")
	}

	if flags.TaskOptions.File == "" && flags.Script == "" {
		printErrorAndExit("'-data' or '-script' is required")
	} else if flags.TaskOptions.File != "" && flags.Script != "" {
		printErrorAndExit("cannot specify both '-data' and '-script'")
	}

	if flags.TaskOptions.Key == "" {
		printErrorAndExit("'-key' dose not allow empty")
	}

	if flags.Script != "" {
		flags.TaskOptions.Key = DefaultJsonKey
	}

	if flags.TaskOptions.Random {
		rand.Seed(time.Now().UnixNano())
	}

	if hi, err := time.ParseDuration(*hinterval); err != nil {
		printErrorAndExit(err.Error())
	} else {
		flags.TaskOptions.HInterval = hi
	}

	return
}

func printUsageAndExit() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(2)
}

func printVersionAndEixt() {
	fmt.Fprintln(os.Stderr, version)
	os.Exit(0)
}

func printErrorAndExit(msg string) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}
