package cli

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/Team-NEEEE/envio-cli/internal/config"
	envcrypto "github.com/Team-NEEEE/envio-cli/internal/crypto"
)

func TestRunHelpUsesEnglishByDefault(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	var errOut bytes.Buffer
	code := Run(context.Background(), Runtime{
		Args:   []string{"--help"},
		Stdout: &out,
		Stderr: &errOut,
		CWD:    t.TempDir(),
	})
	if code != 0 {
		t.Fatalf("Run() exit = %d, stderr = %s", code, errOut.String())
	}
	if !strings.Contains(out.String(), "USAGE") {
		t.Fatalf("help should use gh-style usage heading: %s", out.String())
	}
	if !strings.Contains(out.String(), "ADDITIONAL COMMANDS") || !strings.Contains(out.String(), "login:") {
		t.Fatalf("help should show login as a user-facing command: %s", out.String())
	}
	if strings.Contains(out.String(), "completion") {
		t.Fatalf("help should hide shell completion command: %s", out.String())
	}
	if strings.Contains(out.String(), "api-url") {
		t.Fatalf("help should not expose internal API base URL option: %s", out.String())
	}
	if strings.Contains(out.String(), "Envio는 안전한") || strings.Contains(out.String(), "shared CLI foundation") {
		t.Fatalf("help should not render foundation marketing copy: %s", out.String())
	}
}

func TestRunHelpSupportsKoreanWhenRequested(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	var errOut bytes.Buffer
	code := Run(context.Background(), Runtime{
		Args:   []string{"--lang", "ko", "--help"},
		Stdout: &out,
		Stderr: &errOut,
		CWD:    t.TempDir(),
	})
	if code != 0 {
		t.Fatalf("Run() exit = %d, stderr = %s", code, errOut.String())
	}
	if !strings.Contains(out.String(), "사용법") {
		t.Fatalf("help should use Korean description when requested: %s", out.String())
	}
}

func TestRunCommandHelpIncludesTroubleshootingGuidance(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args []string
		want []string
	}{
		{
			name: "login",
			args: []string{"login", "--help"},
			want: []string{
				"USAGE",
				"envio login [flags]",
				"already logged in",
				"OAuth session expires",
				"envio login --device-name work-laptop",
			},
		},
		{
			name: "create",
			args: []string{"create", "--help"},
			want: []string{
				"USAGE",
				"envio create <repository-url> [flags]",
				"Requirements:",
				"already exists",
				"Envio GitHub App installation",
				"envio create --repo https://github.com/owner/repo",
			},
		},
		{
			name: "link",
			args: []string{"link", "--help"},
			want: []string{
				"USAGE",
				"envio link [repository-url] [flags]",
				"join approval is pending",
				"device key or public key errors",
				"envio link https://github.com/owner/repo",
			},
		},
		{
			name: "push",
			args: []string{"push", "--help"},
			want: []string{
				"USAGE",
				"envio push [env-file] [flags]",
				"Only when env-file is omitted",
				"current working directory",
				"environment file does not exist",
				"version conflict occurs",
				"envio push .env.local",
			},
		},
		{
			name: "pull",
			args: []string{"pull", "--help"},
			want: []string{
				"USAGE",
				"envio pull [env-file] [flags]",
				"Only when env-file is omitted",
				"current working directory",
				"no environment version exists",
				"project key cannot be loaded",
				"envio pull .env.local",
			},
		},
		{
			name: "history",
			args: []string{"history", "--help"},
			want: []string{
				"USAGE",
				"envio history [version] [flags]",
				"interactive list",
				"selected version is missing",
				"envio history v3",
			},
		},
		{
			name: "version",
			args: []string{"version", "--help"},
			want: []string{
				"USAGE",
				"envio version [flags]",
				"release metadata",
				"envio version",
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var out bytes.Buffer
			var errOut bytes.Buffer
			code := Run(context.Background(), Runtime{
				Args:   tt.args,
				Stdout: &out,
				Stderr: &errOut,
				CWD:    t.TempDir(),
			})
			if code != 0 {
				t.Fatalf("Run() exit = %d, stderr = %s", code, errOut.String())
			}
			got := out.String()
			for _, want := range tt.want {
				if !strings.Contains(got, want) {
					t.Fatalf("%s help missing %q:\n%s", tt.name, want, got)
				}
			}
		})
	}
}

