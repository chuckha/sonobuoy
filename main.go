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

package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/golang/glog"
	"github.com/heptio/sonobuoy/pkg/discovery"
)

// setup a signal hander to gracefully exit
func sigHandler() <-chan struct{} {
	stop := make(chan struct{})
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c,
			syscall.SIGINT,  // Ctrl+C
			syscall.SIGTERM, // Termination Request
			syscall.SIGSEGV, // FullDerp
			syscall.SIGABRT, // Abnormal termination
			syscall.SIGILL,  // illegal instruction
			syscall.SIGFPE)  // floating point - this is why we can't have nice things
		sig := <-c
		glog.Warningf("Signal (%v) Detected, Shutting Down", sig)
		close(stop)
	}()
	return stop
}

// main entry point of the program
func main() {
	if errlist := discovery.Run(sigHandler()); errlist != nil {
		for _, err := range errlist {
			glog.Errorf("%v", err)
		}
		os.Exit(1)
	} else {
		os.Exit(0)
	}
}
