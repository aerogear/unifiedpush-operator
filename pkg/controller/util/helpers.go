package util

import (
	"os"
	"strings"
)

// IsValidAppNamespace returns an error when the namespace passed is not present in the APP_NAMESPACES environment variable provided to the operator.
func IsValidAppNamespace(crnamespace string, operatorNamespace string, name string) bool {
	appNamespacesEnvVar, found := os.LookupEnv("APP_NAMESPACES")
	if !found {
		return false
	}

	if operatorNamespace == crnamespace {
		return true
	}

	for _, ns := range strings.Split(appNamespacesEnvVar, ",") {
		if ns == crnamespace {
			return true
		}
	}
	return false
}
