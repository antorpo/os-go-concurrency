package entities

const AppConfigName = "app"

type Configuration struct {
	App AppConfig
}

type AppConfig struct {
	Workers int `json:"workers"`
}
