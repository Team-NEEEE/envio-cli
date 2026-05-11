package browser

import "testing"

func TestOpenCommandBuildsOSCommand(t *testing.T) {
	t.Parallel()

	// 실제 브라우저를 띄우지 않고 OS별 실행 명령이 올바르게 조립되는지만 검증한다.
	tests := []struct {
		name string
		goos string
		args []string
	}{
		{
			name: "windows",
			goos: "windows",
			args: []string{"rundll32", "url.dll,FileProtocolHandler", "https://example.com/login"},
		},
		{
			name: "darwin",
			goos: "darwin",
			args: []string{"open", "https://example.com/login"},
		},
		{
			name: "linux",
			goos: "linux",
			args: []string{"xdg-open", "https://example.com/login"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cmd, err := openCommand(tt.goos, "https://example.com/login")
			if err != nil {
				t.Fatalf("openCommand() error = %v", err)
			}
			if len(cmd.Args) != len(tt.args) {
				t.Fatalf("cmd.Args = %#v, want %#v", cmd.Args, tt.args)
			}
			for index, want := range tt.args {
				if cmd.Args[index] != want {
					t.Fatalf("cmd.Args[%d] = %q, want %q", index, cmd.Args[index], want)
				}
			}
		})
	}
}

func TestOpenCommandRejectsEmptyURL(t *testing.T) {
	t.Parallel()

	// 빈 URL은 브라우저 실행 실패 원인을 숨기지 않도록 사전에 거부한다.
	if _, err := openCommand("linux", " "); err == nil {
		t.Fatal("openCommand() error = nil, want error")
	}
}

func TestOpenCommandRejectsUnsupportedOS(t *testing.T) {
	t.Parallel()

	// 지원하지 않는 OS는 성공처럼 처리하지 않고 호출자에게 명시적으로 알린다.
	if _, err := openCommand("plan9", "https://example.com/login"); err == nil {
		t.Fatal("openCommand() error = nil, want error")
	}
}
