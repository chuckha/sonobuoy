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

rm -rf Godeps vendor
export GOPATH=$GOPATH:$KPATH/staging
godep save ./...
cp -Lr $KPATH/vendor/k8s.io/api* $KPATH/vendor/k8s.io/client-go ./vendor/k8s.io/
# TODO Add generated bin-data file from upstream, this needs to be fixed
