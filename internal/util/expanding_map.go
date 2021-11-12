package util

import (
	"activity_log/api/apperror"
	"fmt"
	"sort"
)

type ExpandingMap struct {
	data map[string]*ExpandingMap
}

func NewExpandingMap(input map[string]interface{}) (*ExpandingMap, error) {
	data := map[string]*ExpandingMap{}

	sortedKeys := []string{}
	for key := range input {
		sortedKeys = append(sortedKeys, key)
	}
	sort.Slice(sortedKeys, func(i, j int) bool { return sortedKeys[i] < sortedKeys[j] })

	for _, key := range sortedKeys {
		valObj := input[key]

		if valObj == nil {
			data[key] = NewEmptyExpandingMap()
			continue
		}

		valSubDict, subDictOk := valObj.(map[string]interface{})

		if !subDictOk {
			return nil, fmt.Errorf("%T is not a nested map of strings", valObj)
		}

		newMap, err := NewExpandingMap(valSubDict)
		if err != nil {
			return nil, fmt.Errorf("error at key %q", key)
		}

		data[key] = newMap
	}

	em := &ExpandingMap{
		data: data,
	}

	return em, nil
}

func NewEmptyExpandingMap() *ExpandingMap {
	return &ExpandingMap{
		data: map[string]*ExpandingMap{},
	}
}

func (em *ExpandingMap) IsEmpty() bool {
	return len(em.data) == 0
}

func (em *ExpandingMap) GetSubMap(path []string) (*ExpandingMap, error) {
	if len(path) == 0 {
		return em, nil
	}

	subMap, ok := em.data[path[0]]
	if !ok {
		return nil, apperror.NewNotFoundError(fmt.Errorf("no submap at %v -- self: %+v", path, em.ToRegularMap()))
	}
	if len(path) == 1 {
		return subMap, nil
	}

	newSubMap, err := subMap.GetSubMap(path[1:])
	if err != nil {
		return nil, fmt.Errorf("GetSubMap(%v) returns err: %w", path[1:], err)
	}

	return newSubMap, nil
}

func (em *ExpandingMap) AddSubMapIncludingParent(path []string, newKey string) error {
	if len(path) == 0 {
		return fmt.Errorf("path cannot be empty if adding submap with parent")
	}

	if err := em.AddSubMap(path, path[len(path)-1]); err != nil {
		return fmt.Errorf("AddSubMap(%v, %s) returns err: %w", path, newKey, err)
	}

	if err := em.AddSubMap(path, newKey); err != nil {
		return fmt.Errorf("AddSubMap(%v, %s) returns err: %w", path, newKey, err)
	}

	return nil
}

func (em *ExpandingMap) AddSubMap(path []string, newKey string) error {
	if len(path) == 0 {
		if _, ok := em.data[newKey]; ok {
			return fmt.Errorf("key %q already exists in %+v", newKey, em.data)
		}
		em.data[newKey] = NewEmptyExpandingMap()
		return nil
	}

	currentVal, ok := em.data[path[0]]
	if !ok {
		return fmt.Errorf("no submap at path %v -- self: %+v", path, em)
	}

	if err := currentVal.AddSubMap(path[1:], newKey); err != nil {
		return fmt.Errorf("AddSubMap at path %v returns err: %w", path, err)
	}

	return nil
}

func (em *ExpandingMap) IsEqual(otherEm *ExpandingMap) error {
	if otherEm == nil {
		return fmt.Errorf("other map is nil")
	}

	thisKeys := []string{}
	for key := range em.data {
		thisKeys = append(thisKeys, key)
	}
	sort.Slice(thisKeys, func(i, j int) bool { return thisKeys[i] < thisKeys[j] })

	otherKeys := []string{}
	for key := range otherEm.data {
		otherKeys = append(otherKeys, key)
	}
	sort.Slice(otherKeys, func(i, j int) bool { return otherKeys[i] < otherKeys[j] })

	if len(thisKeys) != len(otherKeys) {
		return fmt.Errorf("maps differ: want: %v, got: %v", thisKeys, otherKeys)
	}

	for _, key := range thisKeys {
		otherVal, ok := otherEm.data[key]
		if !ok {
			return fmt.Errorf("other missing key %q: want: %v, got: %v", key, thisKeys, otherKeys)
		}

		if em.data[key].IsEmpty() {
			if !otherVal.IsEmpty() {
				return fmt.Errorf("this map is empty, but the other isn't")
			}
		} else {
			if err := em.data[key].IsEqual(otherEm.data[key]); err != nil {
				return fmt.Errorf("difference at key %q: %w", key, err)
			}
		}
	}

	return nil
}

func (em *ExpandingMap) ToRegularMap() map[string]interface{} {
	output := map[string]interface{}{}

	sortedKeys := []string{}
	for key := range em.data {
		sortedKeys = append(sortedKeys, key)
	}
	sort.Slice(sortedKeys, func(i, j int) bool { return sortedKeys[i] < sortedKeys[j] })

	for _, key := range sortedKeys {
		if em.data[key].IsEmpty() {
			output[key] = nil
		} else {
			output[key] = em.data[key].ToRegularMap()
		}
	}

	return output
}