func TestRunCommandHelpSupportsKoreanTroubleshootingGuidance(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	var errOut bytes.Buffer
	code := Run(context.Background(), Runtime{
		Args:   []string{"--lang", "ko", "push", "--help"},
		Stdout: &out,
		Stderr: &errOut,
		CWD:    t.TempDir(),
	})
	if code != 0 {
		t.Fatalf("Run() exit = %d, stderr = %s", code, errOut.String())
	}
	for _, want := range []string{
		"사용법",
		"필요 조건:",
		"주의사항:",
		"env-file을 생략할 때만",
		"문제가 있을 때:",
		"버전 충돌",
		"envio pull",
	} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("Korean help missing %q:\n%s", want, out.String())
		}
	}
}

func TestRunWithoutCommandShowsHelp(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	var errOut bytes.Buffer
	code := Run(context.Background(), Runtime{
		Stdout: &out,
		Stderr: &errOut,
		CWD:    t.TempDir(),
	})
	if code != 0 {
		t.Fatalf("Run() exit = %d, stderr = %s", code, errOut.String())
	}
	if !strings.Contains(out.String(), "USAGE") {
		t.Fatalf("root command should show foundation help: %s", out.String())
	}
	if strings.Contains(out.String(), "Envio는 안전한") || strings.Contains(out.String(), "shared CLI foundation") {
		t.Fatalf("root command should not render foundation marketing copy: %s", out.String())
	}
}

func TestRunVersionPrintsDefaultBuildInfo(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	var errOut bytes.Buffer
	code := Run(context.Background(), Runtime{
		Args:   []string{"version"},
		Stdout: &out,
		Stderr: &errOut,
		CWD:    t.TempDir(),
	})
	if code != 0 {
		t.Fatalf("Run() exit = %d, stderr = %s", code, errOut.String())
	}
	want := "envio version dev\ncommit: none\nbuilt: unknown\n"
	if out.String() != want {
		t.Fatalf("version output = %q, want %q", out.String(), want)
	}
	if errOut.Len() != 0 {
		t.Fatalf("stderr = %s, want empty", errOut.String())
	}
}

func TestRunVersionPrintsRuntimeBuildInfo(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	var errOut bytes.Buffer
	code := Run(context.Background(), Runtime{
		Args:    []string{"version"},
		Stdout:  &out,
		Stderr:  &errOut,
		CWD:     t.TempDir(),
		Version: "v0.1.0",
		Commit:  "abc123",
		Date:    "2026-05-15T00:00:00Z",
	})
	if code != 0 {
		t.Fatalf("Run() exit = %d, stderr = %s", code, errOut.String())
	}
	want := "envio version v0.1.0\ncommit: abc123\nbuilt: 2026-05-15T00:00:00Z\n"
	if out.String() != want {
		t.Fatalf("version output = %q, want %q", out.String(), want)
	}
	if errOut.Len() != 0 {
		t.Fatalf("stderr = %s, want empty", errOut.String())
	}
}

func TestRunUnknownCommandPlainShowsUsageAndAvailableCommands(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	var errOut bytes.Buffer
	code := Run(context.Background(), Runtime{
		Args:   []string{"unknown"},
		Stdout: &out,
		Stderr: &errOut,
		CWD:    t.TempDir(),
	})
	if code != 2 {
		t.Fatalf("Run() exit = %d, want 2", code)
	}
	if !strings.Contains(errOut.String(), `unknown command "unknown" for "envio"`) {
		t.Fatalf("plain error should show cobra-style command error: %s", errOut.String())
	}
	if !strings.Contains(errOut.String(), "Usage:  envio <command> [flags]") {
		t.Fatalf("plain error should show command usage: %s", errOut.String())
	}
	if !strings.Contains(errOut.String(), "Available commands:") ||
		!strings.Contains(errOut.String(), "login") ||
		!strings.Contains(errOut.String(), "create") {
		t.Fatalf("plain error should show available user-facing commands: %s", errOut.String())
	}
	if strings.Contains(errOut.String(), "completion") {
		t.Fatalf("plain error should not show hidden commands: %s", errOut.String())
	}
	if strings.Contains(errOut.String(), "Next step") || strings.Contains(errOut.String(), "UNKNOWN_COMMAND") {
		t.Fatalf("plain error should not use UI hint/debug shape: %s", errOut.String())
	}
}

