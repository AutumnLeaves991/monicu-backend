package hook

import (
	"reflect"

	"github.com/mitchellh/mapstructure"
	"go.uber.org/zap/zapcore"
)

var (
	levelType = reflect.TypeOf(zapcore.InfoLevel)
)

func Level() mapstructure.DecodeHookFuncType {
	return func(in reflect.Type, out reflect.Type, val interface{}) (interface{}, error) {
		if in.Kind() == reflect.String && out == levelType {
			l := zapcore.InfoLevel
			if err := l.UnmarshalText([]byte(val.(string))); err != nil {
				return nil, err
			}
			return l, nil
		}
		return val, nil
	}
}
