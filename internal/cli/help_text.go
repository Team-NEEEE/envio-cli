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

func versionShort(lang i18n.Language) string {
	if lang == i18n.Korean {
		return "Envio 버전 정보를 출력합니다."
	}
	return "Print envio version information"
}

func versionLong(lang i18n.Language) string {
	if lang == i18n.Korean {
		return strings.Join([]string{
			"현재 실행 중인 envio 바이너리의 버전, 커밋, 빌드 시각을 출력합니다.",
			"",
			"문제가 있을 때:",
			"  - 버전이 dev, 커밋이 none, 빌드 시각이 unknown이면 릴리스 빌드 정보가 주입되지 않은 바이너리입니다.",
		}, "\n")
	}
	return strings.Join([]string{
		"Print the version, commit, and build time for the current envio binary.",
		"",
		"When troubleshooting:",
		"  - If the version is dev, commit is none, or built is unknown, the binary was built without release metadata.",
	}, "\n")
}

func versionExample(lang i18n.Language) string {
	if lang == i18n.Korean {
		return "envio version"
	}
	return "envio version"
}

func flagText(lang i18n.Language, name string) string {
	if lang == i18n.Korean {
		if text := koreanFlagText(name); text != "" {
			return text
		}
	}
	return englishFlagText(name)
}

func loginShort(lang i18n.Language) string {
	if lang == i18n.Korean {
		return "GitHub OAuth로 Envio에 로그인합니다."
	}
	return "Log in to Envio with GitHub OAuth."
}

func loginLong(lang i18n.Language) string {
	if lang == i18n.Korean {
		return strings.Join([]string{
			"브라우저에서 GitHub OAuth 로그인을 완료하고, 이 CLI 장치의 세션과 암호화 키를 로컬에 저장합니다.",
			"",
			"알아둘 점:",
			"  - 이미 로그인되어 있으면 새 세션을 만들지 않고 종료합니다.",
			"  - 브라우저가 열리지 않거나 OAuth 세션이 만료되면 다시 실행하세요.",
			"  - 장치 키가 맞지 않아 link가 실패하면 `envio login`을 다시 실행해 현재 장치를 등록하세요.",
		}, "\n")
	}
	return strings.Join([]string{
		"Complete GitHub OAuth in a browser, then save the local CLI session and device key.",
		"",
		"Notes:",
		"  - If you are already logged in, no new session is created.",
		"  - If the browser cannot open or the OAuth session expires, run the command again.",
		"  - If link fails because the device key is missing or mismatched, run `envio login` again to register this device.",
	}, "\n")
}

func loginExample(lang i18n.Language) string {
	if lang == i18n.Korean {
		return strings.Join([]string{
			"envio login",
			"envio login --device-name work-laptop",
		}, "\n")
	}
	return strings.Join([]string{
		"envio login",
		"envio login --device-name work-laptop",
	}, "\n")
}

func createShort(lang i18n.Language) string {
	if lang == i18n.Korean {
		return "현재 Git 저장소를 Envio 프로젝트로 등록합니다."
	}
	return "Register the current Git repository as an Envio project."
}

