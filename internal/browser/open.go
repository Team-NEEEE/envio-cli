package browser

import (
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

func Open(targetURL string) error {
	cmd, err := openCommand(runtime.GOOS, targetURL)
	if err != nil {
		return err
	}

	// Start만 호출해서 외부 브라우저 실행을 기다리지 않고 로그인 폴링을 계속 진행한다.
	return cmd.Start()
}

func openCommand(goos string, targetURL string) (*exec.Cmd, error) {
	if strings.TrimSpace(targetURL) == "" {
		return nil, errors.New("browser url is required")
	}

	// OS별 기본 URL 핸들러를 사용한다. exec.Command는 shell을 거치지 않아 URL 문자열이 명령으로 해석되지 않는다.
	switch goos {
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", targetURL), nil
	case "darwin":
		return exec.Command("open", targetURL), nil
	case "linux":
		return exec.Command("xdg-open", targetURL), nil
	default:
		return nil, fmt.Errorf("open browser is unsupported on %s", goos)
	}
}
