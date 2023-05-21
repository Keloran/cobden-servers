package temp

import (
	"context"
	"fmt"
	"strings"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/keloran/cobden-servers/internal/config"
)

type Temp struct {
	config.Config
	context.Context
}

type Server struct {
	config.Config
	context.Context

	Name           string
	FullName       string
	LastTemp       float64
	Sensor         string
	FirstResult    bool
	Chip           string
	LastReportTime time.Time
	Data           interface{}
}

func NewTempService(ctx context.Context, cfg config.Config) *Temp {
	return &Temp{
		Config:  cfg,
		Context: ctx,
	}
}

func (t *Temp) GetServers() ([]*Server, error) {
	r := []*Server{}

	client := influxdb2.NewClient(fmt.Sprintf("http://%s:%d", t.Config.Influx.Host, t.Config.Influx.Port), t.Config.Influx.Token)
	defer client.Close()
	api := client.QueryAPI(t.Config.Influx.Org)
	result, err := api.Query(t.Context, `sensors_query = from(bucket: "sensors")
  |> range(start: -30s)
  |> filter(fn: (r) => r["_measurement"] == "sensors")
  |> filter(fn: (r) => r["_field"] == "temp_input")
  |> toFloat()
  |> group(columns: ["host"])


ipmi_query = from(bucket: "sensors")
  |> range(start: -30s)
  |> filter(fn: (r) => r["_measurement"] == "ipmi_sensor")
  |> filter(fn: (r) => r["name"] == "temp_1")
  |> toFloat()
  |> group(columns: ["host"])

union(tables: [sensors_query, ipmi_query])
`)
	if err != nil {
		return nil, err
	}
	for result.Next() {
		shortName := result.Record().ValueByKey("host").(string)
		parts := strings.Split(shortName, ".")
		shortName = parts[0]

		s := &Server{
			Config:  t.Config,
			Context: t.Context,

			Name:        shortName,
			FullName:    result.Record().ValueByKey("host").(string),
			LastTemp:    result.Record().ValueByKey("_value").(float64),
			Sensor:      result.Record().ValueByKey("_measurement").(string),
			Chip:        result.Record().ValueByKey("chip").(string),
			FirstResult: true,
			Data:        result,
		}
		r = append(r, s)
	}

	return r, nil
}

func (s *Server) GetTemp(chip string) (float64, error) {
	q := fmt.Sprintf(`from(bucket: "sensors")
  |> range(start: -10m)
  |> filter(fn: (r) => r["_measurement"] == "sensors")
  |> filter(fn: (r) => r["_field"] == "temp_input")
  |> filter(fn: (r) => r["host"] == "%s")
  |> filter(fn: (r) => r["chip"] == "%s")
  |> toFloat()`, s.FullName, chip)

	if s.Sensor == "ipmi_sensor" {
		q = fmt.Sprintf(`from(bucket: "sensors")
  |> range(start: -10m)
  |> filter(fn: (r) => r["_measurement"] == "ipmi_sensor")
  |> filter(fn: (r) => r["name"] == "temp_1")
  |> filter(fn: (r) => r["host"] == "%s")
  |> toFloat()`, s.FullName)
	}

	client := influxdb2.NewClient(fmt.Sprintf("http://%s:%d", s.Config.Influx.Host, s.Config.Influx.Port), s.Config.Influx.Token)
	defer client.Close()
	api := client.QueryAPI(s.Config.Influx.Org)
	result, err := api.Query(s.Context, q)
	if err != nil {
		return 0, err
	}
	for result.Next() {
		return result.Record().ValueByKey("_value").(float64), nil
	}

	return 0, nil
}

func CleanServers(servers []*Server) []*Server {
	tempMap := make(map[string]*Server)
	for _, server := range servers {
		if existingServer, ok := tempMap[server.Name]; ok {
			if server.LastTemp > existingServer.LastTemp {
				tempMap[server.Name] = server
			}
		} else {
			tempMap[server.Name] = server
		}
	}

	finalServers := make([]*Server, 0, len(tempMap))
	for _, server := range tempMap {
		finalServers = append(finalServers, server)
	}

	return finalServers
}
