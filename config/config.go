package config

import (
	"io/ioutil"
	"strings"
	"time"

	"cloud.google.com/go/profiler"
	"github.com/pyroscope-io/client/pyroscope"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

// Config - service configuration
type Config struct {
	LogLevel zerolog.Level  `yaml:"LogLevel" mapstructure:"LogLevel"`
	GRPC     GRPCConfig     `yaml:"GRPC" mapstructure:"GRPC"`
	AWS      AWSConfig      `yaml:"AWS" mapstructure:"AWS"`
	DB       DBConfig       `yaml:"DB" mapstructure:"DB"`
	Metrics  MetricsConfig  `yaml:"DB" mapstructure:"Metrics"`
	Profiler ProfilerConfig `yaml:"Profiler" mapstructure:"Profiler"`
}

type GRPCConfig struct {
	Host       string `yaml:"Host" mapstructure:"Host"`
	Port       int    `yaml:"Port" mapstructure:"Port"`
	Reflection bool   `yaml:"reflection" mapstructure:"Reflection"`
}

type AWSConfig struct {
	Region string `yaml:"Region" mapstructure:"Region"`
	Id     string `yaml:"Id" mapstructure:"Id"`
	Secret string `yaml:"Secret" mapstructure:"Secret"`
	Token  string `yaml:"Token" mapstructure:"Token"`
}

type DBConfig struct {
	Host string `yaml:"Host" mapstructure:"Host"`
	Port int    `yaml:"Port" mapstructure:"Port"`
}

type MetricsConfig struct {
	Host            string        `yaml:"Host" mapstructure:"Host"`
	Port            int           `yaml:"Port" mapstructure:"Port"`
	ShutdownTimeout time.Duration `yaml:"ShutdownTimeout" mapstructure:"ShutdownTimeout"`
}

type ProfilerConfig struct {
	Enabled bool `yaml:"Enabled" mapstructure:"Enabled"`
	// Google Go Profiler, see ../NOTES.md#google-go-profiler
	GoProfiler profiler.Config `yaml:"GoProfiler" mapstructure:"GoProfiler"`
	// Pyroscope, see ../NOTES.md#pyroscope
	Pyroscope pyroscope.Config `yaml:"Pyroscope" mapstructure:"Pyroscope"`
}

// Service - this service configuration
var Service = Config{}

// Load loads service'sconfiguration from config.yml
func (c *Config) Load(configFile string) error {
	viper.AddConfigPath(".")
	viper.AddConfigPath("cmd")
	viper.SetConfigFile(configFile)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	viper.SetEnvPrefix("GTT")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Default port
	viper.SetDefault("GRPC.Port", 3000)

	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	if err := viper.Unmarshal(&Service); err != nil {
		return err
	}

	if err := c.loadSecrets(); err != nil {
		return err
	}

	return nil
}

func (c *Config) loadSecrets() error {
	secretPath := "/run/secrets/"

	if strings.Contains(c.AWS.Secret, secretPath) {
		secret, err := ioutil.ReadFile(c.AWS.Secret)
		if err != nil {
			return err
		}
		c.AWS.Secret = string(secret)
	}

	if strings.Contains(c.AWS.Token, secretPath) {
		token, err := ioutil.ReadFile(c.AWS.Token)
		if err != nil {
			return err
		}
		c.AWS.Token = string(token)
	}

	return nil
}
