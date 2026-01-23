package main

import "net/url"

func buildFilterQuery(values url.Values) string {
	// Create a copy to modify
	newValues := make(url.Values)
	for k, v := range values {
		if k != "sort_by" {
			for _, val := range v {
				newValues.Add(k, val)
			}
		}
	}
	return newValues.Encode()
}

func buildQueryMap(values url.Values) map[string]string {
	result := make(map[string]string)
	for k, v := range values {
		result[k] = v[0]
	}
	return result
}
