# AI Commit Message Generation Instructions

## SCOPE

- APPLIES_TO: all commit message generation in this repository.
- PURPOSE: generate Korean commit messages from verified Git evidence.
- ROLE: repository commit-message generation reference.
- NOT_FOR: general code style, PR descriptions, changelogs, or release notes unless the user explicitly asks to reuse this convention.
- GLOBAL_USAGE_RECOMMENDATION: if a general AI instruction file exists, it should point to this file only for commit-message tasks instead of duplicating these rules.

## OUTPUT_MODE

- FINAL_COMMIT_MESSAGE:
  - Use when the user asks for a commit message.
  - Return only the commit message.
  - Do not include analysis, explanations, markdown fences, headings, or alternatives.
- COMMIT_SPLIT_ADVICE:
  - Use when the user asks how to split commits.
  - Return logical commit groups, each with purpose, suggested files, and candidate message.
- ANALYSIS:
  - Use only when the user explicitly asks to evaluate, explain, or improve this instruction file.

## PRIORITY_ORDER

Apply rules in this order when there is conflict.

1. FACTUAL_ACCURACY: never invent facts, tickets, validation results, or changed behavior.
2. MANDATORY_REF: if a verified Jira key exists, the final commit message must include `Ref:`.
3. USER_SCOPE: respect the staged files, selected files, or user-provided commit scope.
4. LOGICAL_PURPOSE: express why the selected changes belong together.
5. FORMAT_CONVENTION: follow gitmoji, type, subject, body, and Ref formatting.
6. BREVITY: keep the message concise after the above priorities are satisfied.

## EVIDENCE_RULES

- Use staged diff first when staged changes exist.
- Use working tree diff only when the user asks about unstaged or all current changes.
- Include untracked files only when they are explicitly selected or clearly part of the requested commit scope.
- Do not mention files outside the selected commit scope.
- Do not mention tests, builds, lint, vet, or validation unless the command output is provided or the command was actually run.
- Do not claim "통과", "검증", "확인" without verified evidence.
- Do not infer business context from file names alone.
- Do not include local tool artifacts such as `.omc/`, temporary files, or Windows artifact files such as `nul` unless the user explicitly selects them.
- Before omitting `Ref:`, inspect the current or selected branch name when branch information is available.
- A JIRA key found in the current or selected branch name is verified evidence and must be included as `Ref:`.
- When the current branch contains a Jira key, a final commit message without `Ref: #<jira-ticket>` is invalid and must be regenerated before returning.
- If the selected scope is ambiguous, switch to COMMIT_SPLIT_ADVICE instead of forcing one final commit message.

## MANDATORY_REF_GATE

Run this gate before returning any FINAL_COMMIT_MESSAGE.

1. Inspect the current branch name when branch information is available.
2. Extract every Jira key matching `S14P[0-9]{2}[A-Z][0-9]{3}-[0-9]+`.
3. If one or more keys exist, use WITH_VERIFIED_TICKET.
4. Append `Ref: #<jira-ticket>` after one empty line at the end of the message.
5. If the draft message has no `Ref:` while a verified key exists, discard it and regenerate.

This gate overrides brevity, body length preferences, and examples without `Ref:`.

For the current repository branch pattern:

```text
branch: chore/S14P31A209-40
required final line: Ref: #S14P31A209-40
```

## MESSAGE_SCHEMA

### WITH_VERIFIED_TICKET

```text
<gitmoji> <type>: <subject>

- <body item>
- <body item>

Ref: #<jira-ticket>
```

### WITHOUT_VERIFIED_TICKET

```text
<gitmoji> <type>: <subject>

- <body item>
- <body item>
```

### SCHEMA_RULES

- Line 1 must be `<gitmoji> <type>: <subject>`.
- Line 2 must be empty.
- Body must use `-` bullets.
- `Ref:` section is allowed only when the ticket is verified.
- Use WITH_VERIFIED_TICKET when the user-provided task, selected branch, or current branch contains a JIRA key.
- Use WITHOUT_VERIFIED_TICKET only when no user-provided, selected-branch, current-branch, or task-context ticket exists.
- Do not return WITHOUT_VERIFIED_TICKET when the current branch name contains a JIRA key.
- If a branch key exists and the message lacks `Ref:`, the schema check fails.
- Never output placeholder text such as `#<jira-ticket>`, `TICKET-NUMBER`, or `TODO`.
- Subject and body must be Korean.
- Type must be lowercase English.

