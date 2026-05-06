package cli

import (
	"strings"

	"github.com/Team-NEEEE/envio-cli/internal/i18n"
)

const commandGroupAdditional = "additional"

func rootShort(lang i18n.Language) string {
	if lang == i18n.Korean {
		return "암호화된 환경변수를 명령줄에서 관리합니다."
	}
	return "Manage encrypted environment variables from the command line."
}

func rootLong(lang i18n.Language) string {
	return rootShort(lang)
}

func rootExample() string {
	return ""
}

func completionShort(lang i18n.Language) string {
	if lang == i18n.Korean {
		return "셸 자동완성 스크립트를 생성합니다"
	}
	return "Generate shell completion scripts"
}

func flagText(lang i18n.Language, name string) string {
	if lang == i18n.Korean {
		return koreanFlagText(name)
	}
	return englishFlagText(name)
}

func koreanFlagText(name string) string {
	switch name {
	case "plain":
		return "터미널 UI 없이 일반 텍스트로 출력합니다"
	case "json":
		return "디버깅용 JSON으로 출력합니다"
	case "debug":
		return "디버깅용 상세 JSON을 활성화합니다"
	case "lang":
		return "출력 언어를 지정합니다 (en 또는 ko)"
	case "help":
		return "도움말을 표시합니다"
	case "version":
		return "버전을 표시합니다"
	default:
		return ""
	}
}

func englishFlagText(name string) string {
	switch name {
	case "plain":
		return "Disable terminal UI and print plain text"
	case "json":
		return "Print debugging JSON"
	case "debug":
		return "Enable detailed debugging JSON"
	case "lang":
		return "Output language (en or ko)"
	case "help":
		return "Show help for command"
	case "version":
		return "Show envio version"
	default:
		return ""
	}
}

func usageHeading(lang i18n.Language) string {
	if lang == i18n.Korean {
		return "사용법"
	}
	return "USAGE"
}

func commandGroupHeading(lang i18n.Language, groupID string) string {
	switch groupID {
	case commandGroupAdditional:
		if lang == i18n.Korean {
			return "추가 명령어"
		}
		return "ADDITIONAL COMMANDS"
	default:
		return strings.ToUpper(groupID)
	}
}

func flagsHeading(lang i18n.Language) string {
	if lang == i18n.Korean {
		return "옵션"
	}
	return "FLAGS"
}

func examplesHeading(lang i18n.Language) string {
	if lang == i18n.Korean {
		return "예시"
	}
	return "EXAMPLES"
}

func learnMoreHeading(lang i18n.Language) string {
	if lang == i18n.Korean {
		return "더 알아보기"
	}
	return "LEARN MORE"
}
