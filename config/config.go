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

package config

import (
	"github.com/apid/apid-core"
	"github.com/spf13/viper"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	localStoragePathKey     = "local_storage_path"
	localStoragePathDefault = "/var/tmp/apid"

	configFileEnvVar = "APID_CONFIG_FILE"

	configFileType    = "yaml"
	configFileNameKey = "apid_config_filename"
	configPathKey     = "apid_config_path"

	defaultConfigFilename = "apid_config.yaml"
	defaultConfigPath     = "."
)

var configlock sync.Mutex

type ConfigMgr struct {
	sync.Mutex
	vcfg *viper.Viper
}

// Wrapper function to make the viper calls thread safe
var cfg *ConfigMgr

func (c *ConfigMgr) SetDefault(key string, value interface{}) {
	c.Lock()
	c.vcfg.SetDefault(key, value)
	c.Unlock()
}

func (c *ConfigMgr) Set(key string, value interface{}) {
	c.Lock()
	c.vcfg.Set(key, value)
	c.Unlock()
}

func (c *ConfigMgr) Get(key string) interface{} {
	c.Lock()
	defer c.Unlock()
	return c.vcfg.Get(key)
}

func (c *ConfigMgr) GetBool(key string) bool {
	c.Lock()
	defer c.Unlock()
	return c.vcfg.GetBool(key)
}

func (c *ConfigMgr) GetFloat64(key string) float64 {
	c.Lock()
	defer c.Unlock()
	return c.vcfg.GetFloat64(key)
}

func (c *ConfigMgr) GetInt(key string) int {
	c.Lock()
	defer c.Unlock()
	return c.vcfg.GetInt(key)
}

func (c *ConfigMgr) GetString(key string) string {
	cfg.Lock()
	defer cfg.Unlock()
	return c.vcfg.GetString(key)
}

func (c *ConfigMgr) GetDuration(key string) time.Duration {
	cfg.Lock()
	defer cfg.Unlock()
	return c.vcfg.GetDuration(key)
}

func (c *ConfigMgr) IsSet(key string) bool {
	c.Lock()
	defer c.Unlock()
	return c.vcfg.IsSet(key)
}

func GetConfig() apid.ConfigService {
	configlock.Lock()
	defer configlock.Unlock()
	if cfg == nil {

		vcfg := viper.New()

		// for config file search path
		vcfg.SetConfigType(configFileType)

		vcfg.SetDefault(configPathKey, defaultConfigPath)
		configFilePath := vcfg.GetString(configPathKey)
		vcfg.AddConfigPath(configFilePath)

		vcfg.SetDefault(configFileNameKey, defaultConfigFilename)
		configFileName := vcfg.GetString(configFileNameKey)
		configFileName = strings.TrimSuffix(configFileName, ".yaml")
		vcfg.SetConfigName(configFileName)

		// for user-specified absolute config file
		configFile, ok := os.LookupEnv(configFileEnvVar)
		if ok {
			vcfg.SetConfigFile(configFile)
		}

		vcfg.SetDefault(localStoragePathKey, localStoragePathDefault)

		err := vcfg.ReadInConfig()
		if err != nil {
			log.Printf("Error in config file '%s': %s", configFileNameKey, err)
		}

		vcfg.SetEnvPrefix("apid") // eg. env var "APID_SOMETHING" will bind to config var "something"
		vcfg.AutomaticEnv()

		cfg = &ConfigMgr{vcfg: vcfg}
	}
	return cfg
}
