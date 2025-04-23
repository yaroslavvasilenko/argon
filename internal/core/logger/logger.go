package logger

import (
	"github.com/yaroslavvasilenko/argon/config"
	"os"
	"strings"

	"github.com/phuslu/log"
)

type Glog struct {
	log.Logger
}

func (l *Glog) Infof(fmt string, a ...any) { l.Logger.Info().Msgf(fmt, a...) }

func (l *Glog) Warnf(fmt string, a ...any) { l.Logger.Warn().Msgf(fmt, a...) }

func (l *Glog) Errorf(fmt string, a ...any) { l.Logger.Error().Msgf(fmt, a...) }

func NewLogger(cfg config.Config) *Glog {
	DefaultLogger := log.Logger{
		TimeFormat: "15:04:05",
		Caller:     1,
		Writer:     &log.IOWriter{Writer: os.Stderr},
		Level:      parseLogLevel(cfg.Logger.Level),
	}

	if log.IsTerminal(os.Stderr.Fd()) {
		DefaultLogger = log.Logger{
			TimeFormat: "15:04:05",
			Caller:     1,
			Writer: &log.ConsoleWriter{
				ColorOutput:    true,
				QuoteString:    true,
				EndWithMessage: true,
			},
		}
	}

	log.DefaultLogger = DefaultLogger
	logInstance := &Glog{
		Logger: DefaultLogger,
	}

	return logInstance
}

func parseLogLevel(level string) log.Level {
	switch strings.ToLower(level) {
	case "trace":
		return log.TraceLevel
	case "debug":
		return log.DebugLevel
	case "info":
		return log.InfoLevel
	case "warn":
		return log.WarnLevel
	case "error":
		return log.ErrorLevel
	case "fatal":
		return log.FatalLevel
	default:
		return log.InfoLevel // Уровень по умолчанию.
	}
}
