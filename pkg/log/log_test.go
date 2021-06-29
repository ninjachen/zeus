package log_test

import (
	"errors"
	"testing"

	"go.yym.plus/zeus/pkg/log"
)

func TestLog(t *testing.T) {
	config := log.Config{
	}
	config.File.Paths = map[string]string{}
	config.File.Paths["info"] = "./log/info.log"

	log.Init(&config)
	log.WithError(errors.New("123")).Error("123")
	log.Infow("i am a log", "withStr", "data", "withInt", 1)
	log.Debugw("i am a log")
	log.Warnw("i am a log")
	log.Errorw("i am a log")
	log.Fatalw("i am a log")
}
