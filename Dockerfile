# Copyright 2017 Heptio Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

FROM buildpack-deps:jessie-scm
MAINTAINER Timothy St. Clair "tstclair@heptio.com"  

# Ensure we're using the latest ansible
RUN apt-key adv --keyserver keyserver.ubuntu.com --recv-keys 93C4A3FD7BB9C367
RUN echo "deb http://ppa.launchpad.net/ansible/ansible/ubuntu trusty main" | tee -a /etc/apt/sources.list

#TODO remove debugging tools, e.g. vim closer to release
RUN apt-get update && apt-get -y --no-install-recommends install \
    ca-certificates \
    ansible \
    python-pip \
    vim \
    && rm -rf /var/cache/apt/* \
    && rm -rf /var/lib/apt/lists/*
ADD sonobuoy /sonobuoy 
ADD e2e.test /e2e.test
#USER nobody:nobody

CMD ["/bin/sh", "-c", "/sonobuoy -v 3 -logtostderr"]
