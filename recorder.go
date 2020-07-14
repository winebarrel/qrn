package qrn

import (
	"sort"
	"sync"
	"time"

	"github.com/winebarrel/tachymeter"
)

type Recorder struct {
	sync.Mutex
	Files         []string
	Channel       chan []DataPoint
	ResponseTimes []DataPoint
	DSN           string
	Started       time.Time
	Finished      time.Time
	Metrics       *tachymeter.Metrics
	NAgents       int
	Rate          int
	LoopCount     int64
	HBins         int
	HInterval     time.Duration
	QPSHistory    []float64
	QPSInternal   time.Duration
}

type RecordReport struct {
	DSN         string
	Files       []string
	Started     time.Time
	Finished    time.Time
	Elapsed     time.Duration
	Queries     int
	NAgents     int
	Rate        int
	QPS         float64
	MaxQPS      float64
	MinQPS      float64
	MedianQPS   float64
	ExpectedQPS int
	LoopCount   int64
	Response    *tachymeter.Metrics
}

type DataPoint struct {
	Time         time.Time
	ResponseTime time.Duration
}

func (recorder *Recorder) AppendResponseTimes(responseTimes []DataPoint) {
	recorder.Lock()
	defer recorder.Unlock()
	recorder.ResponseTimes = append(recorder.ResponseTimes, responseTimes...)
}

func (recorder *Recorder) Start(bufsize int) {
	recorder.ResponseTimes = []DataPoint{}
	ch := make(chan []DataPoint, bufsize)
	recorder.Channel = ch

	go func() {
		for responseTimes := range ch {
			recorder.AppendResponseTimes(responseTimes)
		}
	}()

	recorder.Started = time.Now()
}

func (recorder *Recorder) Add(responseTimes []DataPoint) {
	recorder.Channel <- responseTimes
}

func (recorder *Recorder) Close() {
	close(recorder.Channel)
	recorder.Finished = time.Now()

	t := tachymeter.New(&tachymeter.Config{
		Size:      len(recorder.ResponseTimes),
		HBins:     recorder.HBins,
		HInterval: recorder.HInterval,
	})

	for _, v := range recorder.ResponseTimes {
		t.AddTime(v.ResponseTime)
	}

	recorder.Metrics = t.Calc()
	recorder.calcQPS()
}

func (recorder *Recorder) calcQPS() {
	responseTimes := recorder.ResponseTimes

	if len(responseTimes) == 0 {
		return
	}

	minTime := responseTimes[0].Time
	maxTime := responseTimes[0].Time

	for _, v := range responseTimes {
		if v.Time.Before(minTime) {
			minTime = v.Time
		}

		if v.Time.After(maxTime) {
			maxTime = v.Time
		}
	}

	sort.Slice(responseTimes, func(i, j int) bool {
		return recorder.ResponseTimes[i].Time.Before(recorder.ResponseTimes[j].Time)
	})

	interval := recorder.QPSInternal
	cntHist := []int{0}

	for _, v := range responseTimes {
		if minTime.Add(1 * time.Second).Before(v.Time) {
			minTime = minTime.Add(interval)
			cntHist = append(cntHist, 0)
		}

		cntHist[len(cntHist)-1]++
	}

	qpsHist := make([]float64, len(cntHist))

	for i, v := range cntHist {
		qpsHist[i] = float64(v) * float64(time.Second) / float64(interval)
	}

	recorder.QPSHistory = qpsHist
}

func (recorder *Recorder) Count() int {
	recorder.Lock()
	defer recorder.Unlock()
	return len(recorder.ResponseTimes)
}

func (recorder *Recorder) Report() *RecordReport {
	nanoElapsed := recorder.Finished.Sub(recorder.Started)
	count := recorder.Count()

	if len(recorder.QPSHistory) < 1 {
		return &RecordReport{}
	}

	qpsHist := recorder.QPSHistory[1:]

	report := &RecordReport{
		DSN:         recorder.DSN,
		Files:       recorder.Files,
		Started:     recorder.Started,
		Finished:    recorder.Finished,
		Elapsed:     nanoElapsed / time.Second,
		Queries:     count,
		NAgents:     recorder.NAgents,
		Rate:        recorder.Rate,
		QPS:         float64(count) * float64(time.Second) / float64(nanoElapsed),
		ExpectedQPS: recorder.NAgents * recorder.Rate,
		LoopCount:   recorder.LoopCount,
		Response:    recorder.Metrics,
	}

	if len(qpsHist) > 0 {
		report.MaxQPS = qpsHist[0]
		report.MinQPS = qpsHist[0]

		for _, v := range qpsHist {
			if v > report.MaxQPS {
				report.MaxQPS = v
			}

			if v < report.MinQPS {
				report.MinQPS = v
			}
		}

		sort.Slice(qpsHist, func(i, j int) bool {
			return qpsHist[i] < qpsHist[j]
		})

		median := len(qpsHist) / 2
		medianNext := median + 1

		if len(qpsHist) == 1 {
			report.MedianQPS = qpsHist[0]
		} else if len(qpsHist) == 2 {
			report.MedianQPS = (qpsHist[0] + qpsHist[1]) / 2
		} else if len(qpsHist)%2 == 0 {
			report.MedianQPS = (qpsHist[median] + qpsHist[medianNext]) / 2
		} else {
			report.MedianQPS = qpsHist[medianNext]
		}
	}

	return report
}

func (recorder *Recorder) WriteHTMLFile(fname string, title string) error {
	return recorder.Metrics.WriteHTMLFile(fname, title)
}
