package qrn

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"golang.org/x/sync/errgroup"
)

type Task struct {
	Agents  []*Agent
	Options *TaskOptions
}

type Files []string

func (files *Files) String() string {
	return fmt.Sprintf("%v", *files)
}

func (files *Files) Set(f string) error {
	*files = append(*files, f)
	return nil
}

type TaskOptions struct {
	Driver      string
	DSN         string
	NAgents     int
	Rate        int
	Files       Files
	Key         string
	Loop        bool
	MaxCount    int64
	Random      bool
	HBins       int
	HInterval   time.Duration
	QPSInterval time.Duration
	Logger      *Logger
}

func NewTask(options *TaskOptions) (*Task, error) {
	agents := make([]*Agent, options.NAgents)

	connInfo := &ConnInfo{
		Driver:       options.Driver,
		DSN:          options.DSN,
		MaxIdleConns: options.NAgents,
	}

	files := options.Files
	flen := len(options.Files)

	for i := 0; i < options.NAgents; i++ {
		data := &Data{
			Path:     files[i%flen],
			Key:      options.Key,
			Loop:     options.Loop,
			Random:   options.Random,
			Rate:     options.Rate,
			MaxCount: options.MaxCount,
		}

		agents[i] = &Agent{
			ConnInfo: connInfo,
			Data:     data,
			Logger:   options.Logger,
		}
	}

	task := &Task{
		Agents:  agents,
		Options: options,
	}

	return task, nil
}

func (task *Task) Prepare() error {
	for _, agent := range task.Agents {
		if err := agent.Prepare(); err != nil {
			return err
		}
	}

	return nil
}

func (task *Task) Run(n time.Duration, reportPeriod time.Duration, report func(*Recorder, int)) (*Recorder, error) {
	recorder := &Recorder{
		DSN:         task.Options.DSN,
		NAgents:     task.Options.NAgents,
		Rate:        task.Options.Rate,
		HBins:       task.Options.HBins,
		HInterval:   task.Options.HInterval,
		QPSInternal: task.Options.QPSInterval,
	}

	defer func() {
		recorder.Close()

		for _, agent := range task.Agents {
			agent.Close()
		}
	}()

	eg, ctx := errgroup.WithContext(context.Background())
	ctxWithCancel, cancel := context.WithCancel(ctx)
	ticker := time.NewTicker(reportPeriod)
	recorder.Start()
	var doneCnt int32

	for _, v := range task.Agents {
		agent := v
		eg.Go(func() error {
			err := agent.Run(ctxWithCancel, recorder)
			atomic.AddInt32(&doneCnt, 1)
			return err
		})
	}

	go func() {
	LOOP:
		for {
			select {
			case <-ctx.Done():
				break LOOP
			case <-ticker.C:
				report(recorder, task.Options.NAgents-int(doneCnt))
			}
		}
	}()

	if n > 0 {
		go func() {
			select {
			case <-ctx.Done():
				// nothing to do
			case <-time.After(n):
				cancel()
			}
		}()
	}

	err := eg.Wait()
	cancel()

	return recorder, err
}
