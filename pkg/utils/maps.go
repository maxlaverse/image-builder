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

func KeyValueOrEmpty(m map[string]interface{}, key string) string {
	v, ok := m[key]
	if !ok {
		return ""
	}
	return v.(string)
}
