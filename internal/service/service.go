package service

import (
	"context"
	"strconv"
	"time"

	"github.com/bugfixes/go-bugfixes/logs"
	"github.com/keloran/cobden-servers/internal/alert"
	"github.com/keloran/cobden-servers/internal/config"
	"github.com/keloran/cobden-servers/internal/temp"
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

	if err := <-errChan; err != nil {
		errCount++

		if errCount > s.Config.Local.ErrorLimit {
			return logs.Errorf("error count: %d, err: %v", errCount, err)
		}
	}

	return nil
}

func startService(cfg *config.Config, errChan chan error) {
	sleepTime := time.Duration(cfg.Local.SleepTime) * time.Second

	t := temp.NewTempService(context.Background(), *cfg)
	s, err := t.GetServers()
	if err != nil {
		errChan <- logs.Errorf("get servers: %v", err)
	}
	s = temp.CleanServers(s)

	tempChangePercentage, err := strconv.ParseFloat(cfg.Local.TempChangePercentage, 64)
	if err != nil {
		errChan <- logs.Errorf("parse temp increase: %v", err)
	}

	iterate(s, errChan, sleepTime, cfg, tempChangePercentage)
}

func iterate(s []*temp.Server, errChan chan error, sleepTime time.Duration, cfg *config.Config, tempChangePercentage float64) {
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
			if newTemp != oldTemp {
				if err := tempChange(newTemp, oldTemp, changePercentage, tempChangePercentage, server.Name, cfg); err != nil {
					errChan <- logs.Errorf("temp change: %v", err)
				}
			}

			server.LastTemp = newTemp
		}

		time.Sleep(sleepTime)
	}
}

func tempChange(newTemp, oldTemp, changePercentage, tempChangePercentage float64, serverName string, cfg *config.Config) error {
	a := alert.NewAlert(context.Background(), *cfg)
	if changePercentage < 0 {
		if changePercentage < (tempChangePercentage * -1) {
			logs.Logf("%s: got cooler by %0.2f, %0.2f%%", serverName, newTemp-oldTemp, changePercentage)
			if err := a.SendAlert(serverName, newTemp, oldTemp, false); err != nil {
				return logs.Errorf("send low alert: %v", err)
			}
		}
	} else {
		if changePercentage > tempChangePercentage {
			logs.Logf("%s: got warmer by %0.2f, %0.2f%%", serverName, oldTemp-newTemp, changePercentage)
			if err := a.SendAlert(serverName, newTemp, oldTemp, true); err != nil {
				return logs.Errorf("send high alert: %v", err)
			}
		}
	}

	return nil
}
