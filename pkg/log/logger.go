package log

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/xhit/go-str2duration"

	"go.yym.plus/zeus/pkg/utils/structs"

	"github.com/pkg/errors"
	"go.uber.org/zap/buffer"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Duration time.Duration

type Logger struct {
	logger *zap.SugaredLogger
	config *Config
	pool   buffer.Pool
}

type Config struct {
	Level  string `default:"debug"`
	level  zapcore.Level
	Rotate struct {
		Suffixes string `default:"_%Y-%m-%d-%H"`
		Time     string `default:"24h"`
	}
	Console struct {
		Enable *bool `default:"true"`
	}
	File struct {
		Enable *bool `default:"true"`
		//ErrorPath string `default:"./log/error.log"`
		//Path      string `default:"./log/log.log"`
		Paths  map[string]string
		MaxAge Duration `default:"1m"`
	}
	Format struct {
		Time         string `default:"2006-01-02 15:04:05"`
		Caller       *bool  `default:"true"`
		FileInfo     *bool  `default:"true"`
		FunctionName *bool  `default:"true"`
		Colorable    *bool  `default:"true"`
	}
}

// New returns a Logger instance.
func New(config *Config) *Logger {

	err := structs.SetDefaults(config)
	if err != nil {
		panic(errors.WithMessage(err, "set log default config error"))
	}
	err = config.level.UnmarshalText([]byte(config.Level))
	if err != nil {
		panic(errors.WithMessage(err, "log level invalid"))
	}
	l := &Logger{
		config: config,
		pool:   buffer.NewPool(),
	}
	var cores []zapcore.Core
	if *config.Console.Enable {
		cores = append(cores, l.consoleCore())
	}

	if *config.File.Enable {
		cores = append(cores, l.fileCore())
	}
	core := zapcore.RegisterHooks(zapcore.NewTee(cores...), func(e zapcore.Entry) error {
		//	return metricsLogCount.Add(1, map[string]string{"level": e.Level.String()})
		return nil
	})
	zapOptions := []zap.Option{}
	if *config.Format.Caller {
		zapOptions = append(zapOptions, zap.AddCaller(), zap.AddCallerSkip(2))
	}

	zapLogger := zap.New(core, zapOptions...)

	l.logger = zapLogger.Sugar()
	l.logger.With()
	return l
}

func (self *Logger) Clone() *Logger {
	return New(self.config)
}

func (self *Logger) consoleCore() zapcore.Core {
	conf := self.encoderConfig()
	if *self.config.Format.Colorable {
		conf.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	logEnabler := zap.LevelEnablerFunc(func(lev zapcore.Level) bool {
		return lev >= self.config.level
	})
	return zapcore.NewCore(zapcore.NewConsoleEncoder(conf), zapcore.Lock(os.Stderr), logEnabler)
}

func (self *Logger) Config() Config {
	return *self.config
}

func (self *Logger) SetLevel(l zapcore.Level) {
	self.config.level = l
}

func (self *Logger) encoderConfig() zapcore.EncoderConfig {
	conf := zap.NewProductionEncoderConfig()
	formatConfig := self.config.Format
	conf.TimeKey = "time"
	conf.EncodeTime = func(t time.Time, encoder zapcore.PrimitiveArrayEncoder) {
		encoder.AppendString(t.Format(formatConfig.Time))
	}

	callEncoder := zapcore.CallerEncoder(func(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
		buffer := self.pool.Get()
		if *formatConfig.FunctionName {
			fn := runtime.FuncForPC(caller.PC)
			if fn != nil {
				buffer.AppendString(" at ")
				buffer.AppendString(fn.Name())
				buffer.AppendString("()")
				if *formatConfig.FileInfo {
					buffer.AppendString(" ")
				}
			}
		}
		if *formatConfig.FileInfo {
			buffer.AppendString(filepath.Base(caller.FullPath()))
		}

		enc.AppendString(buffer.String())
		buffer.Free()
	})
	conf.EncodeCaller = callEncoder
	return conf
}

func (self *Logger) fileCore() zapcore.Core {
	if self.config.File.Paths == nil {
		self.config.File.Paths = map[string]string{}
	}

	if self.config.File.Paths["error"] == "" {
		self.config.File.Paths["error"] = "./log/error.log"
	}

	coreList := []zapcore.Core{}
	writers := map[string]zapcore.WriteSyncer{}
	var err error
	conf := self.encoderConfig()
	encoder := zapcore.NewJSONEncoder(conf)
	for levelName, filePath := range self.config.File.Paths {
		writer := writers[filePath]
		if writer == nil {

			writer, err = self.getWriter(filePath)
			if err != nil {
				panic(err)
			}
		}
		level := zapcore.DebugLevel
		err = level.UnmarshalText([]byte(levelName))
		if err != nil {
			panic(errors.WithMessage(err, "log file path config error"))
		}
		logEnabler := zap.LevelEnablerFunc(func(lev zapcore.Level) bool {
			return lev == level
		})
		logCore := zapcore.NewCore(encoder, writer, logEnabler)
		coreList = append(coreList, logCore)
	}
	defaultPath := self.config.File.Paths["default"]
	if defaultPath == "" {
		defaultPath = "./log/log.log"
	}
	defaultWriter := writers[defaultPath]
	if defaultWriter == nil {
		defaultWriter, err = self.getWriter(defaultPath)
		if err != nil {
			panic(errors.WithMessage(err, "create log writer error"))
		}
	}

	logEnabler := zap.LevelEnablerFunc(func(lev zapcore.Level) bool {
		return lev >= self.config.level
	})

	defaultCore := zapcore.NewCore(encoder, defaultWriter, logEnabler)
	coreList = append(coreList, defaultCore)
	return zapcore.NewTee(coreList...)
}

func (self *Logger) getWriter(path string) (zapcore.WriteSyncer, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}
	rotationTime, err := time.ParseDuration(self.config.Rotate.Time)
	if err != nil {
		return nil, err
	}
	extName := filepath.Ext(path)
	path = strings.TrimSuffix(path, extName) + self.config.Rotate.Suffixes + extName
	log, err := rotatelogs.New(
		path,
		//	rotatelogs.WithLinkName(path),
		rotatelogs.WithRotationTime(rotationTime),
		rotatelogs.WithMaxAge(self.config.File.MaxAge.Value()),
	)

	if err != nil {
		panic(err)
	}
	return zapcore.Lock(zapcore.AddSync(log)), nil
}

