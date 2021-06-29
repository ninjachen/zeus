package sql

import (
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap/zapcore"
	"xorm.io/core"

	"go.yym.plus/zeus/pkg/log"
	"go.yym.plus/zeus/pkg/utils/structs"

	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"

	"xorm.io/xorm"
	xlog "xorm.io/xorm/log"
	"xorm.io/xorm/names"
)

type EngineConfig struct {
	Type            string `default:"mysql" validate:"required"`
	Uri             string `validate:"required"`
	ConnMaxLifeTime time.Duration
	MaxIdleConns    int
	MaxOpenConns    int
	LogLevel        string       `default:"debug"`
	NameMapperType  string       `default:"snake"`
	NameMapper      names.Mapper `json:"-"`
	ShowSql         bool         `default:"true"`
}

type Session struct {
	*xorm.Session
}

func NewEngine(config *EngineConfig) (*xorm.Engine, error) {
	err := structs.SetDefaultsAndValidate(config)
	if err != nil {
		return nil, err
	}

	x, err := xorm.NewEngine(config.Type, config.Uri)
	if err != nil {
		return nil, errors.WithMessage(err, "create xorm engine error")
	}

	if config.ConnMaxLifeTime != 0 {
		x.DB().SetConnMaxLifetime(config.ConnMaxLifeTime)
	}

	if config.MaxIdleConns != 0 {
		x.DB().SetMaxIdleConns(config.MaxIdleConns)
	}

	if config.MaxOpenConns != 0 {
		x.DB().SetMaxOpenConns(config.MaxOpenConns)
	}

	x.SetLogger(&Logger{
		logger: log.Default().Clone(),
	})

	switch config.LogLevel {
	case "debug":
		x.SetLogLevel(xlog.LOG_DEBUG)
	case "info":
		x.SetLogLevel(xlog.LOG_INFO)
	case "warning":
		x.SetLogLevel(xlog.LOG_WARNING)
	case "error":
		x.SetLogLevel(xlog.LOG_ERR)
	case "off":
		x.SetLogLevel(xlog.LOG_OFF)
	}

	if config.NameMapper == nil {
		switch config.NameMapperType {
		case "snake":
			config.NameMapper = core.SnakeMapper{}
		case "lowerCamel":
			config.NameMapper = lowerCamelMapper{}
		default:
			log.Errorw("name mapper type invalid", "mapper", config.NameMapperType)
			config.NameMapper = core.SnakeMapper{}
		}
	}

	x.SetTableMapper(core.NewPrefixMapper(config.NameMapper, ""))
	x.SetColumnMapper(config.NameMapper)
	x.ShowSQL(config.ShowSql)

	return x, nil
}

type Logger struct {
	logger  *log.Logger
	level   xlog.LogLevel
	showSql bool
}

func (self *Logger) Debug(v ...interface{}) {
	self.logger.Debugv(v...)
}

func (self *Logger) Error(v ...interface{}) {
	self.logger.Errorv(v...)
}

func (self *Logger) Info(v ...interface{}) {
	self.logger.Infov(v...)
}

func (self *Logger) Warn(v ...interface{}) {
	self.logger.Warnv(v...)
}

func (self *Logger) Debugf(format string, v ...interface{}) {
	self.logger.Debug(format, v...)
}

func (self *Logger) Errorf(format string, v ...interface{}) {
	self.logger.Error(format, v...)
}

func (self *Logger) Infof(format string, v ...interface{}) {
	self.logger.Info(format, v...)
}

func (self *Logger) Warnf(format string, v ...interface{}) {
	self.logger.Warn(format, v...)
}

func (self *Logger) Level() xlog.LogLevel {
	return self.level
}

func (self *Logger) SetLevel(l xlog.LogLevel) {
	self.level = l
	zl := zapcore.DebugLevel
	switch l {
	case xlog.LOG_DEBUG:
		zl = zapcore.DebugLevel
	case xlog.LOG_INFO:
		zl = zapcore.InfoLevel
	default:
		zl = zapcore.ErrorLevel
	}
	self.logger.SetLevel(zl)
}

func (self *Logger) ShowSQL(show ...bool) {
	isShow := true
	if len(show) > 0 {
		isShow = show[0]
	}
	self.showSql = isShow
}

func (self *Logger) IsShowSQL() bool {
	return self.showSql
}
