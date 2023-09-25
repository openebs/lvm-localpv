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

LVM_OPERATOR="$(realpath deploy/lvm-operator.yaml)"
SNAP_CLASS="$(realpath deploy/sample/lvmsnapclass.yaml)"

export LVM_NAMESPACE="openebs"
export TEST_DIR="tests"
export NAMESPACE="kube-system"

# allow override
if [ -z "${KUBECONFIG}" ]
then
  export KUBECONFIG="${HOME}/.kube/config"
fi

# foreign systemid for the testing environment.
FOREIGN_LVM_SYSTEMID="openebs-ci-test-system"
FOREIGN_LVM_CONFIG="global{system_id_source=lvmlocal}local{system_id=${FOREIGN_LVM_SYSTEMID}}"

# RAID info for corresponding tests
RAID_COUNT=5

# RAID info for corresponding tests
RAID_COUNT=5

# Clean up generated resources for successive tests.
cleanup_loopdev() {
  sudo losetup -l | grep '(deleted)' | awk '{print $1}' \
    | while IFS= read -r disk
      do
        sudo losetup -d "${disk}"
      done
}

cleanup_lvmvg() {
  if [ -f /tmp/openebs_ci_disk.img ]
  then
    sudo vgremove lvmvg -y || true
    rm /tmp/openebs_ci_disk.img
  fi
  cleanup_loopdev
}

cleanup_foreign_lvmvg() {
  if [ -f /tmp/openebs_ci_foreign_disk.img ]
  then
    sudo vgremove foreign_lvmvg --config="${FOREIGN_LVM_CONFIG}" -y || true
    rm /tmp/openebs_ci_foreign_disk.img
  fi
  cleanup_loopdev
}

cleanup_raidvg() {
  sudo vgremove raidvg -y || true

  for IMG in `seq ${RAID_COUNT}`
  do
    if [ -f /tmp/openebs_ci_raid_disk_${IMG}.img ]
    then
      rm /tmp/openebs_ci_raid_disk_${IMG}.img
    fi
  done

  cleanup_loopdev
}

cleanup() {
  set +e

  echo "Cleaning up test resources"

  cleanup_lvmvg
  cleanup_foreign_lvmvg
  cleanup_raidvg

  kubectl delete pvc -n openebs lvmpv-pvc
  kubectl delete -f "${SNAP_CLASS}"
  kubectl delete -f "${LVM_OPERATOR}"

  # always return true
  return 0
}
# trap "cleanup 2>/dev/null" EXIT
[ -n "${CLEANUP_ONLY}" ] && cleanup 2>/dev/null && exit 0
[ -n "${RESET}" ] && cleanup 2>/dev/null

# setup the lvm volume group to create the volume
cleanup_lvmvg
truncate -s 1024G /tmp/openebs_ci_disk.img
disk="$(sudo losetup -f /tmp/openebs_ci_disk.img --show)"
sudo pvcreate "${disk}"
sudo vgcreate lvmvg "${disk}"

# setup a foreign lvm to test
cleanup_foreign_lvmvg
truncate -s 1024G /tmp/openebs_ci_foreign_disk.img
foreign_disk="$(sudo losetup -f /tmp/openebs_ci_foreign_disk.img --show)"
sudo pvcreate "${foreign_disk}"
sudo vgcreate foreign_lvmvg "${foreign_disk}" --config="${FOREIGN_LVM_CONFIG}"

# setup a RAID volume group
cleanup_raidvg
raid_disks=()
for IMG in `seq ${RAID_COUNT}`
do
  truncate -s 1024G /tmp/openebs_ci_raid_disk_${IMG}.img
  raid_disk="$(sudo losetup -f /tmp/openebs_ci_raid_disk_${IMG}.img --show)"
  sudo pvcreate "${raid_disk}"

  raid_disks+=("${raid_disk}")
done
sudo vgcreate raidvg "${raid_disks[@]}"

# setup a RAID volume group
cleanup_raidvg
raid_disks=()
for IMG in `seq ${RAID_COUNT}`
do
  truncate -s 1024G /tmp/openebs_ci_raid_disk_${IMG}.img
  raid_disk="$(sudo losetup -f /tmp/openebs_ci_raid_disk_${IMG}.img --show)"
  sudo pvcreate "${raid_disk}"

  raid_disks+=("${raid_disk}")
done
sudo vgcreate raidvg "${raid_disks[@]}"

# install snapshot and thin volume module for lvm
sudo modprobe dm-snapshot
sudo modprobe dm_thin_pool

# install RAID modules for lvm
sudo modprobe dm_raid
sudo modprobe dm_integrity

# Prepare env for running BDD tests
# Minikube is already running
kubectl apply -f "${LVM_OPERATOR}"
kubectl apply -f "${SNAP_CLASS}"

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

if ! ginkgo -v ; then

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

# last statement formatted to always return true
[ -z "${CLEANUP}" ] || cleanup 2>/dev/null
