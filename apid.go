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

package apid

import (
	"errors"
	"os"
	"time"
)

const (
	SystemEventsSelector  EventSelector = "system event"
	ShutdownEventSelector EventSelector = "shutdown event"
	ShutdownTimeout       time.Duration = 10 * time.Second
	configfwdProxyURL	=   "configfwdproxy_url"
	configfwdProxyProt	=   "configfwdproxy_prot"
	configfwdProxyUser	=   "configfwdproxy_user"
	configfwdProxyPasswd	=   "configfwdproxy_passwd"
	configfwdProxyPort      =   "configfwdproxy_port"
	configfwdProxyPortURL   =   "configcompletefwdp"
)

var (
	APIDInitializedEvent = systemEvent{"apid initialized"}
	APIListeningEvent    = systemEvent{"api listening"}
	PluginVersionTracker []PluginData

	pluginInitFuncs []PluginInitFunc
	services        Services
)

type Services interface {
	API() APIService
	Config() ConfigService
	Data() DataService
	Events() EventsService
	Log() LogService
}

func setFwdProxyConfig(config ConfigService) {
	var pURL string

	config.SetDefault(configfwdProxyProt, "https")
	fwdPrxy := config.GetString(configfwdProxyURL)
	fwdprxyProt := config.GetString(configfwdProxyProt)
	fwdPrxyUser := config.GetString(configfwdProxyUser)
	fwdPrxyPass := config.GetString(configfwdProxyPasswd)
	fwdPrxyPort := config.GetString(configfwdProxyPort)
	if fwdPrxy != "" && fwdPrxyUser != "" && fwdPrxyPort != "" {
		pURL = fwdprxyProt + "://" + fwdPrxyUser + ":" + fwdPrxyPass + "@" + fwdPrxy + ":" + fwdPrxyPort
	} else if fwdPrxy != "" && fwdPrxyPort != "" {
		pURL = fwdprxyProt + "://" + fwdPrxy + ":" + fwdPrxyPort
	}
	config.Set(configfwdProxyPortURL, pURL)
}

type PluginInitFunc func(Services) (PluginData, error)

// passed Services can be a factory - makes copies and maintains returned references
// eg. apid.Initialize(factory.DefaultServicesFactory())

func Initialize(s Services) {
	ss := &servicesSet{}
	services = ss
	// order is important
	ss.config = s.Config()
	ss.log = s.Log()

	// ensure storage path exists
	lsp := ss.config.GetString("local_storage_path")
	if err := os.MkdirAll(lsp, 0700); err != nil {
		ss.log.Panicf("can't create local storage path %s: %v", lsp, err)
	}
	setFwdProxyConfig(ss.config)
	ss.events = s.Events()
	ss.api = s.API()
	ss.data = s.Data()

	ss.events.Emit(SystemEventsSelector, APIDInitializedEvent)
}

func RegisterPlugin(plugin PluginInitFunc, pluginData PluginData) {
	pluginInitFuncs = append(pluginInitFuncs, plugin)
	PluginVersionTracker = append(PluginVersionTracker, pluginData)
}

func InitializePlugins(versionNumber string) {
	log := Log()
	log.Debugf("Initializing %d plugins...", len(pluginInitFuncs))
	pie := PluginsInitializedEvent{
		Description: "plugins initialized",
		ApidVersion: versionNumber,
	}
	for _, pif := range pluginInitFuncs {
		pluginData, err := pif(services)
		if err != nil {
			log.Panicf("Error initializing plugin: %s", err)
		}
		pie.Plugins = append(pie.Plugins, pluginData)
	}
	pluginInitFuncs = nil
	Events().Emit(SystemEventsSelector, pie)
	log.Debugf("done initializing plugins")
}

// Shutdown all the plugins that have registered for ShutdownEventSelector.
// This call will block until either all required plugins shutdown, or a timeout occurred.
func ShutdownPluginsAndWait() error {
	shutdownEvent := ShutdownEvent{"apid is going to shutdown"}
	eventChan := Events().Emit(ShutdownEventSelector, shutdownEvent)
	select {
	case event := <-eventChan:
		if e, ok := event.(ShutdownEvent); ok {
			if e == shutdownEvent {
				return nil
			}
		}
		return errors.New("Emit() problem: wrong event delivered")
	case <-time.After(ShutdownTimeout):
		return errors.New("Shutdown timeout")
	}
}

func AllServices() Services {
	return services
}

func Log() LogService {
	return services.Log()
}

func API() APIService {
	return services.API()
}

func Config() ConfigService {
	return services.Config()
}

func Data() DataService {
	return services.Data()
}

func Events() EventsService {
	return services.Events()
}

type servicesSet struct {
	config ConfigService
	log    LogService
	api    APIService
	data   DataService
	events EventsService
}

func (s *servicesSet) API() APIService {
	return s.api
}

func (s *servicesSet) Config() ConfigService {
	return s.config
}

func (s *servicesSet) Data() DataService {
	return s.data
}

func (s *servicesSet) Events() EventsService {
	return s.events
}

func (s *servicesSet) Log() LogService {
	return s.log
}

type systemEvent struct {
	description string
}

type ShutdownEvent struct {
	Description string
}