func TestRunInvalidCompletionShellShowsUsageAndAvailableValues(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	var errOut bytes.Buffer
	code := Run(context.Background(), Runtime{
		Args:   []string{"completion", "cmd"},
		Stdout: &out,
		Stderr: &errOut,
		CWD:    t.TempDir(),
	})
	if code != 2 {
		t.Fatalf("Run() exit = %d, want 2", code)
	}
	if !strings.Contains(errOut.String(), `invalid argument "cmd" for "envio completion"`) {
		t.Fatalf("plain error should show invalid argument: %s", errOut.String())
	}
	if !strings.Contains(errOut.String(), "Usage:  envio completion [bash|zsh|fish|powershell] [flags]") {
		t.Fatalf("plain error should show completion usage: %s", errOut.String())
	}
	if !strings.Contains(errOut.String(), "Available values:\n  bash\n  zsh\n  fish\n  powershell") {
		t.Fatalf("plain error should show valid completion shells: %s", errOut.String())
	}
	if strings.Contains(errOut.String(), "Hint:") || strings.Contains(errOut.String(), "UNKNOWN_ARGUMENT") {
		t.Fatalf("plain input error should not use UI hint/debug shape: %s", errOut.String())
	}
}

func TestRunCreateInputErrorsShowUsage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		want string
		args []string
	}{
		{
			name: "missing repository URL",
			args: []string{"create"},
			want: "repository URL is required",
		},
		{
			name: "positional and repo flag together",
			args: []string{"create", "https://github.com/Team-NEEEE/envio-cli", "--repo", "https://github.com/Team-NEEEE/envio-cli"},
			want: "use either repository-url argument or --repo",
		},
		{
			name: "too many args",
			args: []string{"create", "https://github.com/Team-NEEEE/envio-cli", "extra"},
			want: "accepts at most 1 arg(s), received 2",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var out bytes.Buffer
			var errOut bytes.Buffer
			code := Run(context.Background(), Runtime{
				Args:   tt.args,
				Stdout: &out,
				Stderr: &errOut,
				CWD:    t.TempDir(),
			})
			if code != 2 {
				t.Fatalf("Run() exit = %d, want 2", code)
			}
			if !strings.Contains(errOut.String(), tt.want) {
				t.Fatalf("stderr = %s, want containing %q", errOut.String(), tt.want)
			}
			if !strings.Contains(errOut.String(), "Usage:  envio create <repository-url> [flags]") {
				t.Fatalf("stderr should include create usage: %s", errOut.String())
			}
		})
	}
}

func TestRunCreateRepositoryMismatchDoesNotCallAPI(t *testing.T) {
	setCLIUserConfigDir(t)
	saveValidCLISession(t)
	cwd := initGitRepository(t, "git@github.com:Team-NEEEE/envio-cli.git")
	called := false
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		called = true
	}))
	defer server.Close()

	var out bytes.Buffer
	var errOut bytes.Buffer
	code := Run(context.Background(), Runtime{
		Args:       []string{"--plain", "--api-url", server.URL, "create", "https://github.com/Team-NEEEE/envio-server"},
		Stdout:     &out,
		Stderr:     &errOut,
		CWD:        cwd,
		IsTerminal: func() bool { return false },
	})
	if code != 1 {
		t.Fatalf("Run() exit = %d, want 1, stdout = %s, stderr = %s", code, out.String(), errOut.String())
	}
	if called {
		t.Fatal("API server should not be called when repository context mismatches")
	}
	if !strings.Contains(errOut.String(), "Current repository: Team-NEEEE/envio-cli") ||
		!strings.Contains(errOut.String(), "Input repository: Team-NEEEE/envio-server") {
		t.Fatalf("stderr should include normalized repository hint: %s", errOut.String())
	}
}

