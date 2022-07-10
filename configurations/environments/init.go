package environments

import (
	"github.com/caarlos0/env"
	"github.com/joho/godotenv"
	"rmrf-slash.com/go-srbackend/configurations/logger"
)

type Configuration struct {
	AllowOrigins    string `env:"ALLOW_ORIGINS"`
	MongoUsername   string `env:"MONGO_USERNAME"`
	MongoPassword   string `env:"MONGO_PASSWORD"`
	MongoConnection string `env:"MONGO_CONN"`
}

func GetVariables() Configuration {
	// load env vars
	godotenv.Load()
	// bind struct
	cfg := Configuration{}
	if err := env.Parse(&cfg); err != nil {
		logger.GetInstance().Println("Failed to parse env vars")
	}
	return cfg
}
