## About this experiment

This experiment verifies the provision and deprovision of raw block volumes by lvm-localpv. There are some specialized applications that require direct access to a block device because, for example, the file system layer introduces unneeded overhead. The most common case is databases, which prefer to organize their data directly on the underlying storage. In this experiment we are not using any such application for testing, but using a simple busybox application to verify successful provisioning and deprovisioning of raw block volume.

To provisione the Raw Block volume, we should create a storageclass without any fstype as Raw block volume does not have any fstype.

```
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: openebs-lvmsc-raw-block
allowVolumeExpansion: true
parameters:
  volgroup: "{{ vg_name }}"
provisioner: local.csi.openebs.io
```    
Note: For running this experiment above storage-class should be present. This storage class will be created as a part of lvm-localpv provisioner experiment. If lvm-localpv components are not deployed using e2e-test script located at `openebs/lvm-localpv/e2e-tests/experiment/lvm-localpv-provisioiner` please make sure you create the storage class from above mentioned yaml.

## Supported platforms:

K8s : 1.18+

OS : Ubuntu, CentOS

lvm : 0.7, 0.8

## Entry-Criteria

- K8s cluster should be in healthy state including all desired nodes in ready state.
- lvm-controller and node-agent daemonset pods should be in running state.
- storage class without any fstype should be present.
- a directory should be present on node with name `raw_block_volume`.

## Steps performed

- deploy the busybox application with given a devicePath.
- verify that application pvc gets bound and application pod is in running state.
- dump some data into raw block device and take the md5sum of data.
- restart the application and verify the data consistency.
- After that update the pvc with double value of previous pvc size, to validate resize support for raw block volumes.
- when resize is successful, then dump some dummy data into application to use the resized space.
- At last deprovision the application and check its successful deletion.

## How to run

- This experiment accepts the parameters in form of kubernetes job environmental variables.
- For running this experiment of lvmpv raw block volume creation, first clone openens/lvm-localpv[https://github.com/openebs/lvm-localpv] repo and then apply rbac and crds for e2e-framework.
```
kubectl apply -f lvm-localpv/e2e-tests/hack/rbac.yaml
kubectl apply -f lvm-localpv/e2e-tests/hack/crds.yaml
```
then update the needed test specific values in run_e2e_test.yml file and create the kubernetes job.
```
kubectl create -f run_e2e_test.yml
```
All the env variables description is provided with the comments in the same file.
After creating kubernetes job, when the job’s pod is instantiated, we can see the logs of that pod which is executing the test-case.

```
kubectl get pods -n e2e
kubectl logs -f <lvmpv-block-volume-xxxxx-xxxxx> -n e2e
```
To get the test-case result, get the corresponding e2e custom-resource `e2eresult` (short name: e2er ) and check its phase (Running or Completed) and result (Pass or Fail).

```
kubectl get e2er
kubectl get e2er lvmpv-raw-block-volume -n e2e --no-headers -o custom-columns=:.spec.testStatus.phase
kubectl get e2er lvmpv-raw-block-volume -n e2e --no-headers -o custom-columns=:.spec.testStatus.result
```