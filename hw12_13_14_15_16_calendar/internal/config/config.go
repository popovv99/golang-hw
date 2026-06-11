package config

type Config struct {
	Logger  LoggerConf
	Storage StorageConf
	Server  ServerConf
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
