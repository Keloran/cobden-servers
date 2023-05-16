package alert

import (
	"context"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/keloran/cobden-servers/internal/config"
)

type Alert struct {
	config.Config
	context.Context
}

func NewAlert(ctx context.Context, cfg config.Config) *Alert {
	return &Alert{
		Config:  cfg,
		Context: ctx,
	}
}

func (a *Alert) SendAlert(name string, newTemp, oldTemp float64, high bool) error {
	var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {}
	var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {}
	var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {}

	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", a.Config.MQTT.Host, a.Config.MQTT.Port))
	opts.SetClientID("cobden-servers")
	opts.SetUsername(a.Config.MQTT.Username)
	opts.SetPassword(a.Config.MQTT.Password)
	opts.SetDefaultPublishHandler(messagePubHandler)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	nt := newTemp - oldTemp
	color := "#00FFFF"
	if high {
		nt = oldTemp - newTemp
		color = "#880808"
	}

	token := client.Publish(a.Config.MQTT.Topic, 0, false, fmt.Sprintf(`{"text": "%s (%.2f) %.2f", "color": "%s", "duration": %d}`, name, newTemp, nt, color, a.Config.MQTT.AlertTime))
	token.Wait()
	if token.Error() != nil {
		return token.Error()
	}
	client.Disconnect(250)

	return nil
}
