package qrn

import (
	"context"
	"database/sql"
	"fmt"
	"sync/atomic"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v4/stdlib"
)

const AgentInterruptPeriod = 1 * time.Second

type ConnInfo struct {
	Driver       string
	DSN          string
	MaxIdleConns int
}

type Agent struct {
	Id       int
	ConnInfo *ConnInfo
	DB       *sql.DB
	Data     *Data
	Logger   *Logger
	Token    string
}

func (agent *Agent) Prepare(preQueries []string) error {
	db, err := sql.Open(agent.ConnInfo.Driver, agent.ConnInfo.DSN)

	if err != nil {
		return err
	}

	db.SetConnMaxLifetime(0)
	db.SetMaxIdleConns(0)

	err = db.Ping()

	if err != nil {
		return err
	}

	db.SetMaxIdleConns(agent.ConnInfo.MaxIdleConns)

	if err != nil {
		return err
	}

	for _, q := range preQueries {
		_, err = db.Exec(q)

		if err != nil {
			return err
		}
	}

	agent.DB = db

	return nil
}

func (agent *Agent) Run(ctx context.Context, recorder *Recorder) error {
	ticker := time.NewTicker(AgentInterruptPeriod)
	defer ticker.Stop()
	responseTimes := []DataPoint{}

	_, err := agent.DB.Exec(fmt.Sprintf("SELECT 'agent(%d) start: token=%s'", agent.Id, agent.Token))

	if err != nil {
		return err
	}

	loopCount, err := agent.Data.EachLine(func(query string) (bool, error) {
		select {
		case <-ctx.Done():
			return false, nil
		case <-ticker.C:
			recorder.Add(responseTimes)
			responseTimes = responseTimes[:0]
		default:
			// nothing to do
		}

		rt, err := agent.Query(query)

		if err != nil {
			return false, err
		}

		tm := time.Now()

		agent.Logger.Log(query, rt, tm)

		responseTimes = append(responseTimes, DataPoint{
			Time:         tm,
			ResponseTime: rt,
		})

		return true, nil
	})

	if err != nil {
		return err
	}

	recorder.Add(responseTimes)
	atomic.StoreInt64(&recorder.LoopCount, loopCount)

	_, err = agent.DB.Exec(fmt.Sprintf("SELECT 'agent(%d) end: token=%s'", agent.Id, agent.Token))

	return err
}

func (agent *Agent) Query(query string) (time.Duration, error) {
	start := time.Now()
	_, err := agent.DB.Exec(query)
	end := time.Now()

	if err != nil {
		return 0, err
	}

	return end.Sub(start), nil
}

func (agent *Agent) Close() {
	agent.DB.Close()
}
