# Sonobuoy
[Kubernetes][kubernetes] is an open source system for managing containerized applications across multiple hosts, providing basic mechanisms for deployment, maintenance, and scaling of applications.  As of today, there are various ways to stand up a kubernetes cluster using an assortment of tools and different deployment patterns.  Some of these tools are: 

* https://github.com/kubernetes/kops
* https://github.com/kubernetes-incubator/kargo
* https://github.com/openshift/openshift-ansible
* https://github.com/kubernetes-incubator/bootkube
* https://github.com/kubernetes/kubeadm
* ...

Needless to say, there are a plethora of solutions that exist today, but there is currently no uniform mechanism by which we can determine if a cluster is “healthy”.  

## Goals
This gap in cluster health and analysis, is the primary role that sonobuoy aims to fill.  In essence, Sonobouy is an operator tool for inspecting a cluster’s configuration and analyzing its behavioral characteristics, e.g.:   

TODO: List more details and link to configuration settings as we are further down the road.

## Non-goals:
Make julian fries!  

# Building 
You can build and test either as a standalone go application or as a Docker container.

## Standalone:
```
$ go get github.com/heptio/sonobuoy
```  

## Containerized: 
```
$ sudo make all 
```

# Configure and Execute

## Standalone:
Assuming your testing on a local cluster, it will use the local `config.json`, which 
you can override.
```
$ ./sonobuoy -v 3 -logtostderr 
```

## On the cluster 
Standup: 
```
$ kubectl create -f yaml/
```
Teardown: 
```
$ kubectl delete -f yaml/
```

[kubernetes]: https://github.com/kubernetes/kubernetes/ "Kubernetes"