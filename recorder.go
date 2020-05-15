package qrn

import (
	"sort"
	"sync"
	"time"

	"github.com/winebarrel/tachymeter"
)

type Recorder struct {
	sync.Mutex
	Channel       chan []DataPoint
	ResponseTimes []DataPoint
	Started       time.Time
	Finished      time.Time
	Metrics       *tachymeter.Metrics
	NAgent        int
	Rate          int
	HBins         int
	HInterval     time.Duration
	QPSHistory    []float64
	QPSInternal   time.Duration
}

type RecordReport struct {
	Started     time.Time
	Finished    time.Time
	Elapsed     time.Duration
	Queries     int
	NAgent      int
	Rate        int
	QPS         float64
	MaxQPS      float64
	ExpectedQPS int
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

func (recorder *Recorder) Start() {
	recorder.ResponseTimes = []DataPoint{}
	ch := make(chan []DataPoint)
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
	qpsHist := recorder.QPSHistory[1:]

	report := &RecordReport{
		Started:     recorder.Started,
		Finished:    recorder.Finished,
		Elapsed:     nanoElapsed / time.Second,
		Queries:     count,
		NAgent:      recorder.NAgent,
		Rate:        recorder.Rate,
		QPS:         float64(count) * float64(time.Second) / float64(nanoElapsed),
		ExpectedQPS: recorder.NAgent * recorder.Rate,
		Response:    recorder.Metrics,
	}

	if len(qpsHist) > 0 {
		report.MaxQPS = qpsHist[0]

		for _, v := range qpsHist {
			if v > report.MaxQPS {
				report.MaxQPS = v
			}
		}
	}

	return report
}

func (recorder *Recorder) WriteHTMLFile(fname string, title string) error {
	return recorder.Metrics.WriteHTMLFile(fname, title)
}
