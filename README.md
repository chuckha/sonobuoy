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
* Health analysis of a newly installed cluster a.k.a 'cluster conformance'
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
**NOTE:** Before continuing, we recommend reading our [docs on how sonobuoy works][sonodocs] in order to properly set your configuration prior to execution.

## Local:
If you want to test locally, be certain to update the local `config.json` to point to a valid `KUBECONFIG`.  Depending on your settings, you may need to update your [RBAC][rbac] rules.  Once built, simply execute: 

```
$ ./sonobuoy -v 5 -logtostderr 
```

The results will be placed under a local ./results directory which can then be uncompressed and inspected.

## Containerized: 
**NOTE:** When running containerized you will need to add a [PVC][pvc] to your submission in order for you to record your results for further analysis.  Be certain to also update your [config-map][results] to point to the [mount location][mount] of your PVC.

Standup: 
```
$ kubectl create -f yaml/
```
Teardown: 
```
$ kubectl delete -f yaml/
```

# Continuous Deployment

The repo at github.com/heptio/sonobuoy is built by Heptio's jenkins instance at https://jenkins.i.heptio.com.

- `master` is continually deployed to `gcr.io/heptio-images/sonobuoy:latest`. [![Build Status](https://jenkins.i.heptio.com/buildStatus/icon?job=sonobuoy-master-deployer)](https://jenkins.i.heptio.com/job/sonobuoy-master-deployer/)
- All tags on the `master` branch are deployed to `gcr.io/heptio-images/sonobuoy:<tag>` [![Build Status](https://jenkins.i.heptio.com/buildStatus/icon?job=sonobuoy-tag-deployer&build=1)](https://jenkins.i.heptio.com/job/sonobuoy-tag-deployer/1/)
- All pull requests destined for `master` are built and tested by jenkins, but no docker images are created or deployed.

"HaVe FuN sToRmInG tHe CaStLe!"

[kubernetes]: https://github.com/kubernetes/kubernetes/
[mount]: https://kubernetes.io/docs/concepts/storage/persistent-volumes/#claims-as-volumes 
[pvc]: https://kubernetes.io/docs/concepts/storage/persistent-volumes/#persistentvolumeclaims
[rbac]: https://kubernetes.io/docs/admin/authorization/rbac/
[results]: https://github.com/heptio/sonobuoy/blob/master/yaml/sonobuoy-configmap.yaml#L44
[sonodocs]: https://github.com/heptio/sonobuoy/blob/master/doc/modusoperandi.md
