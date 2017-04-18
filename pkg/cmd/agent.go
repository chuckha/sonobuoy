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

package cmd

import (
	"os"

	"github.com/golang/glog"
	"github.com/heptio/sonobuoy/pkg/agent"
	"github.com/spf13/cobra"
)

var flagPhoneHomeURL string
var flagNodeName string
var flagChrootDir string

func init() {
	agentCmd.Flags().StringVarP(&flagPhoneHomeURL, "phone-home-url", "u", "", "URL to submit results to")
	agentCmd.Flags().StringVarP(&flagNodeName, "node-name", "n", "", "This node's name to use when phoning home")
	agentCmd.Flags().StringVarP(&flagChrootDir, "chroot", "c", "", "chroot directory for the host's filesystem")
	RootCmd.AddCommand(agentCmd)
}

var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Gather information about the local machine",
	Run:   runAgent,
}

func overrideConfig(cfg *agent.Config) {
	if flagPhoneHomeURL != "" {
		cfg.PhoneHomeURL = flagPhoneHomeURL
	}
	if flagChrootDir != "" {
		cfg.ChrootDir = flagChrootDir
	}
	if flagNodeName != "" {
		cfg.NodeName = flagNodeName
	}
}

func runAgent(cmd *cobra.Command, args []string) {
	agentConfig, err := agent.LoadConfig()
	if err != nil {
		glog.Errorf("Error loading agent configuration: %v\n", err)
		os.Exit(1)
	}

	overrideConfig(agentConfig)

	err = agent.Run(agentConfig)
	if err != nil {
		glog.Errorln(err)
		os.Exit(1)
	}
}
