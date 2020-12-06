// Copyright 2018 Lars Hoogestraat
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package logger

import (
	"io"
	"strings"

	"github.com/sirupsen/logrus"
)

//Log returns a new logrus instance
var Log = logrus.New()

//InitLogger initializes the logger
//Valid log levels are: debug|info|warn|error|fatal|panic
//Fallback: info
func InitLogger(w io.Writer, level string) {
	level = strings.ToLower(level)

	switch level {
	case "debug":
		Log.Level = logrus.DebugLevel
	case "info":
		Log.Level = logrus.InfoLevel
	case "warn":
		Log.Level = logrus.WarnLevel
	case "error":
		Log.Level = logrus.ErrorLevel
	case "fatal":
		Log.Level = logrus.FatalLevel
	case "panic":
		Log.Level = logrus.PanicLevel
	default:
		Log.Warnf("could not read valid log level (%s); falling back to info level", level)
		Log.Level = logrus.InfoLevel
	}

	Log.Out = w
	Log.Formatter = &logrus.TextFormatter{FullTimestamp: true, DisableColors: true}
}
