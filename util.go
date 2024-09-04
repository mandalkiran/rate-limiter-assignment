package main

import (
	"fmt"
	"strings"
)

// ParseID parse the id in the request from route pattern
func ParseID(path, pattern string) (string, error) {
	// Split the path and pattern into their respective segments
	pathSegments := strings.Split(strings.Trim(path, "/"), "/")
	patternSegments := strings.Split(strings.Trim(pattern, "/"), "/")

	// Check if the number of segments match
	if len(pathSegments) != len(patternSegments) {
		return "", fmt.Errorf("path does not match the expected pattern")
	}

	// Iterate over the segments to find and return the :id value
	for i, segment := range patternSegments {
		if strings.HasPrefix(segment, ":") {
			return pathSegments[i], nil
		}
	}

	return "", fmt.Errorf("no dynamic parameter found")
}
