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

package ansible

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"text/template"

	"bytes"

	"github.com/golang/glog"
	"github.com/renstrom/dedent"
)

// Config represents the configuration of ansible for reaching out to physical
// hosts in the cluster.
type Config struct {
	Hosts      []string `json:"hosts"`
	RemoteUser string   `json:"remoteuser"`
	OutputDir  string   `json:"outputdir"`
}

const configDir string = ".ansiblecfg"
const outDir string = "node-data"

func writeAnsibleConfig(cfg *Config) error {
	prefix := path.Join(cfg.OutputDir, configDir)

	confcontents, err := ansibleConfFile(cfg)
	if err != nil {
		return err
	}

	hostcontents, err := ansibleHostFile(cfg)
	if err != nil {
		return err
	}

	if err = os.MkdirAll(prefix, 0755); err != nil {
		return err
	}
	if err = ioutil.WriteFile(path.Join(prefix, "ansible.cfg"), confcontents, 0644); err != nil {
		return err
	}
	if err = ioutil.WriteFile(path.Join(prefix, "hosts"), hostcontents, 0644); err != nil {
		return err
	}

	return nil
}

func runAnsible(cfg *Config) error {
	// Find the ansible command
	ansiblePath, err := exec.LookPath("ansible")
	if err != nil {
		return fmt.Errorf("could not find ansible binary in $PATH: %v", err)
	}
	cmd := exec.Command(ansiblePath, "--ssh-common-args=-o BatchMode=yes", "all", "-m", "setup", "--tree", path.Join(cfg.OutputDir, outDir))

	// Write ansible config
	if err = writeAnsibleConfig(cfg); err != nil {
		return fmt.Errorf("could not write ansible config to disk: %v", err)
	}

	// Customize environment variables for running ansible
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "ANSIBLE_CONFIG="+path.Join(cfg.OutputDir, configDir, "ansible.cfg"))
	cmd.Env = append(cmd.Env, "ANSIBLE_INVENTORY="+path.Join(cfg.OutputDir, configDir, "hosts"))
	cmd.Stderr = os.Stderr

	// Start the command in the background
	err = cmd.Start()
	if err != nil {
		return err
	}

	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("ansible returned an error: %v", err)
	}

	return nil
}

func ansibleConfFile(cfg *Config) ([]byte, error) {
	tmplstr := dedent.Dedent(`
		[defaults]
		ask_pass = False
		ask_sudo_pass = False
		ask_vault_pass = False
		become_ask_pass = False
		host_key_checking = False
		remote_user = {{.RemoteUser}}
	`)

	var result bytes.Buffer
	tmpl := template.New("acfg")
	template.Must(tmpl.Parse(tmplstr))

	if err := tmpl.Execute(&result, cfg); err != nil {
		return nil, fmt.Errorf("could not construct valid ansible template from config: %v", err)
	}

	glog.V(5).Infof("Ansible config: \n%v\n", result.String())
	return result.Bytes(), nil
}

func ansibleHostFile(cfg *Config) ([]byte, error) {
	var result bytes.Buffer
	var err error

	for _, host := range cfg.Hosts {
		if _, err = result.WriteString(host + "\n"); err != nil {
			return result.Bytes(), err
		}
	}
	return result.Bytes(), nil
}

// GatherHostData call out to ansible to SSH to each node and gather host fact
// information, writing them out to the specified output directory.
func GatherHostData(cfg *Config) error {
	glog.Infof("Gathering host data with ansible\n")
	err := runAnsible(cfg)
	if err != nil {
		glog.Errorf("Error running ansible: %v\n", err)
	}

	return err
}
