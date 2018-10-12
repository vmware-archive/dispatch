#!/usr/bin/bash
# Copyright 2017-2018 VMware, Inc. All Rights Reserved.
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
set -euf -o pipefail

echo "installing pyyaml"
pip install pyyaml

echo "installing - docker compose"
curl -o /usr/local/bin/docker-compose -L'#' "https://github.com/docker/compose/releases/download/1.11.1/docker-compose-$(uname -s)-$(uname -m)" 
chmod +x /usr/local/bin/docker-compose

echo "installing - jq"
curl -o /usr/bin/jq -L'#' "https://github.com/stedolan/jq/releases/download/jq-1.5/jq-linux64"
chmod +x /usr/bin/jq
