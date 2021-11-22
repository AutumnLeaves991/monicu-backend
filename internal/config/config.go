package config

import (
	"regexp"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	"go.uber.org/zap/zapcore"
	"pkg.mon.icu/monicu/internal/config/hook"
	"pkg.mon.icu/monicu/internal/storage/model"
)

type Config struct {
	Discord struct {
		Auth     string
		Guilds   []model.Snowflake
		Channels []model.Snowflake
	}

	Posts struct {
		IgnoreRegexp *regexp.Regexp
	}

	Storage struct {
		PostgresDSN string
	}

	Logging struct {
		Level zapcore.Level
	}

	Api struct {
		Port uint16
	}
}

func Read() (*Config, error) {
	v := viper.New()
	configureEnv(v)
	configureLocation(v)
	return readUnmarshalConfig(v)
}

func configureEnv(v *viper.Viper) {
	v.AutomaticEnv()
	v.SetEnvPrefix("conf")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
}

func configureLocation(v *viper.Viper) {
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
}

func readUnmarshalConfig(v *viper.Viper) (*Config, error) {
	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}
	c := &Config{}
	if err := v.Unmarshal(c, viper.DecodeHook(mapstructure.ComposeDecodeHookFunc(
		hook.Regexp(), hook.Level(),
	))); err != nil {
		return nil, err
	}
	return c, nil
}
