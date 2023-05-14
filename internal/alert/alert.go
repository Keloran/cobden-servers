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

func (a *Alert) SendAlert(name string, temp float64) error {
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
	text := fmt.Sprintf(`{"text": "%s %f", "rainbow": "true", "duration": 5}`, name, temp)
	token := client.Publish(a.Config.MQTT.Topic, 0, false, text)
	token.Wait()
	if token.Error() != nil {
		return token.Error()
	}
	client.Disconnect(250)

	return nil
}
