#/bin/bash
#
# Copyright 2017 Heptio Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# NOTES: 
# 1. This assumes you're running from the sonobuoy directory
# 2. This is indeed a hack and is due to how upstream has 
# structured its libraries. see - https://github.com/kubernetes/kubernetes/issues/43246 
# for complete details. 
# 3. Before executing this script, navigate to your $GOPATH/src/k8s.io/kubernetes 
#    and run ./hack/godep-restore.sh

rm -rf Godeps vendor
export KPATH=$GOPATH/src/k8s.io/kubernetes
export GOPATH=$GOPATH:$KPATH/staging

$KPATH/hack/generate-bindata.sh
godep save ./...
cp -L -R $KPATH/vendor/k8s.io/api* $KPATH/vendor/k8s.io/client-go ./vendor/k8s.io/
cp $KPATH/test/e2e/generated/bindata.go ./vendor/k8s.io/kubernetes/test/e2e/generated/bindata.go