func createLong(lang i18n.Language) string {
	if lang == i18n.Korean {
		return strings.Join([]string{
			"현재 작업 디렉터리가 속한 Git 저장소를 새 Envio 프로젝트로 등록하고, 프로젝트 키를 생성해 로컬 `.envio` 설정에 저장합니다.",
			"",
			"필요 조건:",
			"  - `envio login`으로 로그인되어 있어야 합니다.",
			"  - 현재 디렉터리는 origin remote가 있는 Git 저장소 안이어야 합니다.",
			"  - 입력한 GitHub 저장소 URL은 현재 origin과 같은 owner/repo를 가리켜야 합니다.",
			"",
			"문제가 있을 때:",
			"  - 이미 프로젝트가 있으면 `envio link`로 기존 프로젝트에 연결하세요.",
			"  - GitHub App 또는 저장소 권한 문제가 나오면 Envio GitHub App 설치와 접근 권한을 확인하세요.",
			"  - 협업자의 장치 키가 없다는 오류가 나오면 해당 협업자가 `envio login`을 실행한 뒤 다시 시도하세요.",
		}, "\n")
	}
	return strings.Join([]string{
		"Register the Git repository for the current working directory as a new Envio project, create a project key, and save local `.envio` configuration.",
		"",
		"Requirements:",
		"  - You must be logged in with `envio login`.",
		"  - Run this command inside a Git repository with an origin remote.",
		"  - The GitHub repository URL you pass must match the current origin owner/repo.",
		"",
		"When troubleshooting:",
		"  - If the project already exists, run `envio link` to connect to it.",
		"  - If GitHub App or repository access errors appear, check the Envio GitHub App installation and repository permissions.",
		"  - If a collaborator device key is missing, ask that collaborator to run `envio login`, then try again.",
	}, "\n")
}

func createExample(lang i18n.Language) string {
	if lang == i18n.Korean {
		return strings.Join([]string{
			"envio create https://github.com/owner/repo",
			"envio create --repo https://github.com/owner/repo",
		}, "\n")
	}
	return strings.Join([]string{
		"envio create https://github.com/owner/repo",
		"envio create --repo https://github.com/owner/repo",
	}, "\n")
}

func linkShort(lang i18n.Language) string {
	if lang == i18n.Korean {
		return "현재 Git 저장소를 기존 Envio 프로젝트와 연결합니다."
	}
	return "Link the current Git repository to an existing Envio project."
}

func linkLong(lang i18n.Language) string {
	if lang == i18n.Korean {
		return strings.Join([]string{
			"현재 Git 저장소를 이미 생성된 Envio 프로젝트에 연결하고, 현재 장치에서 사용할 프로젝트 키를 내려받아 로컬 `.envio` 설정에 저장합니다.",
			"",
			"필요 조건:",
			"  - `envio login`으로 로그인되어 있어야 합니다.",
			"  - 현재 디렉터리는 origin remote가 있는 Git 저장소 안이어야 합니다.",
			"  - 저장소 URL을 생략하면 현재 origin URL을 사용합니다.",
			"",
			"문제가 있을 때:",
			"  - 프로젝트가 없으면 먼저 `envio create`를 실행하세요.",
			"  - 참여 승인이 대기 중이면 프로젝트 승인이 완료된 뒤 다시 실행하세요.",
			"  - 장치 키나 공개 키 오류가 나오면 `envio login`을 다시 실행해 현재 장치를 등록하세요.",
			"  - 로컬 `.envio` 경로가 파일과 충돌하면 해당 경로를 정리한 뒤 다시 실행하세요.",
		}, "\n")
	}
	return strings.Join([]string{
		"Link the current Git repository to an existing Envio project, download the project key for this device, and save local `.envio` configuration.",
		"",
		"Requirements:",
		"  - You must be logged in with `envio login`.",
		"  - Run this command inside a Git repository with an origin remote.",
		"  - If repository-url is omitted, the current origin URL is used.",
		"",
		"When troubleshooting:",
		"  - If the project does not exist, run `envio create` first.",
		"  - If join approval is pending, wait for project approval and run the command again.",
		"  - If device key or public key errors appear, run `envio login` again to register this device.",
		"  - If the local `.envio` path conflicts with a file, clear that path and retry.",
	}, "\n")
}

func linkExample(lang i18n.Language) string {
	if lang == i18n.Korean {
		return strings.Join([]string{
			"envio link",
			"envio link https://github.com/owner/repo",
		}, "\n")
	}
	return strings.Join([]string{
		"envio link",
		"envio link https://github.com/owner/repo",
	}, "\n")
}

func pushShort(lang i18n.Language) string {
	if lang == i18n.Korean {
		return "로컬 환경변수 파일을 암호화해 업로드합니다."
	}
	return "Encrypt and upload a local environment file."
}

