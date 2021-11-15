package util

import (
	"encoding/json"
	"fmt"
	"sort"
)

func NestedMapsEqual(lhs map[string]interface{}, rhs map[string]interface{}) error {
	if len(lhs) != len(rhs) {
		return fmt.Errorf("maps differ in size")
	}

	for key, val := range lhs {
		otherVal, ok := rhs[key]
		if !ok {
			return fmt.Errorf("rhs does not contain key %q", key)
		}

		if val == nil && otherVal == nil {
			continue
		}

		if fmt.Sprintf("%T", val) != fmt.Sprintf("%T", otherVal) {
			return fmt.Errorf("at key %q lhs value is %T but rhs value is %T", key, val, otherVal)
		}

		valDict, ok := val.(map[string]interface{})
		if ok {
			otherValDict := otherVal.(map[string]interface{})
			if err := NestedMapsEqual(valDict, otherValDict); err != nil {
				return fmt.Errorf("maps differ at key %q: %w", key, err)
			}
			continue
		}

		if val != otherVal {
			return fmt.Errorf("maps differ at key %q. got %v, want %v", key, val, otherVal)
		}
	}

	return nil
}

func CopyNestedMap(inputMap map[string]interface{}) (map[string]interface{}, error) {
	bytes, err := json.Marshal(inputMap)
	if err != nil {
		return nil, fmt.Errorf("json.Marshal(%+v) returns err: %w", inputMap, err)
	}
	var newMap map[string]interface{}
	if err := json.Unmarshal(bytes, &newMap); err != nil {
		return nil, fmt.Errorf("json.Unmarshall() returns err: %w", err)
	}

	return newMap, err
}

func SortedMapKeysAsc(input map[string]interface{}) []string {
	output := []string{}
	for key := range input {
		output = append(output, key)
	}
	sort.Slice(output, func(i, j int) bool { return output[i] < output[j] })
	return output
}
