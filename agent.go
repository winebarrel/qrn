package qrn

import (
	"context"
	"database/sql"
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
	ConnInfo *ConnInfo
	DB       *sql.DB
	Data     *Data
	Logger   *Logger
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
	responseTimes := []DataPoint{}

	err := agent.Data.EachLine(func(query string, params []string) (bool, error) {
		select {
		case <-ctx.Done():
			return false, nil
		case <-ticker.C:
			recorder.Add(responseTimes)
			responseTimes = responseTimes[:0]
		default:
			// nothing to do
		}

		rt, err := agent.Query(query, params)

		if err != nil {
			return false, err
		}

		agent.Logger.Log(query, params, rt)

		responseTimes = append(responseTimes, DataPoint{
			Time:         time.Now(),
			ResponseTime: rt,
		})

		return true, nil
	})

	recorder.Add(responseTimes)
	return err
}

func (agent *Agent) Query(query string, params []string) (time.Duration, error) {
	ifParams := make([]interface{}, len(params))

	for i, p := range params {
		ifParams[i] = p
	}

	start := time.Now()
	_, err := agent.DB.Exec(query, ifParams...)
	end := time.Now()

	if err != nil {
		return 0, err
	}

	return end.Sub(start), nil
}

func (agent *Agent) Close() {
	agent.DB.Close()
}
