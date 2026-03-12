package config

import (
	"fmt"
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env string `yaml:"env"`
	//	DatabaseConfig            `yaml:"DB_INFO"`
	HTTPServerConfig          `yaml:"HTTP_SERVER_INFO"`
	AdditionalAddressesConfig `yaml:"ADDITIONAL_ADDRESSES"`
	//	JwtConfig                 `yaml:"JWT_INFO"`
	RabbitConfig    `yaml:"RABBIT_INFO"`
	KiznaiverConfig `yaml:"KIZNAIVER_INFO"`
	MinioConfig     `yaml:"MINIO_INFO"`
}

//type DatabaseConfig struct {
//	Host     string `yaml:"host"`
//	Name     string `yaml:"name"`
//	User     string `yaml:"user"`
//	Password string `yaml:"pass"`
//	SslModel string `yaml:"ssl_model"`
//}

type HTTPServerConfig struct {
	Port string `yaml:"port"`
	Host string `yaml:"host"`
}

type JwtConfig struct {
	Key string `yaml:"key"`
}

type RabbitConfig struct {
	AddressRabbitPath string `yaml:"address_path"`
	QueueName         string `yaml:"queue_name"`
}

type AdditionalAddressesConfig struct {
	Vue string `yaml:"vue"`
}

type MinioConfig struct {
	Endpoint        string `yaml:"endpoint"`
	AccessKeyID     string `yaml:"access_key"`
	SecretAccessKey string `yaml:"secret_key"`
	Bucket          string `yaml:"bucket"`
}

type KiznaiverConfig struct {
	Services map[string]string `yaml:"services"`
}

//func (cfg *Config) GetDataSourceName() string {
//	// dsn := "host=localhost user=user dbname=db password=password sslmode=disable"
//	return fmt.Sprintf(
//		"host=%s user=%s dbname=%s password=%s sslmode=%s",
//		cfg.DatabaseConfig.Host,
//		cfg.DatabaseConfig.User,
//		cfg.DatabaseConfig.Name,
//		cfg.DatabaseConfig.Password,
//		cfg.DatabaseConfig.SslModel,
//	)
//}

func (cfg *Config) GetAddress() string {
	return fmt.Sprintf("%s:%s", cfg.HTTPServerConfig.Host, cfg.HTTPServerConfig.Port)
}

func GetConfig(configPath string) *Config {
	envConfigPath := os.Getenv("CONFIG_PATH")
	if envConfigPath == "" {
		envConfigPath = configPath
		if envConfigPath == "" {
			log.Fatalf("CONFIG PATH is not set")
		}
	}

	if _, err := os.Stat(envConfigPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", envConfigPath)
	}

	var myconfig Config
	if err := cleanenv.ReadConfig(envConfigPath, &myconfig); err != nil {
		log.Fatalf("Can not read config file %s: %s", envConfigPath, err)
	}

	return &myconfig
}
