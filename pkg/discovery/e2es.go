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

package discovery

import (
	"github.com/golang/glog"
)

// kicker for e2es, look at discovery.
func rune2e(outpath string, dc *Config) error {
	var err error
	// call battery.test with a valid set of args to output into the directory
	if dc.Runtests {
		glog.Info("Running tests...")
	}

	return err
}
