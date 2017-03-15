package apid

import (
	"os"
)

const (
	SystemEventsSelector  EventSelector = "system event"
	ShutdownEventSelector EventSelector = "shutdown event"
)

var (
	APIDInitializedEvent = systemEvent{"apid initialized"}
	APIListeningEvent    = systemEvent{"api listening"}

	pluginInitFuncs []PluginInitFunc
	services        Services

	shutdownChan chan int
)

type Services interface {
	API() APIService
	Config() ConfigService
	Data() DataService
	Events() EventsService
	Log() LogService
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

	ss.events = s.Events()
	ss.api = s.API()
	ss.data = s.Data()

	shutdownChan = make(chan int)

	ss.events.Emit(SystemEventsSelector, APIDInitializedEvent)
}

func RegisterPlugin(plugin PluginInitFunc) {
	pluginInitFuncs = append(pluginInitFuncs, plugin)
}

func InitializePlugins() {
	log := Log()
	log.Debugf("Initializing %d plugins...", len(pluginInitFuncs))
	pie := PluginsInitializedEvent{
		Description: "plugins initialized",
	}
	for _, pif := range pluginInitFuncs {
		pluginData, err := pif(services)
		if err != nil {
			log.Panicf("Error initializing plugin: %s", err)
		}
		pie.Plugins = append(pie.Plugins, pluginData)
	}
	Events().Emit(SystemEventsSelector, pie)
	pluginInitFuncs = nil
	log.Debugf("done initializing plugins")
}

func ShutdownPlugins() {
	Events().EmitWithCallback(ShutdownEventSelector, ShutdownEvent{"apid is going to shutdown"}, shutdownHandler)
}

func shutdownHandler(event Event) {
	log := Log()
	log.Debugf("shutdown apid")
	shutdownChan <- 1
}

/* wait for the shutdown of registered graceful-shutdown plugins, blocking until the required plugins finish shutdown
 * this is used to prevent the main from exiting
 */
func WaitPluginsShutdown() {
	<-shutdownChan
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
