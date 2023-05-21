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
	if s.Config.Development {
		startService(s.Config, errChan)
	} else {
		go startService(s.Config, errChan)
	}

	select {
	case err := <-errChan:
		if err != nil {
			errCount = errCount + 1

			if errCount > s.Config.Local.ErrorLimit {
				return logs.Errorf("error count: %d, err: %v", errCount, err)
			}
		}
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

	tempChangePercentage, err := strconv.ParseFloat(cfg.Local.TempChangePercentage, 64)
	if err != nil {
		errChan <- logs.Errorf("parse temp increase: %v", err)
	}

	for {
		for _, server := range s {
			// skip first result
			if server.FirstResult {
				server.FirstResult = false
				continue
			}
			oldTemp := server.LastTemp

			newTemp, err := server.GetTemp(server.Chip)
			if err != nil {
				errChan <- logs.Errorf("get %s temp: %v", server.Name, err)
			}

			changePercentage := (oldTemp - newTemp) / oldTemp * 100

			logs.Logf("%s: n %.2f, o %.2f, diff %.2f, c %.2f", server.Name, newTemp, oldTemp, newTemp-oldTemp, changePercentage)

			a := alert.NewAlert(context.Background(), *cfg)
			if newTemp > oldTemp {
				if changePercentage > tempChangePercentage {
					if err := a.SendAlert(server.Name, newTemp, oldTemp, true); err != nil {
						errChan <- logs.Errorf("send high alert: %v", err)
					}
				}
			} else if newTemp < oldTemp {
				if changePercentage > -tempChangePercentage {
					if err := a.SendAlert(server.Name, newTemp, oldTemp, false); err != nil {
						errChan <- logs.Errorf("send low alert: %v", err)
					}
				}
			}

			server.LastTemp = newTemp
		}

		iter = iter + 1
		logs.Infof("\n---- iter: %d ----\n", iter)

		time.Sleep(sleepTime)
	}
}