## GITMOJI_TYPE_MAP

| Emoji | Type | Use when the selected change mainly does this |
| --- | --- | --- |
| 🎉 | init | Initializes a project, module, package foundation, or executable entrypoint |
| ✨ | feat | Adds a new behavior, command flow, API contract, renderer, or user-facing capability |
| 🐛 | fix | Corrects wrong behavior, regression, broken output, or invalid error handling |
| ♻️ | refactor | Improves structure without changing intended behavior |
| 🔥 | remove | Removes code, files, commands, dead behavior, or obsolete paths |
| ✅ | test | Adds or updates tests only |
| 📝 | docs | Adds or updates documentation only |
| ➕ | deps | Adds dependencies as the main purpose |
| ➖ | deps | Removes dependencies as the main purpose |
| 🔧 | config | Changes CI, build, tooling, repository configuration, or automation |
| 🚚 | rename | Moves or renames files, packages, paths, or resources |
| ⏪ | revert | Reverts a previous commit or change |
| 💡 | comment | Adds or updates source comments only |

### TYPE_DECISION_RULES

- If production code and tests change together, choose the type from the production-code purpose and mention tests in the body only if relevant.
- If docs and code change together, choose the type from the code purpose unless the selected scope is docs-only.
- If dependency files change only because a feature introduced a library, use `feat`, not `deps`.
- If dependency files change only because of cleanup, upgrade, add, or remove dependency work, use `deps`.
- If CI config and product code are both changed, prefer separate commits unless they are inseparable.
- If the selected scope only changes Markdown instruction, convention, guide, README, or documentation files, use `docs`, not `feat`.
- If the selected scope changes `.github/copilot-instructions.md`, `.github/*instructions*.md`, or other AI guidance Markdown files, use `docs`, not `feat`.
- If the selected scope changes executable automation, workflow YAML, IDE settings, or tool configuration behavior, use `config`.
- If the selected scope only changes production/test code to satisfy lint, vet, static analysis, or CI quality gates without changing intended behavior, use `refactor`.
- Never use `feat` for instruction documents unless the same selected scope also changes production behavior.

### QUALITY_GATE_FIX_RULES

When the selected diff exists because `golangci-lint`, `go vet`, tests, or build
failed, classify and describe the commit by the failed quality gate and the
behavioral contract being preserved.

- Prefer subjects such as `Go 품질 게이트 지적 사항 정리`, `lint 지적 사항 정리`, or `정적 분석 경고 정리`.
- Body bullets must group fixes by responsibility or analyzer class, not by raw field/file movement.
- Mention analyzer names such as `errcheck`, `revive`, `staticcheck`, `fieldalignment`, or `unused` only when they clarify the quality gate.
- Do not make `fieldalignment` the whole commit purpose when it is only one category among several.
- Do not list individual struct fields, function parameters, or file names as the main body unless the selected scope is truly that narrow.
- If the same selected scope also includes README, MR draft, generated binaries, `.omc/`, or `nul`, switch to COMMIT_SPLIT_ADVICE unless the user explicitly selected only the quality-gate files.

Good:

```text
♻️ refactor: Go 품질 게이트 지적 사항 정리

- JSON 렌더링과 테스트 응답 작성의 에러 처리 명시
- unused, revive, staticcheck 지적 사항을 동작 변경 없이 정리
- fieldalignment 기준에 맞춰 내부 구조체 배치 조정
- TUI 모델 타입 단언 검증 보강
```

Bad:

```text
♻️ refactor: 구조체 필드 순서 정리

- Request 구조체의 Body 필드 위치 변경
- HTTPResponseError 구조체의 필드 순서 조정
- Runtime 구조체의 필드 순서 변경
- CLI 명령어의 도움말 렌더링 함수 인자 수정
```

