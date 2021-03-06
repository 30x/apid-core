// Copyright 2017 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package logger

import (
	"fmt"
	"os"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/apid/apid-core"
)

const (
	ConfigLevel = "log_level"

	defaultLevel = logrus.ErrorLevel

	moduleField      = "module"
	environmentField = "env"
)

var std apid.LogService
var config apid.ConfigService
var textFormatter = &logrus.TextFormatter{
	FullTimestamp:   false,
	TimestampFormat: time.StampMilli,
}

func Base() apid.LogService {
	if std == nil {
		config = apid.Config()
		config.SetDefault(ConfigLevel, defaultLevel.String())
		logLevel := config.GetString(ConfigLevel)
		fmt.Printf("Base log level: %s\n", logLevel)
		std = NewLogger(ConfigLevel, logLevel)
	}
	return std
}

func ForModule(name string) apid.LogService {
	return Base().ForModule(name)
}

type logger struct {
	*logrus.Entry
}

// creates new logger for module w/ appropriate log level and field
// note: config module xx log level using config var: xx_log_level = "debug"
func (l *logger) ForModule(name string) apid.LogService {

	configKey := fmt.Sprintf("%s_%s", name, ConfigLevel)
	log := NewLogger(configKey, config.GetString(configKey)).WithField(moduleField, name)
	std.Debugf("created logger '%s' at level %s", name, log.(loggerPlus).Level())
	return log
}

func (l *logger) ForEnvironment(name string) apid.LogService {
	return l.WithField(environmentField, name)
}

func (l *logger) WithField(key string, value interface{}) apid.LogService {
	return &logger{l.Entry.WithField(key, value)}
}

func (l *logger) Level() logrus.Level {
	return l.Entry.Logger.Level
}

func NewLogger(configKey string, lvlString string) apid.LogService {

	var logLevel logrus.Level
	if std != nil {
		logLevel = std.(loggerPlus).Level()
	} else {
		logLevel = defaultLevel
	}

	if lvlString != "" {
		lvl, err := logrus.ParseLevel(lvlString)
		if err == nil {
			logLevel = lvl
		} else {
			std.Warnf("invalid log level '%s' in config key: '%s'", lvlString, configKey)
		}
	}

	log := &logger{
		logrus.NewEntry(
			&logrus.Logger{
				Out:       os.Stderr,
				Formatter: textFormatter,
				Level:     logLevel,
			},
		),
	}

	return log
}

type loggerPlus interface {
	Level() logrus.Level
}
