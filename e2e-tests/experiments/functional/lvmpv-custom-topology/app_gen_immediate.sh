# Copyright 2020-2021 The OpenEBS Authors. All rights reserved.
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
#!/bin/bash
  
set -e

mkdir app_yamls_immediate

for i in $(seq 1 5)
do
        sed "s/pvc-custom-topology/pvc-custom-topology-$i/g" busybox_immediate.yml > app_yamls_immediate/busybox-$i.yml
        sed -i "s/busybox-custom-topology-test/busybox-custom-topology-test-$i/g" app_yamls_immediate/busybox-$i.yml
        sed -i "s/lvmpv-custom-topology/lvmpv-custom-topology-immediate/g" app_yamls_immediate/busybox-$i.yml
done