Reasons invalid:

- The subject overfits one low-level analyzer.
- Body is an implementation/file-field list.
- It hides the verified reason for the change: passing the Go quality gate.
- It omits other meaningful categories such as errcheck, unused, staticcheck, and test assertion safety.

## LOGICAL_COMMIT_SPLITTING

Split by logical purpose, not by file count, package name, or directory alone.

### SPLIT_DECISION_TREE

1. Identify every distinct purpose in the selected changes.
2. Group files that should naturally be reverted together.
3. Separate repository config or CI changes from product code unless one cannot work without the other.
4. Separate API contracts, CLI command wiring, UI rendering, i18n catalog, and documentation when they can be reviewed independently.
5. Keep tests with the behavior they verify when they are directly tied to that behavior.
6. If one subject cannot accurately describe every important change, split the commit.

### GOOD_SPLIT_EXAMPLES

```text
🎉 init: Go CLI 모듈 초기 구성
✨ feat: 공통 API 응답 클라이언트 구성
✨ feat: CLI 출력 렌더링 기반 구성
♻️ refactor: help 렌더링 책임 분리
🔧 config: Go CI 검증 워크플로우 구성
📝 docs: CLI 기반 설계 결정 정리
```

### BAD_SPLIT_SIGNALS

- One commit mixes CI workflow, API client, UI renderer, and documentation without a single shared purpose.
- The subject only names a directory such as `internal/ui`, `internal/api`, or `.github`.
- The body is mostly a file list.
- Reverting part of the commit later would be natural.
- Reviewers would need unrelated review modes for the same commit.

## SUBJECT_RULES

### SUBJECT_REQUIREMENTS

- Format: `<gitmoji> <type>: <subject>`.
- Use Unicode gitmoji, not text codes.
- Type must be lowercase.
- Subject must be Korean.
- Aim for 50 characters or fewer, but do not remove essential meaning just to satisfy length.
- Use noun or concise imperative style: `추가`, `구성`, `정리`, `수정`, `분리`, `보강`, `제거`.
- Do not use past tense: avoid `추가했습니다`, `수정했습니다`.
- Do not end with a period.
- Do not include ticket numbers.
- Do not make the subject file-oriented.
- Do not make the subject implementation-only when the commit defines a broader responsibility.

### SUBJECT_ABSTRACTION_RULE

The subject must name the broadest accurate logical purpose of the selected commit.

- Prefer: behavior, contract, responsibility, workflow, policy, foundation.
- Avoid: file names, package names, library names, raw implementation details.
- Move implementation details to body bullets.

### SUBJECT_ABSTRACTION_EXAMPLES

```text
Good: ✨ feat: CLI 출력 렌더링 기반 구성
Okay: ✨ feat: TUI 및 JSON 모드 지원 추가
Bad:  ✨ feat: internal/ui 파일 추가
Bad:  ✨ feat: AI 커밋 메시지 생성 지침 추가
```

Use `TUI 및 JSON 모드 지원 추가` only when the commit is narrowly limited to TUI and JSON renderers.
Use `CLI 출력 렌더링 기반 구성` when the commit also includes mode selection, plain output, AppError rendering, sensitive value hiding, localization, and tests.
Do not use `AI 커밋 메시지 생성 지침 추가` for Markdown-only instruction changes. That is documentation work, not product feature work.

```text
Good: ✨ feat: 공통 API 응답 클라이언트 구성
Bad:  ✨ feat: response.go 추가
```

Use `공통 API 응답 클라이언트 구성` when the commit defines backend response models, error response models, HTTP request execution, decode behavior, and non-standard response fallback.

```text
Good: ♻️ refactor: help 렌더링 책임 분리
Bad:  ♻️ refactor: help.go 분리
```

Use `help 렌더링 책임 분리` when the commit separates help text, rendering, and input-error handling responsibilities.

```text
Good: 📝 docs: 커밋 메시지 생성 기준 정교화
Bad:  ✨ feat: AI 커밋 메시지 생성 지침 추가
Bad:  📝 docs: 커밋 메시지 생성 시 참고 기준 제공
```

