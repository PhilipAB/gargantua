/*
Copyright The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by client-gen. DO NOT EDIT.

package v1

import (
	"time"

	v1 "github.com/hobbyfarm/gargantua/pkg/apis/hobbyfarm.io/v1"
	scheme "github.com/hobbyfarm/gargantua/pkg/client/clientset/versioned/scheme"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// VirtualMachineTypesGetter has a method to return a VirtualMachineTypeInterface.
// A group's client should implement this interface.
type VirtualMachineTypesGetter interface {
	VirtualMachineTypes(namespace string) VirtualMachineTypeInterface
}

// VirtualMachineTypeInterface has methods to work with VirtualMachineType resources.
type VirtualMachineTypeInterface interface {
	Create(*v1.VirtualMachineType) (*v1.VirtualMachineType, error)
	Update(*v1.VirtualMachineType) (*v1.VirtualMachineType, error)
	Delete(name string, options *metav1.DeleteOptions) error
	DeleteCollection(options *metav1.DeleteOptions, listOptions metav1.ListOptions) error
	Get(name string, options metav1.GetOptions) (*v1.VirtualMachineType, error)
	List(opts metav1.ListOptions) (*v1.VirtualMachineTypeList, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.VirtualMachineType, err error)
	VirtualMachineTypeExpansion
}

// virtualMachineTypes implements VirtualMachineTypeInterface
type virtualMachineTypes struct {
	client rest.Interface
	ns     string
}

// newVirtualMachineTypes returns a VirtualMachineTypes
func newVirtualMachineTypes(c *HobbyfarmV1Client, namespace string) *virtualMachineTypes {
	return &virtualMachineTypes{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the virtualMachineType, and returns the corresponding virtualMachineType object, and an error if there is any.
func (c *virtualMachineTypes) Get(name string, options metav1.GetOptions) (result *v1.VirtualMachineType, err error) {
	result = &v1.VirtualMachineType{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("virtualmachinetypes").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of VirtualMachineTypes that match those selectors.
func (c *virtualMachineTypes) List(opts metav1.ListOptions) (result *v1.VirtualMachineTypeList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1.VirtualMachineTypeList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("virtualmachinetypes").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested virtualMachineTypes.
func (c *virtualMachineTypes) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("virtualmachinetypes").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch()
}

// Create takes the representation of a virtualMachineType and creates it.  Returns the server's representation of the virtualMachineType, and an error, if there is any.
func (c *virtualMachineTypes) Create(virtualMachineType *v1.VirtualMachineType) (result *v1.VirtualMachineType, err error) {
	result = &v1.VirtualMachineType{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("virtualmachinetypes").
		Body(virtualMachineType).
		Do().
		Into(result)
	return
}

// Update takes the representation of a virtualMachineType and updates it. Returns the server's representation of the virtualMachineType, and an error, if there is any.
func (c *virtualMachineTypes) Update(virtualMachineType *v1.VirtualMachineType) (result *v1.VirtualMachineType, err error) {
	result = &v1.VirtualMachineType{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("virtualmachinetypes").
		Name(virtualMachineType.Name).
		Body(virtualMachineType).
		Do().
		Into(result)
	return
}

// Delete takes name of the virtualMachineType and deletes it. Returns an error if one occurs.
func (c *virtualMachineTypes) Delete(name string, options *metav1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("virtualmachinetypes").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *virtualMachineTypes) DeleteCollection(options *metav1.DeleteOptions, listOptions metav1.ListOptions) error {
	var timeout time.Duration
	if listOptions.TimeoutSeconds != nil {
		timeout = time.Duration(*listOptions.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("virtualmachinetypes").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Timeout(timeout).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched virtualMachineType.
func (c *virtualMachineTypes) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.VirtualMachineType, err error) {
	result = &v1.VirtualMachineType{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("virtualmachinetypes").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}