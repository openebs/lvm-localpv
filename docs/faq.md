### 1. How to add custom topology key

To add custom topology key, we can label all the nodes with the required key and value :-

```sh
$ kubectl label node k8s-node-1 openebs.io/rack=rack1
node/k8s-node-1 labeled

$ kubectl get nodes k8s-node-1 --show-labels
NAME           STATUS   ROLES    AGE   VERSION   LABELS
k8s-node-1   Ready    worker   16d   v1.17.4   beta.kubernetes.io/arch=amd64,beta.kubernetes.io/os=linux,kubernetes.io/arch=amd64,kubernetes.io/hostname=k8s-node-1,kubernetes.io/os=linux,node-role.kubernetes.io/worker=true,openebs.io/rack=rack1

```
It is recommended is to label all the nodes with the same key, they can have different values for the given keys, but all keys should be present on all the worker node.

Once we have labeled the node, we can install the lvm driver. The driver will pick the node labels and add that as the supported topology key. If the driver is already installed and you want to add a new topology information, you can label the node with the topology information and then restart the LVM-LocalPV CSI driver daemon sets (openebs-lvm-node) are required so that the driver can pick the labels and add them as supported topology keys. We should restart the pod in kube-system namespace with the name as openebs-lvm-node-[xxxxx] which is the node agent pod for the LVM-LocalPV Driver.

Note that restart of LVM LocalPV CSI driver daemon sets are must in case, if we are going to use WaitForFirstConsumer as volumeBindingMode in storage class. In case of immediate volume binding mode, restart of daemon set is not a must requirement, irrespective of sequence of labelling the node either prior to install lvm driver or after install. However it is recommended to restart the daemon set if we are labeling the nodes after the installation.

```sh
$ kubectl get pods -n kube-system -l role=openebs-lvm

NAME                       READY   STATUS    RESTARTS   AGE
openebs-lvm-controller-0   4/4     Running   0          5h28m
openebs-lvm-node-4d94n     2/2     Running   0          5h28m
openebs-lvm-node-gssh8     2/2     Running   0          5h28m
openebs-lvm-node-twmx8     2/2     Running   0          5h28m
```

We can verify that key has been registered successfully with the LVM LocalPV CSI Driver by checking the CSI node object yaml :-

```yaml
$ kubectl get csinodes pawan-node-1 -oyaml
apiVersion: storage.k8s.io/v1
kind: CSINode
metadata:
  creationTimestamp: "2020-04-13T14:49:59Z"
  name: k8s-node-1
  ownerReferences:
  - apiVersion: v1
    kind: Node
    name: k8s-node-1
    uid: fe268f4b-d9a9-490a-a999-8cde20c4dadb
  resourceVersion: "4586341"
  selfLink: /apis/storage.k8s.io/v1/csinodes/k8s-node-1
  uid: 522c2110-9d75-4bca-9879-098eb8b44e5d
spec:
  drivers:
  - name: local.csi.openebs.io
    nodeID: k8s-node-1
    topologyKeys:
    - beta.kubernetes.io/arch
    - beta.kubernetes.io/os
    - kubernetes.io/arch
    - kubernetes.io/hostname
    - kubernetes.io/os
    - node-role.kubernetes.io/worker
    - openebs.io/rack
```

We can see that "openebs.io/rack" is listed as topology key. Now we can create a storageclass with the topology key created :

```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: openebs-lvmpv
allowVolumeExpansion: true
parameters:
  volgroup: "lvmvg"
provisioner: local.csi.openebs.io
allowedTopologies:
- matchLabelExpressions:
  - key: openebs.io/rack
    values:
      - rack1
```

The LVM LocalPV CSI driver will schedule the PV to the nodes where label "openebs.io/rack" is set to "rack1".

Note that if storageclass is using Immediate binding mode and topology key is not mentioned then all the nodes should be labeled using same key, that means, same key should be present on all nodes, nodes can have different values for those keys. If nodes are labeled with different keys i.e. some nodes are having different keys, then LVMPV's default scheduler can not effectively do the volume capacity based scheduling. Here, in this case the CSI provisioner will pick keys from any random node and then prepare the preferred topology list using the nodes which has those keys defined and LVMPV scheduler will schedule the PV among those nodes only.
