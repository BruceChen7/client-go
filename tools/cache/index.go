/*
Copyright 2014 The Kubernetes Authors.

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

package cache

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/util/sets"
)

// Indexer extends Store with multiple indices and restricts each
// accumulator to simply hold the current object (and be empty after
// Delete).
//
// There are three kinds of strings here:
// 1. a storage key, as defined in the Store interface,
// 2. a name of an index, and
// 3. an "indexed value", which is produced by an IndexFunc and
//    can be a field value or any other string computed from the object.
type Indexer interface {
	// 组合Store接口，在store存储的基础上，扩展了对象查找的能力
	Store
	// Index returns the stored objects whose set of indexed values
	// intersects the set of indexed values of the given object, for
	// the named index
	// IndexName是索引函数名，obj是对象，
	Index(indexName string, obj interface{}) ([]interface{}, error)
	// IndexKeys returns the storage keys of the stored objects whose
	// set of indexed values for the named index includes the given
	// indexed value
	IndexKeys(indexName, indexedValue string) ([]string, error)
	// 列举索引函数indexName 对应的所有对象键
	// ListIndexFuncValues returns all the indexed values of the given index
	ListIndexFuncValues(indexName string) []string
	// ByIndex returns the stored objects whose set of indexed values
	// for the named index includes the given indexed value
	ByIndex(indexName, indexedValue string) ([]interface{}, error)
	// GetIndexer return the indexers
	GetIndexers() Indexers

	// AddIndexers adds more indexers to this store.  If you call this after you already have data
	// in the store, the results are undefined.
	AddIndexers(newIndexers Indexers) error
}

// IndexFunc knows how to compute the set of indexed values for an object.
// 索引函数，输入一个对象，输出是该对象对应的索引链表
type IndexFunc func(obj interface{}) ([]string, error)

// IndexFuncToKeyFuncAdapter adapts an indexFunc to a keyFunc.  This is only useful if your index function returns
// unique values for every object.  This conversion can create errors when more than one key is found.  You
// should prefer to make proper key and index functions.
func IndexFuncToKeyFuncAdapter(indexFunc IndexFunc) KeyFunc {
	return func(obj interface{}) (string, error) {
		indexKeys, err := indexFunc(obj)
		if err != nil {
			return "", err
		}
		if len(indexKeys) > 1 {
			return "", fmt.Errorf("too many keys: %v", indexKeys)
		}
		if len(indexKeys) == 0 {
			return "", fmt.Errorf("unexpected empty indexKeys")
		}
		return indexKeys[0], nil
	}
}

const (
	// NamespaceIndex is the lookup name for the most comment index function, which is to index by the namespace field.
	NamespaceIndex string = "namespace"
)

// MetaNamespaceIndexFunc is a default index function that indexes based on an object's namespace
func MetaNamespaceIndexFunc(obj interface{}) ([]string, error) {
	meta, err := meta.Accessor(obj)
	if err != nil {
		return []string{""}, fmt.Errorf("object has no meta: %v", err)
	}
	// 返回名字空间作为索引
	return []string{meta.GetNamespace()}, nil
}

// Index maps the indexed value to a set of keys in the store that match on that value
// 这个是string是对象键，用于查找资源对象
// Indexers： 索引函数名  ======>  索引函数  ======> 索引key值

type Index map[string]sets.String

// Indexers maps a name to a IndexFunc
type Indexers map[string]IndexFunc

// Indices maps a name to an Index
// 这个string是索引函数名
// Indices： 索引函数名  ======> 对应多个索引key值 ======> 每个索引key值对应 不同的资源。(可能不同的索引key值对应相同的资源对象)
type Indices map[string]Index // name to 多个索引键的映射