func TestRunCreatePlainSuccessDoesNotExposeKeys(t *testing.T) {
	setCLIUserConfigDir(t)
	saveValidCLISession(t)
	cwd := initGitRepository(t, "git@github.com:Team-NEEEE/envio-cli.git")
	_, publicPEM, err := envcrypto.GenerateRSAKeyPairPEM()
	if err != nil {
		t.Fatalf("GenerateRSAKeyPairPEM() error = %v", err)
	}

	var createCalled bool
	var saveCalled bool
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		switch request.URL.Path {
		case "/api/v1/cli/create":
			createCalled = true
			if request.Method != http.MethodPost {
				t.Fatalf("create method = %s", request.Method)
			}
			var body struct {
				RepositoryURL string `json:"repositoryUrl"`
				PublicKey     string `json:"publicKey"`
				DeviceID      int64  `json:"deviceId"`
			}
			if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
				t.Fatalf("decode create body: %v", err)
			}
			if body.RepositoryURL != "https://github.com/Team-NEEEE/envio-cli.git" ||
				body.DeviceID != 20 ||
				body.PublicKey != "public-key" {
				t.Fatalf("create body = %#v", body)
			}
			if err := json.NewEncoder(writer).Encode(map[string]any{
				"success": true,
				"data": map[string]any{
					"message":        "created",
					"projectId":      1,
					"projectName":    "envio-cli",
					"githubRepoName": "envio-cli",
					"installationId": 9,
					"members": []map[string]any{
						{
							"userId":       10,
							"userDeviceId": 20,
							"githubId":     "octocat",
							"publicKey":    publicPEM,
							"projectRole":  "ADMIN",
						},
					},
				},
				"error":     nil,
				"timestamp": "2026-05-06T00:04:31.127Z",
			}); err != nil {
				t.Fatalf("encode create response: %v", err)
			}
		case "/api/projects/1/wrapped-keys":
			saveCalled = true
			if request.Method != http.MethodPut {
				t.Fatalf("save method = %s", request.Method)
			}
			var body struct {
				PublicKey   string `json:"publicKey"`
				WrappedKeys []struct {
					EncryptedKey string `json:"encryptedKey"`
					UserID       int64  `json:"userId"`
					UserDeviceID int64  `json:"userDeviceId"`
				} `json:"wrappedKeys"`
				DeviceID int64 `json:"deviceId"`
			}
			if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
				t.Fatalf("decode save body: %v", err)
			}
			if len(body.WrappedKeys) != 1 ||
				body.WrappedKeys[0].EncryptedKey == "" ||
				body.DeviceID != 20 ||
				body.PublicKey != "public-key" {
				t.Fatalf("save body = %#v", body)
			}
			if err := json.NewEncoder(writer).Encode(map[string]any{
				"success": true,
				"data": map[string]any{
					"message":      "saved",
					"projectId":    1,
					"updatedCount": 1,
				},
				"error":     nil,
				"timestamp": "2026-05-06T00:04:31.127Z",
			}); err != nil {
				t.Fatalf("encode save response: %v", err)
			}
		default:
			t.Fatalf("unexpected path = %s", request.URL.Path)
		}
	}))
	defer server.Close()

	var out bytes.Buffer
	var errOut bytes.Buffer
	code := Run(context.Background(), Runtime{
		Args:       []string{"--plain", "--api-url", server.URL, "create", "https://github.com/Team-NEEEE/envio-cli.git"},
		Stdout:     &out,
		Stderr:     &errOut,
		CWD:        cwd,
		IsTerminal: func() bool { return false },
	})
	if code != 0 {
		t.Fatalf("Run() exit = %d, stderr = %s", code, errOut.String())
	}
	if !createCalled || !saveCalled {
		t.Fatalf("createCalled=%v saveCalled=%v", createCalled, saveCalled)
	}
	if !strings.Contains(out.String(), "OK: Create completed") ||
		!strings.Contains(out.String(), "Project ID: 1") ||
		!strings.Contains(out.String(), "Wrapped key targets: 1") {
		t.Fatalf("stdout = %s", out.String())
	}
	if strings.Contains(out.String(), "BEGIN PUBLIC KEY") ||
		strings.Contains(out.String(), "encryptedKey") ||
		strings.Contains(out.String(), "wrappedKeys") {
		t.Fatalf("create output leaked key material: %s", out.String())
	}
}

func TestRunLoginAlreadyLoggedInReturnsWarning(t *testing.T) {
	setCLIUserConfigDir(t)
	if err := config.SaveGlobalSession(config.GlobalSession{
		UserID:     10,
		GithubID:   "octocat",
		DeviceID:   20,
		DeviceName: "desktop",
		PublicKey:  "public-key",
	}); err != nil {
		t.Fatalf("SaveGlobalSession() error = %v", err)
	}

	var out bytes.Buffer
	var errOut bytes.Buffer
	code := Run(context.Background(), Runtime{
		Args:       []string{"--plain", "login"},
		Stdout:     &out,
		Stderr:     &errOut,
		CWD:        t.TempDir(),
		IsTerminal: func() bool { return false },
	})
	if code != 0 {
		t.Fatalf("Run() exit = %d, want 0, stderr = %s", code, errOut.String())
	}
	if errOut.Len() != 0 {
		t.Fatalf("stderr = %s, want empty", errOut.String())
	}
	if !strings.Contains(out.String(), "WARN: Already logged in") {
		t.Fatalf("login should render warning result: %s", out.String())
	}
	if strings.Contains(out.String(), "Email") {
		t.Fatalf("login warning should not render email: %s", out.String())
	}
}

