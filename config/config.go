package config

import (
	"log"
	"time"

	"github.com/caarlos0/env"
	"github.com/joho/godotenv"
)

// Файл переменных окружения
type Enviroment struct {
	LoggerLevel   string `env:"loggerMode" envDefault:"debug"`
	TgToken       string `env:"TG_TOKEN,required"`
	Dashboard     int    `env:"DASHBOARD" envDefault:"60"`
	PaymentToken  string `env:"PAYMENT_TOKEN,required"`
	PaymentEX     int    `env:"PAYMENT_EX_TIME" envDefault:"3600"`
	PrivateChatID int64  `env:"PRIVATE_CHAT_ID,required"`
	DB            DB
	Redis         Redis
	Cache         Cache
}

type DB struct {
	DBUser     string `env:"DB_USER,required"`
	DBPassword string `env:"DB_PASSWORD,required"`
	DBName     string `env:"DB_NAME,required"`
	DBHost     string `env:"DB_HOST,required"`
}

type Redis struct {
	RedisAddr     string `env:"REDIS_ADDR,required"`
	RedisPort     string `env:"REDIS_PORT" envDefault:"6379"`
	RedisPassword string `env:"REDIS_PASSWORD,required"`
	RedisDBId     int    `env:"REDIS_DB_ID,required"`
}

type Cache struct {
	EXTime         time.Duration `json:"EXTime"`
	UpdateInterval string        `json:"updateInterval"`
}

var enviroment Enviroment

func NewEnv(envPath ...string) (*Enviroment, error) {
	err := godotenv.Load(envPath...)
	if err != nil {
		log.Fatalf("Файл .env не найден: %s", err)
	}

	err = env.Parse(&enviroment)
	if err != nil {
		return nil, err
	}
	err = env.Parse(&enviroment.Redis)
	if err != nil {
		return nil, err
	}
	err = env.Parse(&enviroment.DB)
	if err != nil {
		return nil, err
	}

	return &enviroment, nil
}

func GetEnv() *Enviroment {
	return &Enviroment{}
}