func pushLong(lang i18n.Language) string {
	if lang == i18n.Korean {
		return strings.Join([]string{
			"로컬 환경변수 파일을 읽어 프로젝트 키로 암호화한 뒤 Envio 서버에 새 버전으로 업로드합니다.",
			"",
			"필요 조건:",
			"  - `envio login`으로 로그인되어 있어야 합니다.",
			"  - 현재 저장소가 `envio create` 또는 `envio link`로 연결되어 있어야 합니다.",
			"",
			"주의사항:",
			"  - env-file을 생략할 때만 저장소 루트의 `.env`를 읽습니다.",
			"  - 다른 이름이나 위치의 파일은 `envio push <env-file>`로 지정하세요.",
			"  - 상대 경로는 현재 작업 디렉터리 기준이며, 절대 경로도 사용할 수 있습니다.",
			"",
			"문제가 있을 때:",
			"  - 환경변수 파일이 없으면 파일을 만든 뒤 다시 실행하세요.",
			"  - dotenv 형식이 잘못되면 파일 내용을 수정하세요.",
			"  - 버전 충돌이 발생하면 `envio pull`로 최신 버전을 받은 뒤 다시 push하세요.",
			"  - 로그인 사용자나 저장소 origin이 로컬 설정과 다르면 `envio link`를 다시 실행하세요.",
		}, "\n")
	}
	return strings.Join([]string{
		"Read a local environment file, encrypt it with the project key, and upload it to Envio as a new version.",
		"",
		"Requirements:",
		"  - You must be logged in with `envio login`.",
		"  - This repository must be connected with `envio create` or `envio link`.",
		"",
		"Cautions:",
		"  - Only when env-file is omitted, `.env` at the repository root is read.",
		"  - For another filename or location, pass it as `envio push <env-file>`.",
		"  - Relative paths are resolved from the current working directory; absolute paths are also accepted.",
		"",
		"When troubleshooting:",
		"  - If the environment file does not exist, create it and run the command again.",
		"  - If dotenv parsing fails, fix the file contents.",
		"  - If a version conflict occurs, run `envio pull` to receive the latest version before pushing again.",
		"  - If the login user or repository origin does not match local config, run `envio link` again.",
	}, "\n")
}

func pushExample(lang i18n.Language) string {
	if lang == i18n.Korean {
		return strings.Join([]string{
			"envio push",
			"envio push .env.local",
		}, "\n")
	}
	return strings.Join([]string{
		"envio push",
		"envio push .env.local",
	}, "\n")
}

func pullShort(lang i18n.Language) string {
	if lang == i18n.Korean {
		return "최신 환경변수 파일을 내려받아 복호화합니다."
	}
	return "Download and decrypt the latest environment file."
}

func pullLong(lang i18n.Language) string {
	if lang == i18n.Korean {
		return strings.Join([]string{
			"Envio 서버에서 최신 환경변수 버전을 내려받아 복호화하고 로컬 파일에 씁니다.",
			"",
			"필요 조건:",
			"  - `envio login`으로 로그인되어 있어야 합니다.",
			"  - 현재 저장소가 `envio create` 또는 `envio link`로 연결되어 있어야 합니다.",
			"",
			"주의사항:",
			"  - env-file을 생략할 때만 저장소 루트의 `.env`에 씁니다.",
			"  - 다른 이름이나 위치에 쓰려면 `envio pull <env-file>`로 지정하세요.",
			"  - 상대 경로는 현재 작업 디렉터리 기준이며, 절대 경로도 사용할 수 있습니다.",
			"",
			"문제가 있을 때:",
			"  - 아직 환경변수 버전이 없으면 `envio push`로 첫 버전을 만드세요.",
			"  - 프로젝트 키를 읽거나 복호화할 수 없으면 `envio link`를 다시 실행해 현재 장치의 키를 복구하세요.",
			"  - 저장소 origin이 로컬 설정과 다르면 올바른 저장소에서 실행하거나 `envio link`를 다시 실행하세요.",
			"  - 대상 파일을 쓸 수 없으면 파일 경로와 권한을 확인하세요.",
		}, "\n")
	}
	return strings.Join([]string{
		"Download the latest environment version from Envio, decrypt it, and write it to a local file.",
		"",
		"Requirements:",
		"  - You must be logged in with `envio login`.",
		"  - This repository must be connected with `envio create` or `envio link`.",
		"",
		"Cautions:",
		"  - Only when env-file is omitted, `.env` at the repository root is written.",
		"  - To write another filename or location, pass it as `envio pull <env-file>`.",
		"  - Relative paths are resolved from the current working directory; absolute paths are also accepted.",
		"",
		"When troubleshooting:",
		"  - If no environment version exists yet, run `envio push` to create the first version.",
		"  - If the project key cannot be loaded or decryption fails, run `envio link` again to restore this device's key.",
		"  - If the repository origin does not match local config, run from the correct repository or run `envio link` again.",
		"  - If the target file cannot be written, check the path and file permissions.",
	}, "\n")
}

