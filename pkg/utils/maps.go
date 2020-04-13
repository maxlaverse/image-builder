package utils

import "sort"

func MapKeys(aMap map[string]interface{}) []string {
	keys := []string{}
	for i := range aMap {
		keys = append(keys, i)
	}
	sort.Strings(keys)
	return keys
}
