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

// Code generated by lister-gen. DO NOT EDIT.

package v1

import (
	v1 "github.com/hobbyfarm/gargantua/pkg/apis/hobbyfarm.io/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// ActiveScenarioLister helps list ActiveScenarios.
type ActiveScenarioLister interface {
	// List lists all ActiveScenarios in the indexer.
	List(selector labels.Selector) (ret []*v1.ActiveScenario, err error)
	// ActiveScenarios returns an object that can list and get ActiveScenarios.
	ActiveScenarios(namespace string) ActiveScenarioNamespaceLister
	ActiveScenarioListerExpansion
}

// activeScenarioLister implements the ActiveScenarioLister interface.
type activeScenarioLister struct {
	indexer cache.Indexer
}

// NewActiveScenarioLister returns a new ActiveScenarioLister.
func NewActiveScenarioLister(indexer cache.Indexer) ActiveScenarioLister {
	return &activeScenarioLister{indexer: indexer}
}

// List lists all ActiveScenarios in the indexer.
func (s *activeScenarioLister) List(selector labels.Selector) (ret []*v1.ActiveScenario, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1.ActiveScenario))
	})
	return ret, err
}

// ActiveScenarios returns an object that can list and get ActiveScenarios.
func (s *activeScenarioLister) ActiveScenarios(namespace string) ActiveScenarioNamespaceLister {
	return activeScenarioNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// ActiveScenarioNamespaceLister helps list and get ActiveScenarios.
type ActiveScenarioNamespaceLister interface {
	// List lists all ActiveScenarios in the indexer for a given namespace.
	List(selector labels.Selector) (ret []*v1.ActiveScenario, err error)
	// Get retrieves the ActiveScenario from the indexer for a given namespace and name.
	Get(name string) (*v1.ActiveScenario, error)
	ActiveScenarioNamespaceListerExpansion
}

// activeScenarioNamespaceLister implements the ActiveScenarioNamespaceLister
// interface.
type activeScenarioNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all ActiveScenarios in the indexer for a given namespace.
func (s activeScenarioNamespaceLister) List(selector labels.Selector) (ret []*v1.ActiveScenario, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1.ActiveScenario))
	})
	return ret, err
}

// Get retrieves the ActiveScenario from the indexer for a given namespace and name.
func (s activeScenarioNamespaceLister) Get(name string) (*v1.ActiveScenario, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1.Resource("activescenario"), name)
	}
	return obj.(*v1.ActiveScenario), nil
}