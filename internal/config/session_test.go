package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestGlobalSessionPathUsesUserConfigDir(t *testing.T) {
	// GlobalSessionPath는 실제 사용자 설정 디렉터리를 기준으로 경로를 만든다.
	// 테스트에서는 사용자 홈/설정 환경변수를 임시 디렉터리로 바꿔 실제 PC 설정 파일을 건드리지 않는다.
	userConfigDir := setUserConfigDir(t)

	got, err := GlobalSessionPath()
	if err != nil {
		t.Fatalf("GlobalSessionPath() error = %v", err)
	}

	want := filepath.Join(userConfigDir, appName, globalSessionFile)
	if got != want {
		t.Fatalf("GlobalSessionPath() = %q, want %q", got, want)
	}
}

func TestSaveGlobalSessionWritesSessionJSON(t *testing.T) {
	// SaveGlobalSession은 전역 세션 값을 OS별 user config dir 아래 envio/session.json으로 저장한다.
	// 파일 시스템을 사용하는 함수이므로 임시 config dir을 만들고, 저장된 JSON을 다시 읽어 계약을 검증한다.
	userConfigDir := setUserConfigDir(t)
	session := GlobalSession{
		UserID:     10,
		GithubID:   "octocat",
		DeviceID:   20,
		DeviceName: "desktop",
		PublicKey:  "public-key",
	}

	if err := SaveGlobalSession(session); err != nil {
		t.Fatalf("SaveGlobalSession() error = %v", err)
	}

	path := filepath.Join(userConfigDir, appName, globalSessionFile)
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}

	// MarshalIndent를 사용하는 계약이 유지되는지 확인한다.
	// 이 포맷은 사람이 config 파일을 열어봤을 때 읽기 쉬운 형태를 보장한다.
	if !strings.Contains(string(raw), "\n  \"globalSession\": {") {
		t.Fatalf("session file is not indented JSON: %s", raw)
	}

	// 문자열 비교 대신 JSON으로 역직렬화해서 필드명, 타입, 값이 모두 보존되는지 확인한다.
	var got GlobalConfig
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if got.GlobalSession != session {
		t.Fatalf("saved session = %#v, want %#v", got.GlobalSession, session)
	}
	if strings.Contains(string(raw), "privateKey") || strings.Contains(string(raw), "private-key") {
		t.Fatalf("session file should not include private key: %s", raw)
	}

	// writeJSONFile은 tmp 파일에 먼저 쓴 뒤 Rename으로 교체한다.
	// 성공 후 임시 파일이 남으면 이후 저장 동작이나 사용자 확인에 혼선을 줄 수 있다.
	if _, err := os.Stat(path + ".tmp"); !os.IsNotExist(err) {
		t.Fatalf("temporary session file still exists, stat error = %v", err)
	}
}

func TestHasGlobalSession(t *testing.T) {
	setUserConfigDir(t)

	hasSession, err := HasGlobalSession()
	if err != nil {
		t.Fatalf("HasGlobalSession() error = %v", err)
	}
	if hasSession {
		t.Fatal("HasGlobalSession() = true, want false when session file does not exist")
	}

	if err := SaveGlobalSession(GlobalSession{
		UserID:     10,
		GithubID:   "octocat",
		DeviceID:   20,
		DeviceName: "desktop",
		PublicKey:  "public-key",
	}); err != nil {
		t.Fatalf("SaveGlobalSession() error = %v", err)
	}

	hasSession, err = HasGlobalSession()
	if err != nil {
		t.Fatalf("HasGlobalSession() error = %v", err)
	}
	if !hasSession {
		t.Fatal("HasGlobalSession() = false, want true for valid session")
	}
}

func TestHasGlobalSessionRejectsIncompleteSession(t *testing.T) {
	setUserConfigDir(t)

	if err := SaveGlobalSession(GlobalSession{
		UserID:     10,
		GithubID:   "octocat",
		DeviceName: "desktop",
		PublicKey:  "public-key",
	}); err != nil {
		t.Fatalf("SaveGlobalSession() error = %v", err)
	}

	hasSession, err := HasGlobalSession()
	if err != nil {
		t.Fatalf("HasGlobalSession() error = %v", err)
	}
	if hasSession {
		t.Fatal("HasGlobalSession() = true, want false for incomplete session")
	}
}

func TestWriteJSONFileCreatesParentDirectoryAndReplacesFile(t *testing.T) {
	// writeJSONFile은 SaveGlobalSession의 실제 파일 쓰기 경계다.
	// 부모 디렉터리가 있는 기존 파일을 대상으로 호출했을 때 새 JSON으로 원자적 교체가 되는지 검증한다.
	path := filepath.Join(t.TempDir(), "nested", "config.json")

	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(path, []byte(`{"old":true}`), 0600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	obj := map[string]string{"name": "envio"}
	if err := writeJSONFile(path, obj, 0600); err != nil {
		t.Fatalf("writeJSONFile() error = %v", err)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	// 저장 결과는 JSON 파싱으로 확인해 공백/개행 같은 포맷 차이에 테스트가 과하게 민감해지지 않게 한다.
	var got map[string]string
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if got["name"] != "envio" {
		t.Fatalf("written JSON = %#v, want name=envio", got)
	}
	// 기존 파일의 내용이 일부라도 남아 있으면 Rename 교체가 아니라 덧쓰기/부분 쓰기일 수 있다.
	if strings.Contains(string(raw), "old") {
		t.Fatalf("writeJSONFile() did not replace old content: %s", raw)
	}
	// 성공 경로와 마찬가지로 임시 파일이 남지 않는지 확인한다.
	if _, err := os.Stat(path + ".tmp"); !os.IsNotExist(err) {
		t.Fatalf("temporary file still exists, stat error = %v", err)
	}
}

func TestWriteJSONFileReturnsMarshalError(t *testing.T) {
	// json.MarshalIndent가 처리할 수 없는 값을 넘기면 파일 쓰기 전에 실패해야 한다.
	// 채널 값은 JSON으로 직렬화할 수 없으므로 marshal error를 안정적으로 만들 수 있다.
	path := filepath.Join(t.TempDir(), "config.json")

	err := writeJSONFile(path, map[string]interface{}{"invalid": make(chan int)}, 0600)
	if err == nil {
		t.Fatal("writeJSONFile() error = nil, want marshal error")
	}
	if !strings.Contains(err.Error(), "json") {
		t.Fatalf("writeJSONFile() error = %v, want json marshal context", err)
	}

	// 직렬화 실패는 파일 시스템 변경 전 단계에서 발생해야 한다.
	// 최종 파일과 tmp 파일이 모두 없어야 실패 후 재시도 시 깨끗한 상태를 보장할 수 있다.
	if _, statErr := os.Stat(path); !os.IsNotExist(statErr) {
		t.Fatalf("config file was created after marshal error, stat error = %v", statErr)
	}
	if _, statErr := os.Stat(path + ".tmp"); !os.IsNotExist(statErr) {
		t.Fatalf("temporary file was created after marshal error, stat error = %v", statErr)
	}
}

func setUserConfigDir(t *testing.T) string {
	t.Helper()

	// os.UserConfigDir는 OS별로 다른 환경변수를 참고한다.
	// t.Setenv를 사용하면 테스트 종료 시 원래 환경값이 자동으로 복구된다.
	dir := t.TempDir()
	switch runtime.GOOS {
	case "windows":
		t.Setenv("AppData", dir)
		return dir
	case "darwin":
		t.Setenv("HOME", dir)
		return filepath.Join(dir, "Library", "Application Support")
	default:
		t.Setenv("XDG_CONFIG_HOME", dir)
		return dir
	}
}
