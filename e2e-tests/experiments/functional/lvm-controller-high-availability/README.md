## About this experiment

This functional experiment scale up the lvm-controller replicas to use it in high availability mode and then verify the lvm-localpv behaviour when one of the replicas go down. This experiment checks the initial number of replicas of lvm-controller and scale it by one if a free node is present which should be able to schedule the pods. Default value for lvm-controller deployment replica is one.

## Supported platforms:

K8s : 1.18+

OS : Ubuntu

LVM version: LVM 2

## Entry-Criteria

- k8s cluster should be in healthy state including all the nodes in ready state.
- lvm-controller and csi node-agent daemonset pods should be in running state.

## Exit-Criteria

- lvm-controller deployment should be scaled up by one replica.
- All the replias should be in running state.
- lvm-localpv volumes should be healthy and data after scaling up controller should not be impacted.
- This experiment makes one of the lvm-controller deployment replica to go down, as a result active/master replica of lvm-controller prior to the experiment will be changed to some other remaining replica after the experiment completes. This happens because of the lease mechanism, which is being used to decide which replica will be serving as master. At a time only one replica will be master and other replica will follow the anti-affinity rules so that these replica pods will be present on different nodes only.
- Volumes provisioning / deprovisioning should not be impacted if any one replica goes down.

## How to run

- This experiment accepts the parameters in form of kubernetes job environmental variables.
- For running this experiment of deploying lvm-localpv provisioner, clone openens/lvm-localpv[https://github.com/openebs/lvm-localpv] repo and then first apply rbac and crds for e2e-framework.
```
kubectl apply -f lvm-localpv/e2e-tests/hack/rbac.yaml
kubectl apply -f lvm-localpv/e2e-tests/hack/crds.yaml
```
then update the needed test specific values in run_e2e_test.yml file and create the kubernetes job.
```
kubectl create -f run_e2e_test.yml
```
All the env variables description is provided with the comments in the same file.