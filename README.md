# Sonobuoy
[Kubernetes][kubernetes] is an open source system for managing containerized applications across multiple hosts, providing basic mechanisms for deployment, maintenance, and scaling of applications.  As of today, there are various ways to stand up a kubernetes cluster using an assortment of tools and different deployment patterns.  Some of these tools include: 

* https://github.com/kubernetes/kops
* https://github.com/kubernetes-incubator/kargo
* https://github.com/openshift/openshift-ansible
* https://github.com/kubernetes-incubator/bootkube
* https://github.com/kubernetes/kubeadm
* ...

Needless to say, there are a plethora of solutions that exist today, but there is currently no uniform mechanism by which we can generate reports to determine if a cluster, or its workloads, are “healthy”. 

[![Cluster Verification](http://img.youtube.com/vi/jr0JaXfKj68/0.jpg)](http://www.youtube.com/watch?v=jr0JaXfKj68)

## Goals
This gap in recording/reporting, is the primary role that sonobuoy aims to fill.  In essence, Sonobuoy is an operator tool for inspecting a cluster's configuration and recording its state and behavioral characteristics.

### Use Cases
* Selective dump of kubernetes resource objects for cluster snap-shotting 
  * Workload reporting
  * Managing and rectifying cluster state
  * Workload debugging
  * Configuration validation
  * State of record for disaster recovery
* Health analysis of a newly installed cluster
* ... 

## Non-goals:
Sonobuoy’s primary function is data collection, and therefore, any analysis of the results is specifically a non-goal of sonobuoy.   

# Building 
Sonobuoy can be built as either a standalone application, or as a Docker container that can be run as an introspective job on your cluster.

## Standalone:
```
$ go get github.com/heptio/sonobuoy
$ make local 
```  

## Containerized: 
```
$ sudo make all 
```

# Running Sonobuoy
Sonobuoy takes as input, a single `config.json` file that can either be located in its local directory or under `/etc/sonobuoy/config.json`.  For convenience, there is an example `config.json` provided in the root of the repository to allow operators to simply download the repo and evaluate sonobuoy without having to download containers.

For a complete list of all the input options look [here][inargs]. 

TODO: Add https://www.websequencediagrams.com/ example b/c it's slick.

## Local:
If you want to test locally, be certain to update the local `config.json` to point to a valid `KUBECONFIG`.  Depending on your settings, you may need to update your [RBAC][rbac] rules.  Once built, simply execute: 

```
$ ./sonobuoy -v 5 -logtostderr 
```

The results will be placed under a local ./results directory which can then be uncompressed and inspected.

## Containerized: 

TODO: outline PVC usage.

Standup: 
```
$ kubectl create -f yaml/
```
Teardown: 
```
$ kubectl delete -f yaml/
```

[kubernetes]: https://github.com/kubernetes/kubernetes/
[rbac]: https://kubernetes.io/docs/admin/authorization/rbac/
[inargs]: https://github.com/heptio/sonobuoy/blob/master/pkg/discovery/config.go#L41