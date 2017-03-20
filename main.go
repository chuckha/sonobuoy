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
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/golang/glog"
	"github.com/spf13/viper"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
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

// loadConfig will parse input + config file and return a clientset
func loadConfig() kubernetes.Interface {
	var config *rest.Config
	var err error

	flag.Parse()

	// leverages a file|(ConfigMap)
	// to be located at /etc/sonobuoy/config
	viper.SetConfigType("json")
	viper.SetConfigName("config")
	viper.AddConfigPath("/etc/sonobuoy/")
	viper.AddConfigPath(".")
	viper.SetDefault("kubeconfig", "")
	if err = viper.ReadInConfig(); err != nil {
		panic(err.Error())
	}

	// allows for running both in & out of cluster
	kubeconfig := viper.GetString("kubeconfig")
	if len(kubeconfig) > 0 {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	} else {
		config, err = rest.InClusterConfig()
	}
	if err != nil {
		panic(err.Error())
	}

	// creates the clientset from kubeconfig
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	return clientset
}

// main entry point of the program
func main() {

	//clientset := loadConfig()

}
