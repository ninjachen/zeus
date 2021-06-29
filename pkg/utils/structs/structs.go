package structs

import (
	"github.com/creasty/defaults"
	validator "github.com/go-playground/validator/v10"
	"github.com/mitchellh/mapstructure"
)

var (
	defaultValidator = validator.New()
)

func SetDefaultsAndValidate(i interface{}) error {
	if err := defaults.Set(i); err != nil {
		return err
	}
	if err := defaultValidator.Struct(i); err != nil {
		return err
	}
	return nil
}

func SetDefaults(i interface{}) error {
	if err := defaults.Set(i); err != nil {
		return err
	}
	return nil
}

func Validate(i interface{}) error {
	if err := defaultValidator.Struct(i); err != nil {
		return err
	}
	return nil
}

func ToMap(i interface{}) (map[string]interface{}, error) {
	m := map[string]interface{}{}
	return m, mapstructure.Decode(i, m)
}

func Decode(i interface{}, o interface{}) error {
	return mapstructure.Decode(i, o)
}
