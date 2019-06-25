package config

import "os"

type Config struct {
	UPSContainerName        string
	PostgresContainerName   string
	OauthProxyContainerName string

	UPSImageStreamName        string
	UPSImageStreamTag         string
	OauthProxyImageStreamName string
	OauthProxyImageStreamTag  string

	PostgresImageStreamNamespace string
	PostgresImageStreamName      string
	PostgresImageStreamTag       string

	UPSImageStreamInitialImage        string
	PostgresImageStreamInitialImage   string
	OauthProxyImageStreamInitialImage string
}

func New() Config {
	return Config{
		UPSContainerName:        getEnv("UPS_CONTAINER_NAME", "ups"),
		PostgresContainerName:   getEnv("POSTGRES_CONTAINER_NAME", "postgresql"),
		OauthProxyContainerName: getEnv("OAUTH_PROXY_CONTAINER_NAME", "ups-oauth-proxy"),

		UPSImageStreamName:        getEnv("UPS_IMAGE_STREAM_NAME", "ups-imagestream"),
		UPSImageStreamTag:         getEnv("UPS_IMAGE_STREAM_TAG", "latest"),
		OauthProxyImageStreamName: getEnv("OAUTH_PROXY_IMAGE_STREAM_NAME", "ups-oauth-proxy-imagestream"),
		OauthProxyImageStreamTag:  getEnv("OAUTH_PROXY_IMAGE_STREAM_TAG", "latest"),

		PostgresImageStreamNamespace: getEnv("POSTGRES_IMAGE_STREAM_NAMESPACE", "openshift"),
		PostgresImageStreamName:      getEnv("POSTGRES_IMAGE_STREAM_NAME", "postgresql"),
		PostgresImageStreamTag:       getEnv("POSTGRES_IMAGE_STREAM_TAG", "latest"),

		// these are used when the image stream does not exist and created for the first time by the operator
		UPSImageStreamInitialImage:        getEnv("UPS_IMAGE_STREAM_INITIAL_IMAGE", "docker.io/aerogear/unifiedpush-wildfly-plain:2.2.1.Final"),
		OauthProxyImageStreamInitialImage: getEnv("OAUTH_PROXY_IMAGE_STREAM_INITIAL_IMAGE", "docker.io/openshift/oauth-proxy:v1.1.0"),
	}
}

func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultVal
}
