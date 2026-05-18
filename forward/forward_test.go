package forward

import (
	"testing"
)

func TestKafkaConfig(t *testing.T) {
	cfg := KafkaConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   "gb32960",
	}

	if len(cfg.Brokers) == 0 {
		t.Error("expected brokers to be set")
	}
	if cfg.Topic == "" {
		t.Error("expected topic to be set")
	}
}

func TestMQTTConfig(t *testing.T) {
	cfg := MQTTConfig{
		Broker:   "tcp://localhost:1883",
		ClientID: "gb32960-server",
		Topic:    "gb32960/data",
		Username: "admin",
		Password: "secret",
	}

	if cfg.Broker == "" {
		t.Error("expected broker to be set")
	}
	if cfg.ClientID == "" {
		t.Error("expected client ID to be set")
	}
	if cfg.Topic == "" {
		t.Error("expected topic to be set")
	}
}

func TestKafkaConfigDefaults(t *testing.T) {
	cfg := KafkaConfig{}
	if cfg.Topic != "" {
		t.Error("expected empty topic by default")
	}
}

func TestMQTTConfigDefaults(t *testing.T) {
	cfg := MQTTConfig{}
	if cfg.Username != "" {
		t.Error("expected empty username by default")
	}
	if cfg.Password != "" {
		t.Error("expected empty password by default")
	}
}
