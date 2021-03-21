package util

import (
	"fmt"
	"strings"
)

// GetFileNameFromPath parses the path and extracts the last string after '/'.
// Params:
// 		path string: the path
// Returns:
//		error if path is an empty string
func GetFileNameFromPath(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("received empty string")
	}
	pathSplit := strings.Split(path, "/")
	return pathSplit[len(pathSplit)-1], nil
}
