package github

import (
	"errors"
	"testing"
)

func TestParseRepositoryURLNormalizesSupportedGitHubFormats(t *testing.T) {
	t.Parallel()

	tests := []string{
		"https://github.com/Team-NEEEE/envio-cli.git",
		"https://github.com/Team-NEEEE/envio-cli",
		"https://github.com/Team-NEEEE/envio-cli/",
		"git@github.com:Team-NEEEE/envio-cli.git",
		"ssh://git@github.com/Team-NEEEE/envio-cli.git",
	}

	for _, raw := range tests {
		raw := raw
		t.Run(raw, func(t *testing.T) {
			t.Parallel()

			got, err := ParseRepositoryURL(raw)
			if err != nil {
				t.Fatalf("ParseRepositoryURL() error = %v", err)
			}
			if got.Owner != "Team-NEEEE" || got.Repo != "envio-cli" {
				t.Fatalf("ParseRepositoryURL() = %#v", got)
			}
			if got.String() != "Team-NEEEE/envio-cli" {
				t.Fatalf("String() = %q", got.String())
			}
		})
	}
}

func TestParseRepositoryURLRejectsInvalidValues(t *testing.T) {
	t.Parallel()

	tests := []string{
		"",
		"https://gitlab.com/Team-NEEEE/envio-cli.git",
		"https://github.com/Team-NEEEE",
		"https://github.com/Team-NEEEE/",
		"https://github.com/Team-NEEEE/envio-cli/tree/main",
		"git@github.com:Team-NEEEE/.git",
		"not-a-url",
	}

	for _, raw := range tests {
		raw := raw
		t.Run(raw, func(t *testing.T) {
			t.Parallel()

			_, err := ParseRepositoryURL(raw)
			if !errors.Is(err, ErrInvalidRepositoryURL) {
				t.Fatalf("ParseRepositoryURL() error = %v, want ErrInvalidRepositoryURL", err)
			}
		})
	}
}

func TestRepositoryRefEqualIgnoresCase(t *testing.T) {
	t.Parallel()

	left := RepositoryRef{Owner: "Team-NEEEE", Repo: "Envio-CLI"}
	right := RepositoryRef{Owner: "team-neeee", Repo: "envio-cli"}
	if !left.Equal(right) {
		t.Fatalf("%#v should equal %#v", left, right)
	}
}
