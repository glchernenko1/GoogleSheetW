package logger

import (
	"GoogleSheetW/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"strings"
	"sync"
)

var (
	once   sync.Once
	logger *zap.SugaredLogger
)

func parseLogLevel(level string) zapcore.Level {
	switch strings.ToLower(level) {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "apperrors":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel // По умолчанию Info
	}
}

func Get() *zap.SugaredLogger {
	once.Do(func() {

		if err := os.MkdirAll("log", os.ModePerm); err != nil {
			panic(err)
		}

		file, err := os.OpenFile("log/log.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			panic(err)
		}

		encoderConfig := zapcore.EncoderConfig{
			TimeKey:        "time",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			FunctionKey:    zapcore.OmitKey,
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,                              // Цветное отображение уровней
			EncodeTime:     zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000 UTC"), // Формат времени
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		}

		conf := config.GetConfig()
		atomicLevel := zap.NewAtomicLevelAt(parseLogLevel(conf.App.LogLevel))
		// Создаем core для записи в файл
		core := zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig), // Формат JSON
			zapcore.AddSync(file),                 // Запись в файл
			atomicLevel,                           // Уровень логирования
		)

		_log := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
		logger = _log.Sugar()
	})
	return logger

}
