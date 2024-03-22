### *Prerequisite*

Please do not provision/deprovision any volumes during the upgrade, if we can not control it, then we can scale down the openebs-lvm-controller stateful set to zero replica which will pause all the provisioning/deprovisioning request. And once upgrade is done, we can scale up the controller pod and then volume provisioning/deprovisioning will resume on the upgraded system.

```
$ kubectl edit deploy openebs-lvm-controller -n kube-system

```
And set replicas to zero :

```
spec:
  podManagementPolicy: OrderedReady
    *replicas: 0*
      revisionHistoryLimit: 10
```

### *upgrade the driver*

We can upgrade the lvm driver to the latest stable release version by apply the following command:

```
$ kubectl apply -f https://openebs.github.io/charts/lvm-operator.yaml
```

Please note that if you were using the LVM_NAMESPACE env value other than `openebs` (default value) in which lvm-localpv CR's are created, don't forget to update that value in lvm-operator yaml file under LVM_NAMESPACE env.

For upgrading the driver to any particular release, download the lvm-operator yaml from the desired branch and update the lvm-driver image tag to the corresponding tag. For e.g, to upgrade the lvm-driver to 0.7.0 version, follow these steps:

1. Download operator yaml from specific branch
```
wget https://raw.githubusercontent.com/openebs/lvm-localpv/v0.7.x/deploy/lvm-operator.yaml
```

2. Update the lvm-driver image tag. We have to update this at two places,

one at `openebs-lvm-plugin` container image in lvm-controller deployment
```
        - name: openebs-lvm-plugin
          image: openebs/lvm-driver:ci  // update it to openebs/lvm-driver:0.7.0
          imagePullPolicy: IfNotPresent
          env:
            - name: OPENEBS_CONTROLLER_DRIVER
              value: controller
            - name: OPENEBS_CSI_ENDPOINT
              value: unix:///var/lib/csi/sockets/pluginproxy/csi.sock
            - name: LVM_NAMESPACE
              value: openebs
```
and other one at `openebs-lvm-plugin` container in lvm-node daemonset.
```
        - name: openebs-lvm-plugin
          securityContext:
            privileged: true
            allowPrivilegeEscalation: true
          image: openebs/lvm-driver:ci   // Update it to openebs/lvm-driver:0.7.0
          imagePullPolicy: IfNotPresent
          args:
            - "--nodeid=$(OPENEBS_NODE_ID)"
            - "--endpoint=$(OPENEBS_CSI_ENDPOINT)"
            - "--plugin=$(OPENEBS_NODE_DRIVER)"
            - "--listen-address=$(METRICS_LISTEN_ADDRESS)"
```

3. If you were using lvm-controller in high-availability (HA) mode, make sure to update deployment replicas. By default it is set to one (1).

```
spec:
  selector:
    matchLabels:
      app: openebs-lvm-controller
      role: openebs-lvm
  serviceName: "openebs-lvm"
  replicas: 1                     // update it to desired lvm-controller replicas.
```

4. Now we can apply the lvm-operator.yaml file to upgrade lvm-driver to 0.7.0 version.

### *Note*

While upgrading lvm-driver from v0.8.0 to later version by applying lvm-operator file, we may get this error.
```
The CSIDriver "local.csi.openebs.io" is invalid: spec.storageCapacity: Invalid value: true: field is immutable
```
It occurs due to newly added field `storageCapacity: true` in csi driver spec. In that case, first delete the csi-driver by running this command:
```
$ kubectl delete csidriver local.csi.openebs.io 
```
Now we can again apply the operator yaml file.