func TestRunPushPlainSuccessEncryptsEnvironment(t *testing.T) {
	setCLIUserConfigDir(t)
	saveValidCLISession(t)
	cwd := initGitRepository(t, "git@github.com:Team-NEEEE/envio-cli.git")
	masterKey := []byte("12345678901234567890123456789012")
	writeLegacyProjectSession(t, cwd, masterKey, 2)
	if err := os.WriteFile(filepath.Join(cwd, ".env"), []byte("API_KEY=secret\nDATABASE_URL=postgres://localhost/db\n"), 0600); err != nil {
		t.Fatalf("WriteFile(.env) error = %v", err)
	}

	var called bool
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		called = true
		if request.URL.Path != "/api/core/projects/1/push" || request.Method != http.MethodPost {
			t.Fatalf("unexpected request: %s %s", request.Method, request.URL.Path)
		}
		var body struct {
			EncryptedEnvironment map[string]any `json:"encryptedEnvironment"`
			GithubUserID         string         `json:"githubUserId"`
			ParentVersionID      int64          `json:"parentVersionId"`
		}
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Fatalf("decode push body: %v", err)
		}
		if body.GithubUserID != "octocat" || body.ParentVersionID != 2 {
			t.Fatalf("push body = %#v", body)
		}
		decrypted, err := envcrypto.DecryptEnvironment(body.EncryptedEnvironment, masterKey)
		if err != nil {
			t.Fatalf("DecryptEnvironment() error = %v", err)
		}
		if string(decrypted) != "API_KEY=secret\nDATABASE_URL=postgres://localhost/db\n" {
			t.Fatalf("decrypted = %q", string(decrypted))
		}

		writer.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(writer).Encode(map[string]any{
			"success": true,
			"data": map[string]any{
				"message":         "pushed",
				"historyId":       12,
				"projectId":       1,
				"envName":         "envio-cli",
				"versionId":       3,
				"parentVersionId": 2,
			},
			"error":     nil,
			"timestamp": "2026-05-12T01:02:17",
		}); err != nil {
			t.Fatalf("encode push response: %v", err)
		}
	}))
	defer server.Close()

	var out bytes.Buffer
	var errOut bytes.Buffer
	code := Run(context.Background(), Runtime{
		Args:       []string{"--plain", "--api-url", server.URL, "push"},
		Stdout:     &out,
		Stderr:     &errOut,
		CWD:        cwd,
		IsTerminal: func() bool { return false },
	})
	if code != 0 {
		t.Fatalf("Run() exit = %d, stderr = %s", code, errOut.String())
	}
	if !called {
		t.Fatal("push API was not called")
	}
	if !strings.Contains(out.String(), "OK: Push completed") ||
		!strings.Contains(out.String(), "Version ID: 3") ||
		!strings.Contains(out.String(), "Variable count: 2") {
		t.Fatalf("stdout = %s", out.String())
	}
	if strings.Contains(out.String(), "secret") || strings.Contains(out.String(), "postgres://") {
		t.Fatalf("push output leaked environment values: %s", out.String())
	}
	if got := readProjectVersion(t, cwd); got != 3 {
		t.Fatalf("project session version = %d, want 3", got)
	}
}

func TestRunPullPlainSuccessWritesEnvironment(t *testing.T) {
	setCLIUserConfigDir(t)
	saveValidCLISession(t)
	cwd := initGitRepository(t, "git@github.com:Team-NEEEE/envio-cli.git")
	masterKey := []byte("12345678901234567890123456789012")
	writeLegacyProjectSession(t, cwd, masterKey, 2)
	encrypted, err := envcrypto.EncryptEnvironment([]byte("API_KEY=secret\n"), masterKey)
	if err != nil {
		t.Fatalf("EncryptEnvironment() error = %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/api/core/projects/1/pull/latest" ||
			request.Method != http.MethodPost ||
			request.URL.Query().Get("githubUserId") != "octocat" {
			t.Fatalf("unexpected request: %s %s?%s", request.Method, request.URL.Path, request.URL.RawQuery)
		}
		writer.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(writer).Encode(map[string]any{
			"success": true,
			"data": map[string]any{
				"message":              "pulled",
				"historyId":            12,
				"projectId":            1,
				"envName":              "envio-cli",
				"versionId":            3,
				"encryptedEnvironment": encrypted,
			},
			"error":     nil,
			"timestamp": "2026-05-12T01:02:17",
		}); err != nil {
			t.Fatalf("encode pull response: %v", err)
		}
	}))
	defer server.Close()

	var out bytes.Buffer
	var errOut bytes.Buffer
	code := Run(context.Background(), Runtime{
		Args:       []string{"--plain", "--api-url", server.URL, "pull"},
		Stdout:     &out,
		Stderr:     &errOut,
		CWD:        cwd,
		IsTerminal: func() bool { return false },
	})
	if code != 0 {
		t.Fatalf("Run() exit = %d, stderr = %s", code, errOut.String())
	}
	raw, err := os.ReadFile(filepath.Join(cwd, ".env"))
	if err != nil {
		t.Fatalf("ReadFile(.env) error = %v", err)
	}
	if string(raw) != "API_KEY=secret\n" {
		t.Fatalf(".env = %q", string(raw))
	}
	if strings.Contains(out.String(), "secret") {
		t.Fatalf("pull output leaked environment values: %s", out.String())
	}
	if got := readProjectVersion(t, cwd); got != 3 {
		t.Fatalf("project session version = %d, want 3", got)
	}
}