Use `커밋 메시지 생성 기준 정교화` when the commit improves gitmoji/type selection, evidence rules, logical commit splitting, JIRA handling, or AI prompt structure for commit messages.
Avoid vague subjects such as `참고 기준 제공`; name the actual policy being tightened.

## BODY_RULES

### BODY_REQUIREMENTS

- Use 2 to 5 bullets when possible.
- Each bullet must start with `- `.
- Each bullet should describe one meaningful change.
- Prefer behavior, contract, responsibility, or workflow.
- Mention implementation names only when they clarify the actual change.
- Mention tests only when tests were added, updated, or actually run.
- Do not include a raw file list.
- Do not repeat the subject.
- Do not use past tense.
- Do not add trailing periods.

### BODY_CONTENT_PRIORITY

Use body bullets in this priority order.

1. New behavior or contract introduced by the commit.
2. Error handling, edge cases, or safety policy changed by the commit.
3. User-facing or developer-facing workflow changed by the commit.
4. Tests or validation tied to the commit.
5. Important implementation detail only if needed for review.

### BODY_EXAMPLES

```text
- TTY, non-TTY, CI 환경별 출력 모드 선택 기준 추가
- plain, JSON, TUI 렌더러와 공통 실행 흐름 구성
- AppError를 사용자용 hint와 디버그용 JSON으로 분리
- 민감 요약값 숨김과 localized result/error 변환 검증 추가
```

```text
- 백엔드 공통 응답과 에러 응답 모델 추가
- HTTP 요청 생성과 JSON 응답 디코딩 흐름 구성
- 비표준 HTTP 응답의 status와 body snippet 보존
```

### INVALID_BODY_EXAMPLES

```text
여러 파일 수정
- internal/ui 파일 추가
- TUI랑 JSON 추가
- 렌더러를 추가했습니다
- go test 통과
```

Reasons invalid:

- Too vague.
- File-oriented.
- Too narrow for broader output policy changes.
- Past tense.
- Claims validation without evidence.

## JIRA_RULES

### VERIFIED_TICKET_SOURCES

A ticket is verified only when it is provided by one of these sources.

- User explicitly provides the ticket.
- The current branch name is visible and clearly contains a JIRA key.
- The selected branch name is visible and clearly contains a JIRA key.
- Existing commit scope or task context explicitly includes the ticket.
- A user-provided candidate commit message already includes a valid `Ref:` ticket.

Branch tickets are not guesses. They are verified repository context.

### JIRA_FORMAT

```text
Ref: #S14P31A209-32
Ref: #S14P31A209-32, #S14P31A209-45
```

### JIRA_REQUIREMENTS

- Add an empty line before `Ref:`.
- Start with `Ref: #`.
- Separate multiple tickets with `, #`.
- Do not put the ticket in the subject.
- Do not invent, guess, or normalize unknown tickets.
- Do not omit `Ref:` when the current or selected branch contains a JIRA key.
- If exactly one JIRA key is found in the current branch, include it automatically.
- If the current branch contains `S14P31A209-40`, the final commit message must end with `Ref: #S14P31A209-40`.
- Never treat `Ref:` as optional for branches matching the Jira key pattern.
- If the user explicitly provides a different ticket for the current commit task, use the user-provided ticket.
- If branch and user-provided tickets conflict and the intended ticket is unclear, ask before producing a final commit message.
- If a ticket is required but unavailable, ask for the ticket instead of producing a final commit message.
- If a ticket is optional and unavailable, omit the `Ref:` section.

### BRANCH_REF_EXAMPLES

```text
branch: chore/S14P31A209-40
Ref: #S14P31A209-40
```

```text
user-provided ticket: S14P31A209-41
Ref: #S14P31A209-41
```

## REPOSITORY_SPECIFIC_INTENT

Use these project-level preferences when they are supported by the selected diff.

