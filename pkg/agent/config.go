/*
Copyright 2017 Heptio Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package agent

import (
	"os"

	"github.com/spf13/viper"
)

// Config represents the configuration of what the sonobuoy agent should be doing
type Config struct {
	// Ansible determines whether we use the ansible `setup` gatherer
	Ansible bool `json:"ansible,omitempty"`
	// SystemdLogs determines whether we grab systemd logs with journalctl
	SystemdLogs bool `json:"systemdlogs,omitempty"`
	// SystemdLogMinutes determines how many minutes of logs we want to gather
	SystemdLogMinutes int `json:"systemdlogminutes,omitempty"`
	// PhoneHomeUrl is the URL we talk to for submitting results
	PhoneHomeURL string `json:"phonehomeurl,omitempty"`
	// NodeName is the node name we should call ourselves when sending results
	NodeName string `json:"nodename,omitempty"`
	// ChrootDir is the directory that's expected to contain the host's root filesystem
	ChrootDir string `json:"chrootdir,omitempty"`
}

func setConfigDefaults(ac *Config) {
	ac.Ansible = true
	ac.SystemdLogs = true
	ac.SystemdLogMinutes = 60 * 24
	ac.ChrootDir = "/node"
}

// LoadConfig loads the config file for the sonobuoy agent from known locations, returns a
// Config struct with defaults applied
func LoadConfig() (*Config, error) {
	config := &Config{}
	var err error

	viper.SetConfigType("json")
	viper.SetConfigName("agent")
	viper.AddConfigPath("/etc/sonobuoy")
	viper.AddConfigPath(".")
	viper.BindEnv("phonehomeurl", "PHONE_HOME_URL")
	viper.BindEnv("nodename", "NODE_NAME")

	// Allow specifying a custom config file via the SONOBUOY_CONFIG env var
	if forceCfg := os.Getenv("SONOBUOY_CONFIG"); forceCfg != "" {
		viper.SetConfigFile(forceCfg)
	}

	setConfigDefaults(config)

	if err = viper.ReadInConfig(); err != nil {
		return nil, err
	}

	if err = viper.Unmarshal(config); err != nil {
		return nil, err
	}

	return config, nil
}
