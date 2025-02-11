// Logs

package main

import (
	"github.com/AgustinSRG/glog"
	"github.com/pion/logging"
)

type LoggerWrapperFactory struct {
	logger *glog.Logger
}

func (f *LoggerWrapperFactory) NewLogger(scope string) logging.LeveledLogger {
	return &LoggerWrapper{
		logger: f.logger.CreateChildLogger("[" + scope + "] "),
	}
}

type LoggerWrapper struct {
	logger *glog.Logger
}

func (w *LoggerWrapper) Trace(msg string) {
	w.logger.Trace(msg)
}

func (w *LoggerWrapper) Tracef(format string, args ...interface{}) {
	w.logger.Tracef(format, args...)
}

func (w *LoggerWrapper) Debug(msg string) {
	w.logger.Debug(msg)
}

func (w *LoggerWrapper) Debugf(format string, args ...interface{}) {
	w.logger.Debugf(format, args...)
}

func (w *LoggerWrapper) Info(msg string) {
	w.logger.Info(msg)
}

func (w *LoggerWrapper) Infof(format string, args ...interface{}) {
	w.logger.Infof(format, args...)
}

func (w *LoggerWrapper) Warn(msg string) {
	w.logger.Warning(msg)
}

func (w *LoggerWrapper) Warnf(format string, args ...interface{}) {
	w.logger.Warningf(format, args...)
}

func (w *LoggerWrapper) Error(msg string) {
	w.logger.Error(msg)
}

func (w *LoggerWrapper) Errorf(format string, args ...interface{}) {
	w.logger.Errorf(format, args...)
}