func TestRunHistoryPlainListsVersionsWithoutDecrypting(t *testing.T) {
	setCLIUserConfigDir(t)
	saveValidCLISession(t)
	cwd := initGitRepository(t, "git@github.com:Team-NEEEE/envio-cli.git")
	masterKey := []byte("12345678901234567890123456789012")
	writeLegacyProjectSession(t, cwd, masterKey, 2)

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/api/cli/projects/1/history" || request.Method != http.MethodGet {
			t.Fatalf("unexpected request: %s %s", request.Method, request.URL.Path)
		}
		writer.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(writer).Encode(map[string]any{
			"success": true,
			"data": map[string]any{
				"message": "histories",
				"histories": []map[string]any{
					{
						"history_id":            12,
						"project_id":            1,
						"version_id":            8,
						"base_version_id":       7,
						"github_id":             "octocat",
						"created_at":            "2026-04-20 11:00:00",
						"encrypted_environment": map[string]any{"ciphertext": "secret-ciphertext"},
					},
				},
			},
			"error":     nil,
			"timestamp": "2026-05-12T01:02:17",
		}); err != nil {
			t.Fatalf("encode history response: %v", err)
		}
	}))
	defer server.Close()

	var out bytes.Buffer
	var errOut bytes.Buffer
	code := Run(context.Background(), Runtime{
		Args:       []string{"--plain", "--api-url", server.URL, "history"},
		Stdout:     &out,
		Stderr:     &errOut,
		CWD:        cwd,
		IsTerminal: func() bool { return false },
	})
	if code != 0 {
		t.Fatalf("Run() exit = %d, stderr = %s", code, errOut.String())
	}
	if !strings.Contains(out.String(), "v8") ||
		!strings.Contains(out.String(), "2026-04-20 11:00") ||
		!strings.Contains(out.String(), "octocat") ||
		!strings.Contains(out.String(), "latest") {
		t.Fatalf("history output = %s", out.String())
	}
	if strings.Contains(out.String(), "secret-ciphertext") {
		t.Fatalf("history list output leaked encrypted payload: %s", out.String())
	}
}

func TestRunHistoryPlainDecryptsSelectedVersion(t *testing.T) {
	setCLIUserConfigDir(t)
	saveValidCLISession(t)
	cwd := initGitRepository(t, "git@github.com:Team-NEEEE/envio-cli.git")
	masterKey := []byte("12345678901234567890123456789012")
	writeLegacyProjectSession(t, cwd, masterKey, 2)
	encrypted, err := envcrypto.EncryptEnvironment([]byte("API_KEY=secret\nDATABASE_URL=postgres://localhost/db\n"), masterKey)
	if err != nil {
		t.Fatalf("EncryptEnvironment() error = %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/api/cli/projects/1/history" || request.Method != http.MethodGet {
			t.Fatalf("unexpected request: %s %s", request.Method, request.URL.Path)
		}
		writer.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(writer).Encode(map[string]any{
			"success": true,
			"data": map[string]any{
				"message": "histories",
				"histories": []map[string]any{
					{
						"history_id":            12,
						"project_id":            1,
						"version_id":            8,
						"base_version_id":       7,
						"github_id":             "octocat",
						"created_at":            "2026-04-20 11:00:00",
						"encrypted_environment": encrypted,
					},
				},
			},
			"error":     nil,
			"timestamp": "2026-05-12T01:02:17",
		}); err != nil {
			t.Fatalf("encode history response: %v", err)
		}
	}))
	defer server.Close()

	var out bytes.Buffer
	var errOut bytes.Buffer
	code := Run(context.Background(), Runtime{
		Args:       []string{"--plain", "--api-url", server.URL, "history", "v8"},
		Stdout:     &out,
		Stderr:     &errOut,
		CWD:        cwd,
		IsTerminal: func() bool { return false },
	})
	if code != 0 {
		t.Fatalf("Run() exit = %d, stderr = %s", code, errOut.String())
	}
	if !strings.Contains(out.String(), "Version v8") ||
		!strings.Contains(out.String(), "Base Version: v7") ||
		!strings.Contains(out.String(), "API_KEY=secret") ||
		!strings.Contains(out.String(), "DATABASE_URL=postgres://localhost/db") {
		t.Fatalf("history version output = %s", out.String())
	}
	if got := readProjectVersion(t, cwd); got != 2 {
		t.Fatalf("history should not update local sync version, got %d", got)
	}
}

