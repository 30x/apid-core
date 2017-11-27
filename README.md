# apid-core

[![Build Status](https://travis-ci.org/apid/apid-core.svg)](https://travis-ci.org/apid/apid-core) [![GoDoc](https://godoc.org/github.com/apid/apid-core?status.svg)](https://godoc.org/github.com/apid/apid-core) [![Go Report Card](https://goreportcard.com/badge/github.com/apid/apid-core)](https://goreportcard.com/report/github.com/apid/apid-core)

apid-core is a library that provides a container for publishing APIs that provides core services to its plugins 
including configuration, API publishing, data access, and a local pub/sub event system.

Disambiguation: You might be looking for the executable builder, [apid](https://github.com/apid/apid).  

## Services

apid provides the following services:

* apid.API()
* apid.Config()
* apid.Data()
* apid.Events()
* apid.Log()
 
### Initialization of services and plugins

A driver process must initialize apid and its plugins like this:

    apid.Initialize(factory.DefaultServicesFactory()) // when done, all services are available
    apid.InitializePlugins() // when done, all plugins are running
    api := apid.API() // access the API service
    err := api.Listen() // start the listener


Once apid.Initialize() has been called, all services are accessible via the apid package functions as details above. 

## Plugins

The only requirement of an apid plugin is to register itself upon init(). However, generally plugins will access
the Log service and some kind of driver (via API or Events), so it's common practice to see something like this:
 
    var log apid.LogService
     
    func init() {
      apid.RegisterPlugin(initPlugin)
    }
    
    func initPlugin(services apid.Services) error {
    
      log = services.Log().ForModule("myPluginName") // note: could also access via `apid.Log().ForModule()`
      
      services.API().HandleFunc("/verifyAPIKey", handleRequest)
    }
    
    func handleRequest(w http.ResponseWriter, r *http.Request) {
      // respond to request
    }

## Utils
apid-core/util package offers common util functions for apid plugins:

* Generate/Validate UUIDs
* Long Polling
* Debounce Events


## Running Tests

    go test $(glide novendor)

## apid.Data() service
This service provides the primitives to perform SQL operations on the database. It also provides the
provision to alter DB connection pool settings via ConfigDBMaxConns, ConfigDBIdleConns and configDBConnsTimeout configuration parameters. They currently are defaulted to 1000 connections, 1000 connections and 120 seconds respectively.
More details on this can be found at https://golang.org/pkg/database/sql


## Making http.Client calls through Forward proxy server
If forward proxy server related parameters are set, util.Transport() will provide the Transport roundtripper with the forward proxy parameters set.

