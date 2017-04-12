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

package dispatch

// A Dispatcher is something that can declare resources in kubernetes and clean
// them up, for instance to run sonobuoy agents
type Dispatcher interface {
	Dispatch() error
	Cleanup() error
	SessionID() string
}

// A Waiter is something that can dispatch "ephemeral" kubernetes
// resources that can complete in some way like Jobs, etc.
type Waiter interface {
	DispatchAndWait() error
	Wait() error
}
