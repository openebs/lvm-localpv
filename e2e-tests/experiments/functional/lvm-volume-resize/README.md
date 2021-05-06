## About this experiment

This experiment verifies the volume resize feature of lvm-localpv. For resize the volume we just need to update the pvc yaml with desired size and apply it. We can directly edit the pvc by ```kubectl edit pvc <pvc_name> -n <namespace>``` command and update the spec.resources.requests.storage field with desired volume size. One thing need to be noted that volume resize can only be done from lower pvc size to higher pvc size. We can not resize the volume from higher pvc size to lower one, in-short volume shrink is not possible. lvm driver supports online volume expansion, so that for using the resized volume, application pod restart is not required. For resize, storage-class which will provision the pvc should have `allowVolumeExpansion: true` field.

for e.g.
```
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: openebs-lvmsc
allowVolumeExpansion: true
parameters:
  volgroup: "lvmvg"
provisioner: local.csi.openebs.io
```

## Supported platforms:

K8S : 1.18+

OS : Ubuntu

LVM version : LVM 2

## Entry-criteria

- K8s cluster should be in healthy state including all the nodes in ready state.
- lvm-controller and csi node-agent daemonset pods should be in running state.
- storage class with `allowVolumeExpansion: true` enable should be present.
- Application should be deployed succesfully consuming the lvm-localpv storage.

## Exit-criteria

- Volume should be resized successfully and application should be accessible seamlessly.
- Application should be able to use the new resize volume space.

## How to run

- This experiment accepts the parameters in form of kubernetes job environmental variables.
- For running this experiment of lvm volume resize, clone openens/lvm-localpv[https://github.com/openebs/lvm-localpv] repo and then first apply rbac and crds for e2e-framework.
```
kubectl apply -f lvm-localpv/e2e-tests/hack/rbac.yaml
kubectl apply -f lvm-localpv/e2e-tests/hack/crds.yaml
```
then update the needed test specific values in run_e2e_test.yml file and create the kubernetes job.
```
kubectl create -f run_e2e_test.yml
```
All the env variables description is provided with the comments in the same file.
