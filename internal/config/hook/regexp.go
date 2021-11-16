package hook

import (
	"reflect"
	"regexp"

	"github.com/mitchellh/mapstructure"
)

var (
	regexpType = reflect.TypeOf(&regexp.Regexp{})
)

func Regexp() mapstructure.DecodeHookFuncType {
	return func(in reflect.Type, out reflect.Type, val interface{}) (interface{}, error) {
		if in.Kind() == reflect.String && out == regexpType {
			return regexp.Compile(val.(string))
		}
		return val, nil
	}
}
