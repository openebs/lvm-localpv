## About the experiment

- This functional test verifies the lvm-localpv shared mount volume support via multiple pods. Applications who wants to share the volume can use the storage-class as below.

```
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: lvmsc-shared
parameters:
  shared: "yes"
  volgroup: "< volume_group name >"
provisioner: local.csi.openebs.io
```

## Supported platforms:

K8s : 1.18+

OS : Ubuntu

LVM version : LVM 2

## Entry-Criteria

- K8s cluster should be in healthy state including all desired nodes in ready state.
- lvm-controller and csi node-agent daemonset pods should be in running state.

## Steps performed in this experiment:

1. First deploy the busybox application using `shared: yes` enabled storage-class
2. Then we dump some dummy data into the application pod mount point.
3. Scale the busybox deployment replicas so that multiple pods (here replicas = 2) can share the volume.
4. After that data consistency is verified from the scaled application pod in the way that data is accessible from both the pods and after restarting the application pod data consistency should be maintained.

## How to run

- This experiment accepts the parameters in form of kubernetes job environmental variables.
- For running this experiment of lvm-localpv shared mount volume, first clone [openebs/lvm-localpv](https://github.com/openebs/lvm-localpv) repo and then apply rbac and crds for e2e-framework.
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
kubectl logs -f <lvmpv-shared-mount-volume-xxxxx-xxxxx> -n e2e
```
To get the test-case result, get the corresponding e2e custom-resource `e2eresult` (short name: e2er ) and check its phase (Running or Completed) and result (Pass or Fail).

```
kubectl get e2er
kubectl get e2er lvmpv-shared-mount-volume -n e2e --no-headers -o custom-columns=:.spec.testStatus.phase
kubectl get e2er lvmpv-shared-mount-volume -n e2e --no-headers -o custom-columns=:.spec.testStatus.result
```