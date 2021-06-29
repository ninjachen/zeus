package conf

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type Source struct {
	Type        string
	ContentType string
	URI         string
}

func NewViperFromFile(p string, dirs ...string) (*viper.Viper, error) {
	v := viper.New()
	for _, dir := range dirs {
		v.AddConfigPath(dir)
	}
	v.SetConfigName(p)
	err := v.ReadInConfig()
	if err != nil {
		return nil, err
	}
	source := Source{}
	err = v.Unmarshal(&source)
	if err != nil {
		return nil, errors.WithMessage(err, "parse source file failed")
	}
	return NewViperFromSource(&source)
}

func NewViperFromSource(s *Source) (*viper.Viper, error) {
	v := viper.New()
	v.SetConfigType(s.ContentType)
	switch s.Type {
	case "file":
		v.SetConfigFile(s.URI)
		err := v.ReadInConfig()
		if err != nil {
			return nil, err
		}
	case "nacos":
		MakeVipperSupportNacos()
		v.AddRemoteProvider("nacos", s.URI, "")
		err := v.ReadRemoteConfig()
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("config source type invalid")
	}
	return v, nil
}
