//go:build kafka

package forward

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/IBM/sarama"
)

type KafkaForwarder struct {
	producer sarama.SyncProducer
	topic    string
}

func NewKafkaForwarder(cfg KafkaConfig) (*KafkaForwarder, error) {
	if len(cfg.Brokers) == 0 {
		return nil, errors.New("kafka: no brokers configured")
	}
	if cfg.Topic == "" {
		return nil, errors.New("kafka: no topic configured")
	}

	saramaCfg := sarama.NewConfig()
	saramaCfg.Producer.RequiredAcks = sarama.WaitForLocal
	saramaCfg.Producer.Compression = sarama.CompressionSnappy
	saramaCfg.Producer.Return.Successes = true
	saramaCfg.Producer.Retry.Max = 3
	saramaCfg.Net.DialTimeout = 5 * time.Second
	saramaCfg.Net.ReadTimeout = 5 * time.Second
	saramaCfg.Net.WriteTimeout = 5 * time.Second

	producer, err := sarama.NewSyncProducer(cfg.Brokers, saramaCfg)
	if err != nil {
		return nil, err
	}

	return &KafkaForwarder{
		producer: producer,
		topic:    cfg.Topic,
	}, nil
}

func (k *KafkaForwarder) Forward(ctx context.Context, msg interface{}) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	_, _, err = k.producer.SendMessage(&sarama.ProducerMessage{
		Topic: k.topic,
		Key:   nil,
		Value: sarama.ByteEncoder(body),
	})
	return err
}

func (k *KafkaForwarder) Close() error {
	return k.producer.Close()
}
