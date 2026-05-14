package forward

type KafkaConfig struct {
	Brokers []string
	Topic   string
}

type MQTTConfig struct {
	Broker   string
	ClientID string
	Topic    string
	Username string
	Password string
}
