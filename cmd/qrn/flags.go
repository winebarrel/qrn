package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"qrn"
	"strconv"
	"time"
)

var version string

const DefaultDriver = "mysql"
const DefaultTime = 60
const DefaultJsonKey = "query"
const DefaultHBins = 10

type Flags struct {
	Time        time.Duration
	Histogram   bool
	HTML        bool
	Query       string
	TaskOptions *qrn.TaskOptions
}

type xBool struct {
	set   bool
	value bool
}

func (b *xBool) String() string {
	return strconv.FormatBool(b.value)
}

func (b *xBool) Set(s string) error {
	v, err := strconv.ParseBool(s)

	if err != nil {
		return err
	}

	*b = xBool{true, v}
	return nil
}

func parseFlags() (flags *Flags) {
	flags = &Flags{
		TaskOptions: &qrn.TaskOptions{},
	}

	var random xBool

	flag.StringVar(&flags.TaskOptions.Driver, "driver", DefaultDriver, "database driver")
	flag.StringVar(&flags.TaskOptions.DSN, "dsn", "", "data source name")
	flag.IntVar(&flags.TaskOptions.NAgents, "nagents", 0, "number of agents")
	argTime := flag.Int("time", DefaultTime, "test run time (sec). zero is unlimited")
	flag.Var(&flags.TaskOptions.Files, "data", "file path of execution queries for each agent")
	flag.StringVar(&flags.Query, "query", "", "execution query")
	logOpt := flag.String("log", "", "file path of query log")
	logTime := flag.String("logtime", "0", "execution time threshold for logged queries")
	flag.IntVar(&flags.TaskOptions.Rate, "rate", 0, "rate limit for each agent (qps). zero is unlimited")
	flag.StringVar(&flags.TaskOptions.Key, "key", DefaultJsonKey, "json key of query")
	flag.BoolVar(&flags.TaskOptions.Loop, "loop", true, "input data loop flag")
	flag.BoolVar(&flags.TaskOptions.Force, "force", false, "ignore query error")
	flag.Int64Var(&flags.TaskOptions.MaxCount, "maxcount", 0, "maximum number of queries for each agent. zero is unlimited")
	flag.Var(&random, "random", "randomize the start position of input data")
	flag.IntVar(&flags.TaskOptions.HBins, "hbins", DefaultHBins, "histogram bins")
	hinterval := flag.String("hinterval", "0", "histogram interval")
	flag.BoolVar(&flags.Histogram, "histogram", false, "show histogram")
	flag.BoolVar(&flags.HTML, "html", false, "output histogram html")
	argVersion := flag.Bool("version", false, "Print version and exit")
	flag.Parse()

	flen := len(flags.TaskOptions.Files)

	if flag.NFlag() == 0 {
		printUsageAndExit()
	}

	if *argVersion {
		printVersionAndEixt()
	}

	if flags.TaskOptions.DSN == "" {
		printErrorAndExit("'-dsn' is required")
	}

	if flags.TaskOptions.NAgents < 1 {
		if flen > 1 {
			flags.TaskOptions.NAgents = flen
		} else {
			flags.TaskOptions.NAgents = 1
		}
	}

	if *argTime < 0 {
		printErrorAndExit("'-time' must be >= 0")
	}

	flags.Time = time.Duration(*argTime) * time.Second

	if flags.TaskOptions.Rate < 0 {
		printErrorAndExit("'-rate' must be >= 0")
	}

	if flags.TaskOptions.MaxCount < 0 {
		printErrorAndExit("'-maxcount' must be >= 0")
	}

	if flen == 0 && flags.Query == "" {
		printErrorAndExit("'-data' or '-query' is required")
	} else if flen != 0 && flags.Query != "" {
		printErrorAndExit("please specify one of '-data' or '-query'")
	}

	if flags.TaskOptions.Key == "" {
		printErrorAndExit("'-key' dose not allow empty")
	}

	if flags.Query != "" {
		flags.TaskOptions.Key = DefaultJsonKey
	}

	if random.set {
		flags.TaskOptions.Random = random.value
	} else if flags.TaskOptions.Loop {
		flags.TaskOptions.Random = true
	} else {
		flags.TaskOptions.Random = false
	}

	if flags.TaskOptions.Random {
		rand.Seed(time.Now().UnixNano())
	}

	if hi, err := time.ParseDuration(*hinterval); err != nil {
		printErrorAndExit(err.Error())
	} else {
		flags.TaskOptions.HInterval = hi
	}

	if *logOpt == "" {
		devNull := &qrn.ClosableDiscard{}
		logger := qrn.NewLogger(devNull, 0)
		flags.TaskOptions.Logger = logger
	} else {
		file, err := os.OpenFile(*logOpt, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)

		if err != nil {
			printErrorAndExit(err.Error())
		}

		if lt, err := time.ParseDuration(*logTime); err != nil {
			printErrorAndExit(err.Error())
		} else {
			logger := qrn.NewLogger(file, lt)
			flags.TaskOptions.Logger = logger
		}
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
