# e2e-tests

e2e-tests is a suite of functional and chaos test scenarios for workloads which are consuming lvm-localpv storage. Our vision includes enabling end users to easily execute chaos experiments in their environments using a Kubernetes native approach, where each test scenario is specified in a declarative way. The primary objective of e2e-tests is to ensure a consistent and reliable behavior of workloads running on lvm-localpv storage in Kubernetes. 

The test logic is packaged into dedicated containers which makes them portable across Kubernetes deployments. 
This containerization also helps to integrate e2e-tests into CI/CD environments. 


## Getting Started

e2e-tests experiment jobs run using a dedicated ServiceAccount in the e2e namespace. So first of all clone this repository and head to `e2e-tests/hack` folder to setup RBAC & custom resource definitions (CRDs) via kubectl, as shown below: 

```
git clone https://github.com/openebs/lvm-localpv.git
cd lvm-localpv/e2e-tests
kubectl apply -f hack/rbac.yaml
kubectl apply -f hack/crds.yaml  
```


## Directory structure

```
├── e2e-tests
    ├── apps         # test scripts for provision/deprovision application deployment/statefulset. 
    ├── chaoslib     # collection of chaos utils, used in test scripts for some specific chaos/infra-chaos
    ├── experiments  # functional,chaos and infra-chaos e2e-tests playbooks.
    ├── hack         # e2e-framework setup and test-result updation related yaml files.
    ├── utils        # general use-case playbooks for k8s objects creation and check their status
```

- All the functional and chaos test scenarios are located inside `experiments` folder. Each test experiment in-general will have the below directory structure.

```
├── data_persistence.j2   # to call data-consistency check util, based on application used
├── README.md             # description about the experiment scenario.
├── run_e2e_test.yml   # kubernetes job spec for running e2e-tests
├── test_vars.yml         # variables used in ansible-playbook for e2e-tests.
└── test.yml              # actual e2e-test logic in the form of ansible playbook.
```


## How to run experiments 

Let's say, you'd like to test resiliency of a stateful application pod upon container crash. If you already have some application you can run this experiment directly, else you can first deploy one test application from apps directory.

- Locate the Experiment: experiments are typically placed in `experiments/<type>` folders. In this case, the corresponding e2e-test is present at `experiments/chaos/app_pod_failure` 

- Update the application information (generally, the namespace and app labels) & other test-specific information being passed as ENVs to the e2e-test job (`run_e2e_test.yml`). 

- Run the e2e-test:

  ```
  kubectl create -f experiments/chaos/app_pod_failure/run_e2e_test.yml
  ```
  

## Get Experiment Results

- After creating kubernetes job, when the job’s pod is instantiated, we can see the logs of that pod which is executing the test-case.

```
kubectl get pods -n e2e
kubectl logs -f <application-pod-failure-xxxxx-xxxxx> -n e2e
```

Results are maintained in a custom resource (`e2eresult`) that bears the same name as the experiment. In this case,
`application-pod-failure`. To get the test-case result, get the corresponding e2e custom-resource `e2eresult` (short name: e2er ) and check its phase (Running or Completed) and result (Pass or Fail).

```
kubectl get e2er
kubectl get e2er application-pod-failure -n e2e --no-headers -o custom-columns=:.spec.testStatus.phase
kubectl get e2er application-pod-failure -n e2e --no-headers -o custom-columns=:.spec.testStatus.result
```


## Test case modification/updation as per need

- Before performing any e2e-test read `steps performed`,`entry and exit criteria` section within `README.md` files. If you want to modify the test-logic or required more env's as per your need, you can modify test.yml or run_e2e_test file and then build the docker image and push it to some repository (from where you can pull images in your cluster environment). Update the container image name in run_e2e_test.yml file and take very care of ImagePullPolicy. Run the test with your changes.

- If you find that your changes are generic and can be upstreamed as well, or you have idea with some new test-cases for lvm-localpv e2e, please raise a Pull Request with your changes and don't forget to attach ansible-playbooks logs for the updated e2e-tests. You can raise a issue as well if find any misbehaviour regarding test-cases. 