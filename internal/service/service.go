package service

import (
	"context"
	"github.com/bugfixes/go-bugfixes/logs"
	"github.com/keloran/cobden-servers/internal/alert"
	"github.com/keloran/cobden-servers/internal/config"
	"github.com/keloran/cobden-servers/internal/temp"
	"strconv"
	"time"
)

type Service struct {
	Config *config.Config
	context.Context
}

func (s *Service) Start() error {
	errChan := make(chan error)

	errCount := 0
	go startService(s.Config, errChan)

	errOrig := <-errChan
	if errOrig != nil {
		errCount = errCount + 1
	}
	if errCount > s.Config.Local.ErrorLimit {
		return logs.Errorf("error count: %d, err: %v", errCount, errOrig)
	}

	return nil
}

func startService(cfg *config.Config, errChan chan error) {
	var sleepTime time.Duration
	sleepTime = time.Duration(cfg.Local.SleepTime) * time.Second

	t := temp.NewTempService(context.Background(), *cfg)
	s, err := t.GetServers()
	if err != nil {
		errChan <- logs.Errorf("get servers: %v", err)
	}
	s = temp.CleanServers(s)
	iter := 0

	tempIncrease, err := strconv.ParseFloat(cfg.Local.TempIncrease, 64)
	if err != nil {
		errChan <- logs.Errorf("parse temp increase: %v", err)
	}

	if cfg.Development {
		a := alert.NewAlert(context.Background(), *cfg)
		if err := a.SendAlert("test", 99, 89, true); err != nil {
			errChan <- logs.Errorf("send low alert: %v", err)
		}
		return
	}

	for {
		for _, server := range s {
			// skip first result
			if server.FirstResult {
				server.FirstResult = false
				continue
			}

			newTemp, err := server.GetTemp()
			if err != nil {
				errChan <- logs.Errorf("get %s temp: %v", server.Name, err)
			}

			logs.Logf("%s: n %.2f, o %.2f", server.Name, newTemp, server.LastTemp)

			if newTemp > server.LastTemp && newTemp > (tempIncrease*server.LastTemp) {
				a := alert.NewAlert(context.Background(), *cfg)
				if err := a.SendAlert(server.Name, newTemp, server.LastTemp, true); err != nil {
					errChan <- logs.Errorf("send high alert: %v", err)
				}
			}

			if newTemp < server.LastTemp && newTemp < (tempIncrease*server.LastTemp) {
				a := alert.NewAlert(context.Background(), *cfg)
				if err := a.SendAlert(server.Name, newTemp, server.LastTemp, false); err != nil {
					errChan <- logs.Errorf("send low alert: %v", err)
				}
			}

			server.LastTemp = newTemp
		}

		iter = iter + 1
		logs.Infof("\n---- iter: %d ----\n", iter)

		time.Sleep(sleepTime)
	}
}
