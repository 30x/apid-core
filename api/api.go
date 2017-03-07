package api

import (
	"expvar"
	"fmt"
	"net/http"

	"net"

	"github.com/30x/apid-core"
	"github.com/30x/goscaffold"
	"github.com/gorilla/mux"
)

const (
	configAPIListen   = "api_listen"
	configExpVarPath  = "api_expvar_path"
	configReadyPath   = "api_ready"
	configHealthyPath = "api_healthy"
)

var log apid.LogService
var config apid.ConfigService
var requests *expvar.Map = expvar.NewMap("requests")

func CreateService() apid.APIService {
	config = apid.Config()
	log = apid.Log().ForModule("api")

	config.SetDefault(configAPIListen, "127.0.0.1:9000")
	config.SetDefault(configReadyPath, "/ready")
	config.SetDefault(configHealthyPath, "/healthy")

	listen := config.GetString(configAPIListen)
	h, p, err := net.SplitHostPort(listen)
	if err != nil {
		log.Panicf("%s config: err parsing '%s': %v", configAPIListen, listen, err)
	}
	var ip net.IP
	if h != "" {
		ips, err := net.LookupIP(h)
		if err != nil {
			log.Panicf("%s config: unable to resolve IP for '%s': %v", configAPIListen, listen, err)
		}
		ip = ips[0]
	}
	port, err := net.LookupPort("tcp", p)
	if err != nil {
		log.Panicf("%s config: unable to resolve port for '%s': %v", configAPIListen, listen, err)
	}

	log.Infof("will open api port %d bound to %s", port, ip)

	r := mux.NewRouter()
	rw := &router{r}
	scaffold := goscaffold.CreateHTTPScaffold()
	if ip != nil {
		scaffold.SetlocalBindIPAddressV4(ip)
	}
	scaffold.SetInsecurePort(port)
	scaffold.CatchSignals()

	// Set an URL that may be used by a load balancer to test if the server is ready to handle requests
	if config.GetString(configReadyPath) != "" {
		scaffold.SetReadyPath(config.GetString(configReadyPath))
	}

	// Set an URL that may be used by infrastructure to test
	// if the server is working or if it needs to be restarted or replaced
	if config.GetString(configHealthyPath) != "" {
		scaffold.SetReadyPath(config.GetString(configHealthyPath))
	}

	return &service{rw, scaffold}
}

type service struct {
	*router
	scaffold *goscaffold.HTTPScaffold
}

func (s *service) Listen() error {
	err := s.scaffold.StartListen(s.r)
	if err != nil {
		return err
	}

	apid.Events().Emit(apid.SystemEventsSelector, apid.APIListeningEvent)

	return s.scaffold.WaitForShutdown()
}

func (s *service) Close() {
	s.scaffold.Shutdown(nil)
	s.scaffold = nil
}

func (s *service) InitExpVar() {
	if config.IsSet(configExpVarPath) {
		log.Infof("expvar available on path: %s", config.Get(configExpVarPath))
		s.HandleFunc(config.GetString(configExpVarPath), expvarHandler)
	}
}

// for testing
func (s *service) Router() apid.Router {
	s.InitExpVar()
	return s
}

func (s *service) Vars(r *http.Request) map[string]string {
	return mux.Vars(r)
}

func expvarHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	fmt.Fprint(w, "{\n")
	first := true
	expvar.Do(func(kv expvar.KeyValue) {
		if !first {
			fmt.Fprint(w, ",\n")
		}
		first = false
		fmt.Fprintf(w, "%q: %s", kv.Key, kv.Value)
	})
	fmt.Fprint(w, "\n}\n")
}

type router struct {
	r *mux.Router
}

func (r *router) Handle(path string, handler http.Handler) apid.Route {
	log.Infof("Handle %s: %v", path, handler)
	return &route{r.r.Handle(path, handler)}
}

func (r *router) HandleFunc(path string, handlerFunc http.HandlerFunc) apid.Route {
	log.Infof("Handle %s: %v", path, handlerFunc)
	return &route{r.r.HandleFunc(path, handlerFunc)}
}

func (r *router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	requests.Add(req.URL.Path, 1)
	log.Infof("Handling %s", req.URL.Path)
	r.r.ServeHTTP(w, req)
}

type route struct {
	r *mux.Route
}

func (r *route) Methods(methods ...string) apid.Route {
	return &route{r.r.Methods(methods...)}
}
