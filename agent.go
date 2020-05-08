package qrn

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

const AgentInterruptPeriod = 1 * time.Second

type ConnInfo struct {
	Driver       string
	DSN          string
	MaxIdleConns int
}

type Agent struct {
	ConnInfo *ConnInfo
	DB       *sql.DB
	Data     *Data
}

func (agent *Agent) Prepare() error {
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

	agent.DB = db

	return nil
}

func (agent *Agent) Run(ctx context.Context, recorder *Recorder) error {
	ticker := time.NewTicker(AgentInterruptPeriod)
	defer ticker.Stop()
	responseTimes := []time.Duration{}

	err := agent.Data.EachLine(func(query string) (bool, error) {
		select {
		case <-ctx.Done():
			return false, nil
		case <-ticker.C:
			recorder.Add(responseTimes)
			responseTimes = responseTimes[:0]
		default:
			rt, err := agent.Query(query)

			if err != nil {
				return false, err
			}

			responseTimes = append(responseTimes, rt)
		}

		return true, nil
	})

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