func pullExample(lang i18n.Language) string {
	if lang == i18n.Korean {
		return strings.Join([]string{
			"envio pull",
			"envio pull .env.local",
		}, "\n")
	}
	return strings.Join([]string{
		"envio pull",
		"envio pull .env.local",
	}, "\n")
}

func historyShort(lang i18n.Language) string {
	if lang == i18n.Korean {
		return "서버에 저장된 환경변수 버전 이력을 조회합니다."
	}
	return "Inspect environment version history."
}

func historyLong(lang i18n.Language) string {
	if lang == i18n.Korean {
		return strings.Join([]string{
			"Envio 서버에 저장된 환경변수 버전 목록을 조회합니다. 버전을 지정하면 해당 버전을 복호화해 출력합니다.",
			"",
			"필요 조건:",
			"  - `envio login`으로 로그인되어 있어야 합니다.",
			"  - 현재 저장소가 `envio create` 또는 `envio link`로 연결되어 있어야 합니다.",
			"  - 터미널에서는 버전을 생략하면 선택형 목록을 보여주고, `--plain` 또는 `--json`에서는 목록을 바로 출력합니다.",
			"",
			"문제가 있을 때:",
			"  - 아직 환경변수 버전이 없으면 `envio push`로 첫 버전을 만드세요.",
			"  - 선택한 버전이 없으면 `envio history`로 목록을 확인한 뒤 v숫자 형식으로 다시 지정하세요.",
			"  - 복호화가 실패하면 `envio link`를 다시 실행해 현재 장치의 프로젝트 키를 복구하세요.",
		}, "\n")
	}
	return strings.Join([]string{
		"List environment versions stored on Envio. If a version is provided, decrypt and print that version.",
		"",
		"Requirements:",
		"  - You must be logged in with `envio login`.",
		"  - This repository must be connected with `envio create` or `envio link`.",
		"  - In a terminal, omitting the version shows an interactive list; with `--plain` or `--json`, the list is printed directly.",
		"",
		"When troubleshooting:",
		"  - If no environment version exists yet, run `envio push` to create the first version.",
		"  - If the selected version is missing, run `envio history` and retry with a v-number such as v3.",
		"  - If decryption fails, run `envio link` again to restore this device's project key.",
	}, "\n")
}

func historyExample(lang i18n.Language) string {
	if lang == i18n.Korean {
		return strings.Join([]string{
			"envio history",
			"envio history v3",
			"envio --json history",
		}, "\n")
	}
	return strings.Join([]string{
		"envio history",
		"envio history v3",
		"envio --json history",
	}, "\n")
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
	case "repo":
		return "등록할 GitHub 저장소 URL"
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
	case "api-url":
		return "Envio API server URL"
	case "device-name":
		return "CLI device name to register"
	case "help":
		return "Show help for command"
	case "version":
		return "Show envio version"
	case "repo":
		return "GitHub repository URL to register"
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
