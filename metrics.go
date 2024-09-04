package main

import (
	"strings"
)

type Metric struct {
	Endpoint     string `json:"endpoint"`
	SuccessCount int    `json:"success_count"`
	DeniedCount  int    `json:"denied_count"`
}

var metrics = make(map[string]Metric)

// RecordSuccess records successful counts of endpoints
func RecordSuccess(path string) {
	endpoint := GetProfilingKey(path)
	metric := metrics[endpoint]
	metric.Endpoint = endpoint
	metric.SuccessCount++
	metrics[endpoint] = metric

}

// RecordDenied records denial counts of endpoints
func RecordDenied(path string) {
	endpoint := GetProfilingKey(path)
	metric := metrics[endpoint]
	metric.Endpoint = endpoint
	metric.DeniedCount++
	metrics[endpoint] = metric
}

// GetProfilingKey gets the profile for the endpoint : user, public or admin
func GetProfilingKey(path string) string {
	parts := strings.Split(path, "/")
	return parts[1]
}
