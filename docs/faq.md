### 1. How to add custom topology key

To add custom topology key:
* Label the nodes with the required key and value.
* Set env variables in the LVM driver daemonset yaml(openebs-lvm-node), if already deployed, you can edit the daemonSet directly.
* "openebs.io/nodename" has been added as default topology key. 
* Create storageclass with above specific labels keys.


```sh
$ kubectl label node k8s-node-1 openebs.io/rack=rack1
node/k8s-node-1 labeled

$ kubectl get nodes k8s-node-1 --show-labels
NAME           STATUS   ROLES    AGE   VERSION   LABELS
k8s-node-1   Ready    worker   16d   v1.17.4   beta.kubernetes.io/arch=amd64,beta.kubernetes.io/os=linux,kubernetes.io/arch=amd64,kubernetes.io/hostname=k8s-node-1,kubernetes.io/os=linux,node-role.kubernetes.io/worker=true,openebs.io/rack=rack1


$ kubectl get ds -n kube-system openebs-lvm-node -o yaml
...
env:
  - name: OPENEBS_NODE_ID
    valueFrom:
      fieldRef:
        fieldPath: spec.nodeName
  - name: OPENEBS_CSI_ENDPOINT
    value: unix:///plugin/csi.sock
  - name: OPENEBS_NODE_DRIVER
    value: agent
  - name: LVM_NAMESPACE
    value: openebs
  - name: ALLOWED_TOPOLOGIES
    value: "openebs.io/rack"

```
It is recommended is to label all the nodes with the same key, they can have different values for the given keys, but all keys should be present on all the worker node.

Once we have labeled the node, we can install the lvm driver. The driver will pick the keys from env "ALLOWED_TOPOLOGIES" and add that as the supported topology key. If the driver is already installed and you want to add a new topology information, you can edit the LVM-LocalPV CSI driver daemon sets (openebs-lvm-node).


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
    - openebs.io/nodename
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

Note that if storageclass is using Immediate binding mode and storageclass allowedTopologies is not mentioned then all the nodes should be labeled using "ALLOWED_TOPOLOGIES" keys, that means, "ALLOWED_TOPOLOGIES" keys should be present on all nodes, nodes can have different values for those keys. If some nodes don't have those keys, then LVMPV's default scheduler can not effectively do the volume capacity based scheduling. Here, in this case the CSI provisioner will pick keys from any random node and then prepare the preferred topology list using the nodes which has those keys defined and LVMPV scheduler will schedule the PV among those nodes only.
