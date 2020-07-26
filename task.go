package qrn

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
)

type Task struct {
	Agents  []*Agent
	Options *TaskOptions
	Token   string
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
	Force       bool
	MaxCount    int64
	Random      bool
	HBins       int
	HInterval   time.Duration
	QPSInterval time.Duration
	Logger      *Logger
}

func NewTask(options *TaskOptions) *Task {
	agents := make([]*Agent, options.NAgents)
	uuid, _ := uuid.NewRandom()

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
			Force:    options.Force,
			Random:   options.Random,
			Rate:     options.Rate,
			MaxCount: options.MaxCount,
		}

		agents[i] = &Agent{
			ConnInfo: connInfo,
			Data:     data,
			Logger:   options.Logger,
			Token:    uuid.String(),
		}
	}

	task := &Task{
		Agents:  agents,
		Options: options,
		Token:   uuid.String(),
	}

	return task
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
		Files:       task.Options.Files,
		NAgents:     task.Options.NAgents,
		Rate:        task.Options.Rate,
		HBins:       task.Options.HBins,
		HInterval:   task.Options.HInterval,
		QPSInterval: task.Options.QPSInterval,
		Token:       task.Token,
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
	recorder.Start(task.Options.NAgents * 3)
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
