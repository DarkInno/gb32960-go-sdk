//go:build mqtt

package forward

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type MQTTForwarder struct {
	client mqtt.Client
	topic  string
}

func NewMQTTForwarder(cfg MQTTConfig) (*MQTTForwarder, error) {
	if cfg.Broker == "" {
		return nil, errors.New("mqtt: no broker configured")
	}
	if cfg.Topic == "" {
		return nil, errors.New("mqtt: no topic configured")
	}

	opts := mqtt.NewClientOptions()
	opts.AddBroker(cfg.Broker)
	opts.SetClientID(cfg.ClientID)
	if cfg.Username != "" {
		opts.SetUsername(cfg.Username)
	}
	if cfg.Password != "" {
		opts.SetPassword(cfg.Password)
	}
	opts.SetConnectTimeout(10 * time.Second)
	opts.SetAutoReconnect(true)
	opts.SetMaxReconnectInterval(30 * time.Second)

	client := mqtt.NewClient(opts)
	token := client.Connect()
	if !token.WaitTimeout(15 * time.Second) {
		return nil, fmt.Errorf("mqtt: connect timeout")
	}
	if err := token.Error(); err != nil {
		return nil, err
	}

	return &MQTTForwarder{
		client: client,
		topic:  cfg.Topic,
	}, nil
}

func (m *MQTTForwarder) Forward(ctx context.Context, msg interface{}) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	token := m.client.Publish(m.topic, 1, false, body)
	if !token.WaitTimeout(5 * time.Second) {
		return fmt.Errorf("mqtt: publish timeout")
	}
	return token.Error()
}

func (m *MQTTForwarder) Close() error {
	m.client.Disconnect(500)
	return nil
}
