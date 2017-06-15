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

package events

import "github.com/30x/apid-core"

// simple pub/sub to deliver events to listeners based on a selector string

const configChannelBufferSize = "events_buffer_size"

var log apid.LogService
var config apid.ConfigService

func CreateService() apid.EventsService {
	if log == nil {
		log = apid.Log().ForModule("events")
		config = apid.Config()
		config.SetDefault(configChannelBufferSize, 5)
	}
	return &eventManager{}
}
