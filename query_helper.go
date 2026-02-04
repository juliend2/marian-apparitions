package main

import "net/url"

func buildQueryMap(values url.Values) map[string]string {
	result := make(map[string]string)
	for k, v := range values {
		result[k] = v[0]
	}
	return result
}
