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

package util


import (
	"github.com/google/uuid"
	"net/http"
	"net/url"
	"github.com/apid/apid-core"
)

const (
	configfwdProxyURL	=   "configfwdProxyURL"
	configfwdProxyUser	=   "configfwdProxyUser"
	configfwdProxyPasswd	=   "configfwdProxyPasswd"
	configfwdProxyProtocol  =   "configfwdProxyProtocol"
	configfwdProxyPort      =   "configfwdProxyPort"
)

var config apid.ConfigService


func IsValidUUID(id string) bool {
	_, err := uuid.Parse(id)
	return err == nil
}

func GenerateUUID() string {
	return uuid.New().String()
}

func Transport() *http.Transport {
	var tr http.Transport
	var pURL *url.URL
	var err error
	// Apigee Forward Proxy
	fwdPrxy := config.GetString(configfwdProxyURL)
	fwdPrxyPro := config.GetString(configfwdProxyProtocol)
	fwdPrxyUser := config.GetString(configfwdProxyUser)
	fwdPrxyPass := config.GetString(configfwdProxyPasswd)
	fwdPrxyPort := config.GetString(configfwdProxyPort)

	if fwdPrxy != "" && fwdPrxyPro != "" && fwdPrxyUser != "" && fwdPrxyPort != "" {
		pURL, err = url.Parse(fwdPrxyPro + "//" + fwdPrxyUser + ":" + fwdPrxyPass + "@" + fwdPrxy + ":" + fwdPrxyPort)
		if err != nil {
			panic("Error parsing proxy URL")
		}
	} else if fwdPrxy != "" && fwdPrxyPro != "" && fwdPrxyPort != "" {
		pURL, err = url.Parse(fwdPrxyPro + "//" + fwdPrxy + ":" + fwdPrxyPort)
		if err != nil {
			panic("Error parsing proxy URL")
		}
	}

	if pURL != nil {
		tr = http.Transport{
			Proxy:           http.ProxyURL(pURL),
		}
	} else {
		tr = http.Transport{
		}
	}
	return &tr
}

