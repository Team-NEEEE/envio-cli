package config

import (
	"strings"
)

const (
	EnvAPIURL           = "ENVIO_API_URL"
	DefaultAPIURL       = "http://localhost:8080"
	GitHubAppInstallURL = "https://github.com/apps/envio-official/installations/new"
)

func APIURLFromEnv(env map[string]string) string {
	return APIURLOrDefault(env[EnvAPIURL])
}

func APIURLOrDefault(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return DefaultAPIURL
	}
	return trimmed
}
