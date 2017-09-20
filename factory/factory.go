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

package factory

import (
	"github.com/apid/apid-core"
	"github.com/apid/apid-core/api"
	"github.com/apid/apid-core/config"
	"github.com/apid/apid-core/data"
	"github.com/apid/apid-core/events"
	"github.com/apid/apid-core/logger"
)

// Don't use values directly - pass to apid.Initialize()
// eg. apid.Initialize(factory.DefaultServicesFactory())

func DefaultServicesFactory() apid.Services {
	return &defaultServices{}
}

type defaultServices struct {
}

func (d *defaultServices) API() apid.APIService {
	return api.CreateService()
}

func (d *defaultServices) Config() apid.ConfigService {
	return config.GetConfig()
}

func (d *defaultServices) Data() apid.DataService {
	return data.CreateDataService()
}

func (d *defaultServices) Events() apid.EventsService {
	return events.CreateService()
}

func (d *defaultServices) Log() apid.LogService {
	return logger.Base()
}