func TestRunHistoryInteractiveCanReturnToVersionList(t *testing.T) {
	setCLIUserConfigDir(t)
	saveValidCLISession(t)
	cwd := initGitRepository(t, "git@github.com:Team-NEEEE/envio-cli.git")
	masterKey := []byte("12345678901234567890123456789012")
	writeLegacyProjectSession(t, cwd, masterKey, 2)
	encryptedV3, err := envcrypto.EncryptEnvironment([]byte("API_KEY=v3\n"), masterKey)
	if err != nil {
		t.Fatalf("EncryptEnvironment(v3) error = %v", err)
	}
	encryptedV2, err := envcrypto.EncryptEnvironment([]byte("API_KEY=v2\n"), masterKey)
	if err != nil {
		t.Fatalf("EncryptEnvironment(v2) error = %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/api/cli/projects/1/history" || request.Method != http.MethodGet {
			t.Fatalf("unexpected request: %s %s", request.Method, request.URL.Path)
		}
		writer.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(writer).Encode(map[string]any{
			"success": true,
			"data": []map[string]any{
				{
					"histories_id":          13,
					"project_id":            1,
					"version_id":            3,
					"base_version_id":       2,
					"github_id":             "octocat",
					"created_at":            "2026-05-19T14:02:51",
					"encrypted_environment": encryptedV3,
				},
				{
					"histories_id":          12,
					"project_id":            1,
					"version_id":            2,
					"base_version_id":       1,
					"github_id":             "octocat",
					"created_at":            "2026-05-19T13:42:51",
					"encrypted_environment": encryptedV2,
				},
			},
			"error":     nil,
			"timestamp": "2026-05-12T01:02:17",
		}); err != nil {
			t.Fatalf("encode history response: %v", err)
		}
	}))
	defer server.Close()

	var out bytes.Buffer
	var errOut bytes.Buffer
	code := Run(context.Background(), Runtime{
		Args:       []string{"--api-url", server.URL, "history"},
		Stdin:      strings.NewReader("1\n1\n2\n2\n"),
		Stdout:     &out,
		Stderr:     &errOut,
		CWD:        cwd,
		IsTerminal: func() bool { return true },
	})
	if code != 0 {
		t.Fatalf("Run() exit = %d, stderr = %s", code, errOut.String())
	}
	output := out.String()
	if strings.Count(output, "Select a version to inspect") != 2 ||
		!strings.Contains(output, "Back to version list") ||
		!strings.Contains(output, "Version v3") ||
		!strings.Contains(output, "API_KEY=v3") ||
		!strings.Contains(output, "Version v2") ||
		!strings.Contains(output, "API_KEY=v2") {
		t.Fatalf("interactive history output = %s", output)
	}
}

func TestRunUnknownCommandDebugUsesJSONContract(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	var errOut bytes.Buffer
	code := Run(context.Background(), Runtime{
		Args:   []string{"--debug", "unknown"},
		Stdout: &out,
		Stderr: &errOut,
		CWD:    t.TempDir(),
	})
	if code != 2 {
		t.Fatalf("Run() exit = %d, want 2", code)
	}
	var payload struct {
		Status string `json:"status"`
		Error  struct {
			Code     string `json:"code"`
			Message  string `json:"message"`
			Hint     string `json:"hint"`
			Severity string `json:"severity"`
			ExitCode int    `json:"exitCode"`
		} `json:"error"`
	}
	if err := json.Unmarshal(errOut.Bytes(), &payload); err != nil {
		t.Fatalf("json error output invalid: %v, output = %s", err, errOut.String())
	}
	if payload.Status != "error" || payload.Error.Code != "UNKNOWN_COMMAND" {
		t.Fatalf("payload = %#v", payload)
	}
	if payload.Error.ExitCode != 2 || payload.Error.Severity != "error" {
		t.Fatalf("payload has invalid error contract: %#v", payload.Error)
	}
	if !strings.Contains(errOut.String(), "\n  ") {
		t.Fatalf("json output should be pretty printed: %s", errOut.String())
	}
}

