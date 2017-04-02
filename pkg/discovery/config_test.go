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
	"encoding/json"
	"io/ioutil"
	"reflect"
	"testing"
)

func TestDefaults(t *testing.T) {
	var dc1, dc2 Config
	SetConfigDefaults(&dc1)
	SetConfigDefaults(&dc2)

	if reflect.DeepEqual(&dc2, &dc1) {
		t.Fatalf("Defaults should not match UUIDs collided")
	}

	// set UUIDs to be the same
	dc1.UUID = "0xDEADBEEF"
	dc2.UUID = "0xDEADBEEF"

	if !reflect.DeepEqual(&dc2, &dc1) {
		t.Fatalf("Defaults should match but didn't")
	}
}

func TestSaveAndLoad(t *testing.T) {
	var cfg SonoCfg
	SetConfigDefaults(&cfg.DC)

	if blob, err := json.Marshal(&cfg); err == nil {
		if err = ioutil.WriteFile("./config.json", blob, 0644); err != nil {
			t.Fatalf("Failed to write default config.json: %v", err)
		}
	} else {
		t.Fatalf("Failed to serialize ", err)
	}

	// TODO: Randomly change values

	_, dc2 := LoadConfig()
	if !reflect.DeepEqual(dc2, &cfg.DC) {
		t.Fatalf("Defaults should match but didn't \n\n%v\n\n%v", dc2, &cfg.DC)
	}

}
