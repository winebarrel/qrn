package qrn

import (
	"context"
	"time"

	"golang.org/x/sync/errgroup"
)

type Task struct {
	Agents  []*Agent
	Options *TaskOptions
}

type TaskOptions struct {
	Driver      string
	DSN         string
	NAgents     int
	Rate        int
	File        string
	Key         string
	Loop        bool
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

	for i := 0; i < options.NAgents; i++ {
		data := &Data{
			Path:   options.File,
			Key:    options.Key,
			Loop:   options.Loop,
			Random: options.Random,
			Rate:   options.Rate,
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

func (task *Task) Run(n time.Duration, reportPeriod time.Duration, report func(*Recorder)) (*Recorder, error) {
	recorder := &Recorder{
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

	for _, agent := range task.Agents {
		eg.Go(func() error {
			err := agent.Run(ctxWithCancel, recorder)
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
				report(recorder)
			}
		}
	}()

	eg.Go(func() error {
		select {
		case <-ctx.Done():
			// nothing to do
		case <-time.After(n):
			cancel()
		}

		return nil
	})

	return recorder, eg.Wait()
}