func TestRunDebugFromEnvUsesJSONContract(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	var errOut bytes.Buffer
	code := Run(context.Background(), Runtime{
		Args:    []string{"unknown"},
		Environ: []string{"ENVIO_DEBUG=true"},
		Stdout:  &out,
		Stderr:  &errOut,
		CWD:     t.TempDir(),
	})
	if code != 2 {
		t.Fatalf("Run() exit = %d, want 2", code)
	}
	if !strings.Contains(errOut.String(), `"code": "UNKNOWN_COMMAND"`) {
		t.Fatalf("ENVIO_DEBUG should force JSON debug output: %s", errOut.String())
	}
}

func TestRunCompletionPowershell(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	var errOut bytes.Buffer
	code := Run(context.Background(), Runtime{
		Args:   []string{"completion", "powershell"},
		Stdout: &out,
		Stderr: &errOut,
		CWD:    t.TempDir(),
	})
	if code != 0 {
		t.Fatalf("Run() exit = %d, stderr = %s", code, errOut.String())
	}
	if !strings.Contains(out.String(), "powershell completion for envio") {
		t.Fatalf("completion output = %s", out.String())
	}
}

func setCLIUserConfigDir(t *testing.T) {
	t.Helper()

	dir := t.TempDir()
	switch runtime.GOOS {
	case "windows":
		t.Setenv("AppData", dir)
	case "darwin":
		t.Setenv("HOME", dir)
	default:
		t.Setenv("XDG_CONFIG_HOME", dir)
	}
}

func saveValidCLISession(t *testing.T) {
	t.Helper()

	if err := config.SaveGlobalSession(config.GlobalSession{
		UserID:     10,
		GithubID:   "octocat",
		DeviceID:   20,
		DeviceName: "desktop",
		PublicKey:  "public-key",
	}); err != nil {
		t.Fatalf("SaveGlobalSession() error = %v", err)
	}
}

func initGitRepository(t *testing.T, originURL string) string {
	t.Helper()

	if _, err := exec.LookPath("git"); err != nil {
		t.Skipf("git executable not available: %v", err)
	}

	dir := t.TempDir()
	runGit(t, dir, "init")
	runGit(t, dir, "remote", "add", "origin", originURL)
	return dir
}

func runGit(t *testing.T, cwd string, args ...string) {
	t.Helper()

	allArgs := append([]string{"-C", cwd}, args...)
	cmd := exec.Command("git", allArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s error = %v, output = %s", strings.Join(args, " "), err, string(output))
	}
}

func writeLegacyProjectSession(t *testing.T, cwd string, masterKey []byte, versionID int64) {
	t.Helper()

	session := map[string]any{
		"session": map[string]any{
			"projectId":      1,
			"projectName":    "envio-cli",
			"githubRepoName": "Team-NEEEE/envio-cli",
			"repositoryUrl":  "https://github.com/Team-NEEEE/envio-cli",
			"versionId":      versionID,
			"masterKey": map[string]any{
				"algorithm": "project-master-key-v1",
				"encoding":  "base64",
				"value":     base64.StdEncoding.EncodeToString(masterKey),
			},
		},
	}
	raw, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		t.Fatalf("MarshalIndent(.envio) error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(cwd, ".envio"), raw, 0600); err != nil {
		t.Fatalf("WriteFile(.envio) error = %v", err)
	}
}

func readProjectVersion(t *testing.T, cwd string) int64 {
	t.Helper()

	raw, err := os.ReadFile(filepath.Join(cwd, ".envio", "config"))
	if err != nil {
		if shouldTryLegacyProjectPath(err) {
			raw, err = os.ReadFile(filepath.Join(cwd, ".envio", "session"))
		}
		if shouldTryLegacyProjectPath(err) {
			raw, err = os.ReadFile(filepath.Join(cwd, ".envio"))
		}
	}
	if err != nil {
		t.Fatalf("ReadFile(project config) error = %v", err)
	}
	var payload struct {
		VersionID int64 `json:"versionId"`
		Session   struct {
			VersionID int64 `json:"versionId"`
		} `json:"session"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("Unmarshal(.envio) error = %v", err)
	}
	if payload.VersionID != 0 {
		return payload.VersionID
	}
	return payload.Session.VersionID
}

func shouldTryLegacyProjectPath(err error) bool {
	if err == nil {
		return false
	}
	if os.IsNotExist(err) {
		return true
	}
	return strings.Contains(strings.ToLower(err.Error()), "not a directory")
}
