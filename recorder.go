package qrn

import (
	"sync"
	"time"

	"github.com/winebarrel/tachymeter"
)

type Recorder struct {
	sync.Mutex
	Channel       chan []time.Duration
	ResponseTimes []time.Duration
	Started       time.Time
	Finished      time.Time
	Metrics       *tachymeter.Metrics
	NAgent        int
	Rate          int
	HBins         int
	HInterval     time.Duration
}

type RecordReport struct {
	Started     time.Time
	Finished    time.Time
	Elapsed     time.Duration
	Queries     int
	NAgent      int
	Rate        int
	QPS         float64
	ExpectedQPS int
	Response    *tachymeter.Metrics
}

func (recorder *Recorder) AppendResponseTimes(responseTimes []time.Duration) {
	recorder.Lock()
	defer recorder.Unlock()
	recorder.ResponseTimes = append(recorder.ResponseTimes, responseTimes...)
}

func (recorder *Recorder) Start() {
	recorder.ResponseTimes = []time.Duration{}
	ch := make(chan []time.Duration)
	recorder.Channel = ch

	go func() {
		for responseTimes := range ch {
			recorder.AppendResponseTimes(responseTimes)
		}
	}()

	recorder.Started = time.Now()
}

func (recorder *Recorder) Add(responseTimes []time.Duration) {
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
		t.AddTime(v)
	}

	recorder.Metrics = t.Calc()
}

func (recorder *Recorder) Count() int {
	recorder.Lock()
	defer recorder.Unlock()
	return len(recorder.ResponseTimes)
}

func (recorder *Recorder) Report() *RecordReport {
	nanoElapsed := recorder.Finished.Sub(recorder.Started)
	count := recorder.Count()
	report := &RecordReport{
		Started:     recorder.Started,
		Finished:    recorder.Finished,
		Elapsed:     nanoElapsed / time.Second,
		Queries:     count,
		NAgent:      recorder.NAgent,
		Rate:        recorder.Rate,
		QPS:         float64(count) / float64(nanoElapsed) * float64(time.Second),
		ExpectedQPS: recorder.NAgent * recorder.Rate,
		Response:    recorder.Metrics,
	}

	return report
}

func (recorder *Recorder) WriteHTMLFile(fname string, title string) error {
	return recorder.Metrics.WriteHTMLFile(fname, title)
}
