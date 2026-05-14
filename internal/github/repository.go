package github

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

var ErrInvalidRepositoryURL = errors.New("invalid GitHub repository URL")

type RepositoryRef struct {
	Owner string
	Repo  string
}

func (r RepositoryRef) String() string {
	if r.Owner == "" || r.Repo == "" {
		return ""
	}
	return r.Owner + "/" + r.Repo
}

func (r RepositoryRef) Equal(other RepositoryRef) bool {
	return strings.EqualFold(r.Owner, other.Owner) &&
		strings.EqualFold(r.Repo, other.Repo)
}

func ParseRepositoryURL(raw string) (RepositoryRef, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return RepositoryRef{}, ErrInvalidRepositoryURL
	}

	if host, path, ok := parseSCPStyleURL(raw); ok {
		return repositoryRefFromHostPath(host, path)
	}

	parsed, err := url.Parse(raw)
	if err != nil || parsed.Host == "" {
		return RepositoryRef{}, fmt.Errorf("%w: %s", ErrInvalidRepositoryURL, "missing host")
	}

	return repositoryRefFromHostPath(parsed.Hostname(), parsed.Path)
}

func parseSCPStyleURL(raw string) (host string, path string, ok bool) {
	if strings.Contains(raw, "://") {
		return "", "", false
	}

	at := strings.LastIndex(raw, "@")
	if at < 0 || at == len(raw)-1 {
		return "", "", false
	}

	afterUser := raw[at+1:]
	colon := strings.Index(afterUser, ":")
	if colon <= 0 || colon == len(afterUser)-1 {
		return "", "", false
	}

	host = afterUser[:colon]
	path = afterUser[colon+1:]
	if strings.Contains(host, "/") {
		return "", "", false
	}
	return host, path, true
}

func repositoryRefFromHostPath(host, path string) (RepositoryRef, error) {
	if !strings.EqualFold(strings.TrimSpace(host), "github.com") {
		return RepositoryRef{}, fmt.Errorf("%w: host must be github.com", ErrInvalidRepositoryURL)
	}

	path = strings.Trim(strings.TrimSpace(path), "/")
	parts := strings.Split(path, "/")
	if len(parts) != 2 {
		return RepositoryRef{}, fmt.Errorf("%w: expected owner/repo path", ErrInvalidRepositoryURL)
	}

	owner := strings.TrimSpace(parts[0])
	repo := strings.TrimSpace(parts[1])
	if strings.HasSuffix(strings.ToLower(repo), ".git") {
		repo = repo[:len(repo)-len(".git")]
	}
	if owner == "" || repo == "" {
		return RepositoryRef{}, fmt.Errorf("%w: owner and repo are required", ErrInvalidRepositoryURL)
	}

	return RepositoryRef{Owner: owner, Repo: repo}, nil
}