- This repository is a Go CLI project.
- Prefer commit boundaries that keep CLI foundation, API contract, UI rendering, i18n catalog, config helpers, CI, and docs reviewable as separate logical units.
- For Markdown-only AI instruction changes, classify as `docs` and describe the specific policy refined, not that an AI instruction file was added.
- For Go CLI foundation work, prefer subjects such as `Go CLI 모듈 초기 구성`, `Cobra 루트 명령 구성`, or `CLI 공통 실행 기반 구성` according to actual scope.
- For backend response work, prefer `공통 API 응답 클라이언트 구성` when the selected diff defines response structs, backend error structs, HTTP client behavior, and decode fallback.
- For output work, prefer `CLI 출력 렌더링 기반 구성` when the selected diff includes plain, JSON, TUI, mode selection, localized messages, sensitive value hiding, and output tests.
- For help/error command UX work, prefer gh-style wording such as `help 렌더링 책임 분리`, `입력 오류 usage 렌더링 보강`, or `숨김 명령어 노출 오류 수정` according to actual scope.
- For i18n work, prefer file-based catalog wording when the selected diff uses locale files, embedded messages, or resolver behavior.
- Do not preserve feature-specific command wording when the selected diff intentionally moves the branch toward a common foundation.

## FINAL_SELF_CHECK

Before returning a final commit message, verify all items mentally.

- Unicode gitmoji is used.
- Type is lowercase and matches the selected purpose.
- Subject is Korean, concise, purpose-oriented, and not file-oriented.
- Subject has no ticket and no final period.
- Body uses `-` bullets.
- Body describes verified changes only.
- Body does not claim unverified tests or builds.
- `Ref:` exists only for verified tickets.
- `Ref:` is included when the current branch contains a JIRA key.
- If the current branch contains a JIRA key and `Ref:` is missing, do not return the message.
- The final non-empty line is `Ref: #<verified-jira-ticket>` when exactly one branch key exists.
- No placeholder ticket remains.
- The message can be understood without reading the file list.
- The selected commit scope would be natural to revert as one unit.

## CANONICAL_EXAMPLES

### Go CLI module initialization

```text
🎉 init: Go CLI 모듈 초기 구성

- Go module과 CLI 실행 진입점 추가
- 기본 패키지 경로와 빌드 대상 구성
- 로컬 실행 기준 명령어 정리

Ref: #S14P31A209-1
```

### Common API response client

```text
✨ feat: 공통 API 응답 클라이언트 구성

- 백엔드 공통 응답과 에러 응답 모델 추가
- HTTP 요청 생성과 JSON 응답 디코딩 흐름 구성
- 비표준 HTTP 응답의 status와 body snippet 보존

Ref: #S14P31A209-24
```

### CLI output rendering foundation

```text
✨ feat: CLI 출력 렌더링 기반 구성

- TTY, non-TTY, CI 환경별 출력 모드 선택 기준 추가
- plain, JSON, TUI 렌더러와 공통 실행 흐름 구성
- AppError를 사용자용 hint와 디버그용 JSON으로 분리
- 민감 요약값 숨김과 localized result/error 변환 검증 추가

Ref: #S14P31A209-12
```

### Help rendering refactor

```text
♻️ refactor: help 렌더링 책임 분리

- help 문구, 렌더링, 입력 오류 처리를 역할별 파일로 분리
- 숨김 명령어와 허용 인자 출력 흐름을 별도 책임으로 정리
- gh 스타일 usage 렌더링 동작을 유지하도록 테스트 보강

Ref: #S14P31A209-31
```

### CI configuration

```text
🔧 config: Go CI 검증 워크플로우 구성

- pull request 변경에 대한 Go 테스트와 vet 검증 추가
- 빌드 실패를 조기에 확인하도록 검증 단계를 분리
- 워크플로우 권한과 실행 조건을 저장소 기준으로 정리

Ref: #S14P31A209-40
```

### Commit instruction refinement

```text
📝 docs: 커밋 메시지 생성 기준 정교화

- 지침 문서 변경을 feat가 아닌 docs로 분류하도록 기준 보강
- 검증된 변경 범위와 JIRA 번호만 커밋 메시지에 반영하도록 제한
- 논리적 커밋 분리와 목적 중심 제목 선택 규칙 구체화

Ref: #S14P31A209-41
```