func (self *Logger) With(args ...interface{}) *Logger {
	l := Logger{
		logger: self.logger.With(args...),
		config: self.config,
		pool:   self.pool,
	}
	return &l
}

func (self *Logger) WithError(err error) *Logger {
	l := Logger{
		logger: self.logger.With(zap.Error(err)),
		config: self.config,
		pool:   self.pool,
	}
	return &l
}

func (self *Logger) WithOptions(opts ...zap.Option) *zap.Logger {
	return self.logger.Desugar().WithOptions(opts...)
}

func (self *Logger) Debug(format string, args ...interface{}) {
	self.logger.Debugf(format, args...)
}

func (self *Logger) Info(format string, args ...interface{}) {
	self.logger.Infof(format, args...)
}

func (self *Logger) Warn(format string, args ...interface{}) {
	self.logger.Warnf(format, args...)
}

func (self *Logger) Error(format string, args ...interface{}) {
	self.logger.Errorf(format, args...)
}

func (self *Logger) Fatal(format string, args ...interface{}) {
	self.logger.Fatalf(format, args...)
}

func (self *Logger) Panic(format string, args ...interface{}) {
	self.logger.Panicf(format, args...)
}

func (self *Logger) Debugw(msg string, kvs ...interface{}) {
	self.logger.Debugw(msg, kvs...)
}

func (self *Logger) Infow(msg string, kvs ...interface{}) {
	self.logger.Infow(msg, kvs...)
}

func (self *Logger) Warnw(msg string, kvs ...interface{}) {
	self.logger.Warnw(msg, kvs...)
}

func (self *Logger) Errorw(msg string, kvs ...interface{}) {
	self.logger.Errorw(msg, kvs...)
}

func (self *Logger) Fatalw(msg string, kvs ...interface{}) {
	self.logger.Fatalw(msg, kvs...)
}

func (self *Logger) Panicw(msg string, kvs ...interface{}) {
	self.logger.Panicw(msg, kvs...)
}

func (self *Logger) Debugv(args ...interface{}) {
	self.logger.Debug(args...)
}

func (self *Logger) Infov(args ...interface{}) {
	self.logger.Info(args...)
}

func (self *Logger) Warnv(args ...interface{}) {
	self.logger.Warn(args...)
}

func (self *Logger) Errorv(args ...interface{}) {
	self.logger.Error(args...)
}

func (self *Logger) Fatalv(args ...interface{}) {
	self.logger.Fatal(args...)
}

func (self *Logger) Panicv(args ...interface{}) {
	self.logger.Panic(args...)
}

func (self *Duration) UnmarshalJSON(bytes []byte) error {
	d, err := str2duration.Str2Duration(string(bytes))
	if err != nil {
		return err
	}
	*self = Duration(d)
	return nil
}

func (self *Duration) Value() time.Duration {
	return time.Duration(*self)
}
