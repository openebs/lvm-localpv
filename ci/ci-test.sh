#!/usr/bin/env bash
# Copyright 2021 The OpenEBS Authors.
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


set -e

# setup the lvm volume group to create the volume
truncate -s 1024G /tmp/disk.img
disk=`losetup -f /tmp/disk.img --show`
pvcreate "$disk"
vgcreate lvmvg "$disk"

LVM_OPERATOR=deploy/lvm-operator.yaml
SNAP_CLASS=deploy/sample/lvmsnapclass.yaml

export LVM_NAMESPACE="openebs"
export TEST_DIR="tests"
export NAMESPACE="kube-system"

# Prepare env for running BDD tests
# Minikube is already running
kubectl apply -f $LVM_OPERATOR
kubectl apply -f $SNAP_CLASS

dumpAgentLogs() {
  NR=$1
  AgentPOD=$(kubectl get pods -l app=openebs-lvm-node -o jsonpath='{.items[0].metadata.name}' -n "$NAMESPACE")
  kubectl describe po "$AgentPOD" -n "$NAMESPACE"
  printf "\n\n"
  kubectl logs --tail="${NR}" "$AgentPOD" -n "$NAMESPACE" -c openebs-lvm-plugin
  printf "\n\n"
}

dumpControllerLogs() {
  NR=$1
  ControllerPOD=$(kubectl get pods -l app=openebs-lvm-controller -o jsonpath='{.items[0].metadata.name}' -n "$NAMESPACE")
  kubectl describe po "$ControllerPOD" -n "$NAMESPACE"
  printf "\n\n"
  kubectl logs --tail="${NR}" "$ControllerPOD" -n "$NAMESPACE" -c openebs-lvm-plugin
  printf "\n\n"
}

isPodReady(){
  [ "$(kubectl get po "$1" -o 'jsonpath={.status.conditions[?(@.type=="Ready")].status}' -n "$NAMESPACE")" = 'True' ]
}

isDriverReady(){
  for pod in $lvmDriver;do
    isPodReady "$pod" || return 1
  done
}

waitForLVMDriver() {
  period=120
  interval=1

  i=0
  while [ "$i" -le "$period" ]; do
    lvmDriver="$(kubectl get pods -l role=openebs-lvm -o 'jsonpath={.items[*].metadata.name}' -n "$NAMESPACE")"
    if isDriverReady "$lvmDriver"; then
      return 0
    fi

    i=$(( i + interval ))
    echo "Waiting for lvm-driver to be ready..."
    sleep "$interval"
  done

  echo "Waited for $period seconds, but all pods are not ready yet."
  return 1
}

# wait for lvm-driver to be up
waitForLVMDriver

cd $TEST_DIR

kubectl get po -n "$NAMESPACE"

set +e

echo "running ginkgo test case"

ginkgo -v

if [ $? -ne 0 ]; then

sudo pvscan --cache

sudo lvdisplay

echo "******************** LVM Controller logs***************************** "
dumpControllerLogs 1000

echo "********************* LVM Agent logs *********************************"
dumpAgentLogs 1000

echo "get all the pods"
kubectl get pods -owide --all-namespaces

echo "get pvc and pv details"
kubectl get pvc,pv -oyaml --all-namespaces

echo "get snapshot details"
kubectl get volumesnapshot.snapshot -oyaml --all-namespaces

echo "get sc details"
kubectl get sc --all-namespaces -oyaml

echo "get lvm volume details"
kubectl get lvmvolumes.local.openebs.io -n openebs -oyaml

echo "get lvm snapshot details"
kubectl get lvmsnapshots.local.openebs.io -n openebs -oyaml

exit 1
fi

printf "\n\n######### All test cases passed #########\n\n"
