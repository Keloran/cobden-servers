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

	go startService(s.Config, errChan)

	return <-errChan
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

	for {
		for _, server := range s {
			// skip first result
			if server.FirstResult {
				server.FirstResult = false
				continue
			}

			n, err := server.GetTemp()
			if err != nil {
				errChan <- logs.Errorf("get %s temp: %v", server.Name, err)
			}

			if n > server.LastTemp && n > (tempIncrease*server.LastTemp) {
				a := alert.NewAlert(context.Background(), *cfg)
				if err := a.SendAlert(server.Name, n); err != nil {
					errChan <- logs.Errorf("send alert: %v", err)
				}
			}

			logs.Infof("%s: oldTemp: %f, newTemp: %f\n", server.Name, server.LastTemp, n)
			server.LastTemp = n
		}

		iter = iter + 1
		logs.Infof("\n---- iter: %d ----\n", iter)

		time.Sleep(sleepTime)
	}
}
