package unifiedpushserver

import "k8s.io/client-go/discovery"

// apiVersionChecker is just a container to make it easier to fake the
// check function for tests
type apiVersionChecker struct {
	check func(apiGroupVersion string) (bool, error)
}

func getApiVersionChecker(dc discovery.DiscoveryInterface) *apiVersionChecker {
	// Modified from https://github.com/operator-framework/operator-sdk/blob/947a464dbe968b8af147049e76e40f787ccb0847/pkg/k8sutil/k8sutil.go#L93
	// The Operator Framework one checks a specific resource exists, but this function checks if an API version exists.
	// Theoretically, there can be 2 resources in an API version, 1 existing and 1 not.
	check := func(apiGroupVersion string) (bool, error) {
		apiLists, err := dc.ServerResources()
		if err != nil {
			return false, err
		}
		for _, apiList := range apiLists {
			if apiList.GroupVersion == apiGroupVersion {
				return true, nil
			}
		}
		return false, nil
	}
	return &apiVersionChecker{check: check}
}
