package config

import "testing"

func TestAPIURLFromEnvTrimsValue(t *testing.T) {
	t.Parallel()

	got := APIURLFromEnv(map[string]string{
		EnvAPIURL: " https://api.envio.dev ",
	})
	if got != "https://api.envio.dev" {
		t.Fatalf("APIURLFromEnv() = %q", got)
	}
}

func TestAPIURLFromEnvFallsBackToDefault(t *testing.T) {
	t.Parallel()

	got := APIURLFromEnv(map[string]string{})
	if got != DefaultAPIURL {
		t.Fatalf("APIURLFromEnv() = %q, want %q", got, DefaultAPIURL)
	}
}

func TestAPIURLOrDefault(t *testing.T) {
	t.Parallel()

	got := APIURLOrDefault(" https://api.envio.dev ")
	if got != "https://api.envio.dev" {
		t.Fatalf("APIURLOrDefault() = %q", got)
	}

	got = APIURLOrDefault(" ")
	if got != DefaultAPIURL {
		t.Fatalf("APIURLOrDefault() = %q, want %q", got, DefaultAPIURL)
	}
}
