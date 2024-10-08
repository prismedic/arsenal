package loggerfx

import (
	"fmt"
	"os"
	"path"

	"github.com/fatih/color"
	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/prismedic/scalpel/config"
	"github.com/prismedic/scalpel/logger"
)

var Module = fx.Options(
	fx.Provide(New),
	fx.WithLogger(func(logger *zap.SugaredLogger) fxevent.Logger {
		return &fxevent.ZapLogger{Logger: logger.Desugar()}
	}),
	fx.Decorate(RegisterLogLevelValidation),
)

func RegisterLogLevelValidation(validate *validator.Validate) (*validator.Validate, error) {
	if err := validate.RegisterValidation("loglevel", validateLogLevel); err != nil {
		return nil, err
	}
	return validate, nil
}

type LogLevel string

var (
	DebugLevel  = LogLevel(zapcore.DebugLevel.String())
	InfoLevel   = LogLevel(zapcore.InfoLevel.String())
	WarnLevel   = LogLevel(zapcore.WarnLevel.String())
	ErrorLevel  = LogLevel(zapcore.ErrorLevel.String())
	DPanicLevel = LogLevel(zapcore.DPanicLevel.String())
	PanicLevel  = LogLevel(zapcore.PanicLevel.String())
	FatalLevel  = LogLevel(zapcore.FatalLevel.String())
)

var logLevelMap = map[LogLevel]zapcore.Level{
	DebugLevel:  zapcore.DebugLevel,
	InfoLevel:   zapcore.InfoLevel,
	WarnLevel:   zapcore.WarnLevel,
	ErrorLevel:  zapcore.ErrorLevel,
	DPanicLevel: zapcore.DPanicLevel,
	PanicLevel:  zapcore.PanicLevel,
	FatalLevel:  zapcore.FatalLevel,
}

func validateLogLevel(fieldLevel validator.FieldLevel) bool {
	logLevel := fieldLevel.Field().String()
	_, ok := logLevelMap[LogLevel(logLevel)]
	return ok
}

type LoggerConfig struct {
	File struct {
		Level LogLevel `mapstructure:"level" yaml:"level" validate:"required,loglevel"`
		Path  string   `mapstructure:"path" yaml:"path" validate:"required"`
	} `mapstructure:"file" yaml:"file" validate:"required"`
	Console struct {
		Level LogLevel `mapstructure:"level" yaml:"level" validate:"required,loglevel"`
	} `mapstructure:"console" yaml:"console" validate:"required"`
}

func init() {
	// config must have a default value for viper to load config from env variables
	// default value of empty string (zero value) will not pass the "required" config validation
	viper.SetDefault("logs.file.path", path.Join("/var/log", config.GetPackageName()))
	viper.SetDefault("logs.file.level", InfoLevel)
	viper.SetDefault("logs.console.level", InfoLevel)
}

func New(config *LoggerConfig) (*zap.SugaredLogger, error) {
	// create directory if needed
	err := os.MkdirAll(config.File.Path, os.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("error in creating log file folder for writing: %w", err)
	}

	// create a new writer for log rotation
	fileWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename: path.Join(config.File.Path, "server.log"),
	})

	// setting the log level for file/console log output
	fileLogLevel := logLevelMap[config.File.Level]
	consoleLogLevel := logLevelMap[config.Console.Level]

	// setup the encoders
	fileEncoderConfig := zap.NewProductionEncoderConfig()
	fileEncoder := zapcore.NewJSONEncoder(fileEncoderConfig)
	consoleEncoderConfig := zap.NewProductionEncoderConfig()
	colorMap := map[zapcore.Level]*color.Color{
		zapcore.DebugLevel:  logger.DebugColor,
		zapcore.InfoLevel:   logger.InfoColor,
		zapcore.WarnLevel:   logger.WarnColor,
		zapcore.ErrorLevel:  logger.ErrorColor,
		zapcore.DPanicLevel: logger.FatalColor,
		zapcore.FatalLevel:  logger.FatalColor,
		zapcore.PanicLevel:  logger.FatalColor,
	}
	consoleEncoderConfig.EncodeLevel = func(l zapcore.Level, pae zapcore.PrimitiveArrayEncoder) {
		// custom encoding of level string as [INFO] style
		pae.AppendString(colorMap[l].Sprintf("[%s]", l.CapitalString()))
	}
	consoleEncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	consoleEncoderConfig.EncodeCaller = func(ec zapcore.EntryCaller, pae zapcore.PrimitiveArrayEncoder) {
		// custom encoding of the caller, now is set to the trimmed file path
		pae.AppendString(ec.TrimmedPath())
	}
	consoleEncoder := zapcore.NewConsoleEncoder(consoleEncoderConfig)

	// create the two cores for the logger
	// when writing to a file, the *os.File need to be locked with Lock() for concurrent access
	core := zapcore.NewTee(
		zapcore.NewCore(fileEncoder, fileWriter, fileLogLevel),
		zapcore.NewCore(consoleEncoder, zapcore.Lock(os.Stderr), consoleLogLevel),
	)

	return zap.New(core, zap.AddCaller()).Sugar(), nil
}
