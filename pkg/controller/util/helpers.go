package util

import (
	"fmt"
	"os"
	"strings"
)

// IsValidAppNamespace returns an error when the namespace passed is not present in the APP_NAMESPACES environment variable provided to the operator.
func IsValidAppNamespace(crnamespace string, operatorNamespace string, name string) error {
	appNamespacesEnvVar, found := os.LookupEnv("APP_NAMESPACES")
	if !found {
		return fmt.Errorf("APP_NAMESPACES environment variable is required for the creation of the app cr")
	}

	if operatorNamespace == crnamespace {
		return nil
	}

	for _, ns := range strings.Split(appNamespacesEnvVar, ",") {
		if ns == crnamespace {
			return nil
		}
	}

	return fmt.Errorf("The app cr %s was created in a namespace which is not present in the APP_NAMESPACES environment variable provided to the operator", name)
}
