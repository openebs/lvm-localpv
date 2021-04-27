// Copyright © 2018-2019 The OpenEBS Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pv

import (
	"context"
	"strings"

	"github.com/openebs/lib-csi/pkg/common/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/openebs/lib-csi/pkg/common/kubernetes/client"
	"k8s.io/client-go/kubernetes"
)

const (
	//KeyNode represents the key values used for specifying the Node Affinity
	// based on the hostname
	KeyNode = "kubernetes.io/hostname"
)

// getClientsetFn is a typed function that
// abstracts fetching of clientset
type getClientsetFn func() (clientset *kubernetes.Clientset, err error)

// getClientsetFromPathFn is a typed function that
// abstracts fetching of clientset from kubeConfigPath
type getClientsetForPathFn func(kubeConfigPath string) (clientset *kubernetes.Clientset, err error)

// getpvcFn is a typed function that
// abstracts fetching of pv
type getFn func(cli *kubernetes.Clientset, name string, opts metav1.GetOptions) (*corev1.PersistentVolume, error)

// listFn is a typed function that abstracts
// listing of pvs
type listFn func(cli *kubernetes.Clientset, opts metav1.ListOptions) (*corev1.PersistentVolumeList, error)

// deleteFn is a typed function that abstracts
// deletion of pvs
type deleteFn func(cli *kubernetes.Clientset, name string, deleteOpts *metav1.DeleteOptions) error

// deleteFn is a typed function that abstracts
// deletion of pv's collection
type deleteCollectionFn func(cli *kubernetes.Clientset, listOpts metav1.ListOptions, deleteOpts *metav1.DeleteOptions) error

// createFn is a typed function that abstracts
// creation of pv
type createFn func(cli *kubernetes.Clientset, pv *corev1.PersistentVolume) (*corev1.PersistentVolume, error)

// Kubeclient enables kubernetes API operations
// on pv instance
type Kubeclient struct {
	// clientset refers to pvc clientset
	// that will be responsible to
	// make kubernetes API calls
	clientset *kubernetes.Clientset

	// kubeconfig path to get kubernetes clientset
	kubeConfigPath string

	// functions useful during mocking
	getClientset        getClientsetFn
	getClientsetForPath getClientsetForPathFn
	list                listFn
	get                 getFn
	create              createFn
	del                 deleteFn
	delCollection       deleteCollectionFn
}

// KubeclientBuildOption abstracts creating an
// instance of kubeclient
type KubeclientBuildOption func(*Kubeclient)

// withDefaults sets the default options
// of kubeclient instance
func (k *Kubeclient) withDefaults() {
	if k.getClientset == nil {
		k.getClientset = func() (clients *kubernetes.Clientset, err error) {
			return client.New().Clientset()
		}
	}
	if k.getClientsetForPath == nil {
		k.getClientsetForPath = func(kubeConfigPath string) (clients *kubernetes.Clientset, err error) {
			return client.New(client.WithKubeConfigPath(kubeConfigPath)).Clientset()
		}
	}
	if k.get == nil {
		k.get = func(cli *kubernetes.Clientset, name string, opts metav1.GetOptions) (*corev1.PersistentVolume, error) {
			return cli.CoreV1().PersistentVolumes().Get(context.TODO(), name, opts)
		}
	}
	if k.list == nil {
		k.list = func(cli *kubernetes.Clientset, opts metav1.ListOptions) (*corev1.PersistentVolumeList, error) {
			return cli.CoreV1().PersistentVolumes().List(context.TODO(), opts)
		}
	}
	if k.del == nil {
		k.del = func(cli *kubernetes.Clientset, name string, deleteOpts *metav1.DeleteOptions) error {
			return cli.CoreV1().PersistentVolumes().Delete(context.TODO(), name, *deleteOpts)
		}
	}
	if k.delCollection == nil {
		k.delCollection = func(cli *kubernetes.Clientset, listOpts metav1.ListOptions, deleteOpts *metav1.DeleteOptions) error {
			return cli.CoreV1().PersistentVolumes().DeleteCollection(context.TODO(), *deleteOpts, listOpts)
		}
	}
	if k.create == nil {
		k.create = func(cli *kubernetes.Clientset, pv *corev1.PersistentVolume) (*corev1.PersistentVolume, error) {
			return cli.CoreV1().PersistentVolumes().Create(context.TODO(), pv, metav1.CreateOptions{})
		}
	}
}

// WithClientSet sets the kubernetes client against
// the kubeclient instance
func WithClientSet(c *kubernetes.Clientset) KubeclientBuildOption {
	return func(k *Kubeclient) {
		k.clientset = c
	}
}

// WithKubeConfigPath sets the kubeConfig path
// against client instance
func WithKubeConfigPath(path string) KubeclientBuildOption {
	return func(k *Kubeclient) {
		k.kubeConfigPath = path
	}
}

// NewKubeClient returns a new instance of kubeclient meant for
// cstor volume replica operations
func NewKubeClient(opts ...KubeclientBuildOption) *Kubeclient {
	k := &Kubeclient{}
	for _, o := range opts {
		o(k)
	}
	k.withDefaults()
	return k
}

func (k *Kubeclient) getClientsetForPathOrDirect() (*kubernetes.Clientset, error) {
	if k.kubeConfigPath != "" {
		return k.getClientsetForPath(k.kubeConfigPath)
	}
	return k.getClientset()
}

// getClientsetOrCached returns either a new instance
// of kubernetes client or its cached copy
func (k *Kubeclient) getClientsetOrCached() (*kubernetes.Clientset, error) {
	if k.clientset != nil {
		return k.clientset, nil
	}

	cs, err := k.getClientsetForPathOrDirect()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get clientset")
	}
	k.clientset = cs
	return k.clientset, nil
}

// Get returns a pv resource
// instances present in kubernetes cluster
func (k *Kubeclient) Get(name string, opts metav1.GetOptions) (*corev1.PersistentVolume, error) {
	if strings.TrimSpace(name) == "" {
		return nil, errors.New("failed to get pv: missing pv name")
	}
	cli, err := k.getClientsetOrCached()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get pv {%s}", name)
	}
	return k.get(cli, name, opts)
}

// List returns a list of pv
// instances present in kubernetes cluster
func (k *Kubeclient) List(opts metav1.ListOptions) (*corev1.PersistentVolumeList, error) {
	cli, err := k.getClientsetOrCached()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to list pv listoptions: '%v'", opts)
	}
	return k.list(cli, opts)
}

// Delete deletes a pv instance from the
// kubecrnetes cluster
func (k *Kubeclient) Delete(name string, deleteOpts *metav1.DeleteOptions) error {
	if strings.TrimSpace(name) == "" {
		return errors.New("failed to delete pvc: missing pv name")
	}
	cli, err := k.getClientsetOrCached()
	if err != nil {
		return errors.Wrapf(err, "failed to delete pv {%s}", name)
	}
	return k.del(cli, name, deleteOpts)
}

// Create creates a pv in kubernetes cluster
func (k *Kubeclient) Create(pv *corev1.PersistentVolume) (*corev1.PersistentVolume, error) {
	cli, err := k.getClientsetOrCached()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create pv: %s", pv.Name)
	}
	return k.create(cli, pv)
}

// DeleteCollection deletes a collection of pv objects.
func (k *Kubeclient) DeleteCollection(listOpts metav1.ListOptions, deleteOpts *metav1.DeleteOptions) error {
	cli, err := k.getClientsetOrCached()
	if err != nil {
		return errors.Wrapf(err, "failed to delete the collection of pvs")
	}
	return k.delCollection(cli, listOpts, deleteOpts)
}
