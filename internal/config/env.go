package config

import "strings"

func EnvMap(environ []string) map[string]string {
	env := make(map[string]string, len(environ))
	for _, pair := range environ {
		key, value, ok := strings.Cut(pair, "=")
		if ok {
			env[key] = value
		}
	}
	return env
}

func IsTruthy(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "y", "on":
		return true
	default:
		return false
	}
}
