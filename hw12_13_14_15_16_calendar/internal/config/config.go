package config

type Config struct {
	Logger    LoggerConf
	Storage   StorageConf
	Server    ServerConf
	Kafka     KafkaConf
	Scheduler SchedulerConf
}

type LoggerConf struct {
	Level string
}

type StorageConf struct {
	Type     string // "memory" or "sql"
	Database DatabaseConf
}

type DatabaseConf struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	Schema   string
}

type ServerConf struct {
	Host string
	Port string
}

type KafkaConf struct {
	Brokers []string `toml:"brokers"`
	Topic   string   `toml:"topic"`
	GroupID string   `toml:"group_id"`
	Retry   int      `toml:"retry"`
	Timeout string   `toml:"timeout"` // Go duration string
}

type SchedulerConf struct {
	Interval string `toml:"interval"` // Go duration string, e.g., "1m"
}
