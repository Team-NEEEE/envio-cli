package ui

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/Team-NEEEE/envio-cli/internal/i18n"
)

type HistoryEntryView struct {
	GithubID      string `json:"githubId"`
	CreatedAt     string `json:"createdAt"`
	HistoryID     int64  `json:"historyId"`
	ProjectID     int64  `json:"projectId"`
	VersionID     int64  `json:"versionId"`
	BaseVersionID int64  `json:"baseVersionId,omitempty"`
	Latest        bool   `json:"latest,omitempty"`
}

type HistoryVersionView struct {
	Environment   string `json:"environment"`
	GithubID      string `json:"githubId"`
	CreatedAt     string `json:"createdAt"`
	HistoryID     int64  `json:"historyId"`
	ProjectID     int64  `json:"projectId"`
	VersionID     int64  `json:"versionId"`
	BaseVersionID int64  `json:"baseVersionId,omitempty"`
	VariableCount int    `json:"variableCount"`
}

type HistoryAction string

const (
	HistoryActionBack HistoryAction = "back"
	HistoryActionExit HistoryAction = "exit"
)

type HistoryPromptSession struct {
	scanner *bufio.Scanner
	output  io.Writer
	lang    i18n.Language
}

func NewHistoryPromptSession(input io.Reader, output io.Writer, lang i18n.Language) *HistoryPromptSession {
	if output == nil {
		output = io.Discard
	}
	if input == nil {
		return &HistoryPromptSession{output: output, lang: lang}
	}
	return &HistoryPromptSession{
		scanner: bufio.NewScanner(input),
		output:  output,
		lang:    lang,
	}
}

func RenderHistoryList(output io.Writer, histories []HistoryEntryView, lang i18n.Language) int {
	if output == nil {
		output = io.Discard
	}
	_, _ = fmt.Fprintln(output, historySelectionPrompt(lang))
	_, _ = fmt.Fprintln(output)
	if len(histories) == 0 {
		_, _ = fmt.Fprintln(output, historyEmptyText(lang))
		return 0
	}
	_, _ = fmt.Fprintf(
		output,
		"  %-3s %-8s %-16s %-16s %s\n",
		historyNumberHeader(lang),
		historyVersionHeader(lang),
		historyCreatedAtHeader(lang),
		historyAuthorHeader(lang),
		historyStatusHeader(lang),
	)
	for index, history := range histories {
		latest := ""
		if history.Latest {
			latest = "latest"
		}
		_, _ = fmt.Fprintf(
			output,
			"  %-3d %-8s %-16s %-16s %s\n",
			index+1,
			historyVersionLabel(history.VersionID),
			shortHistoryTimestamp(history.CreatedAt),
			history.GithubID,
			latest,
		)
	}
	return 0
}

func RenderHistoryVersion(output io.Writer, history HistoryVersionView, lang i18n.Language) int {
	if output == nil {
		output = io.Discard
	}
	_, _ = fmt.Fprintf(output, "Version %s\n", historyVersionLabel(history.VersionID))
	_, _ = fmt.Fprintf(output, "%s: %s\n", historyAuthorLabel(lang), history.GithubID)
	_, _ = fmt.Fprintf(output, "%s: %s\n", historyCreatedAtLabel(lang), history.CreatedAt)
	_, _ = fmt.Fprintf(output, "%s: %s\n", historyBaseVersionLabel(lang), historyVersionLabelOrDash(history.BaseVersionID))
	_, _ = fmt.Fprintln(output)
	_, _ = fmt.Fprintln(output, ".env")
	_, _ = fmt.Fprintln(output, "----------------------------")
	_, _ = fmt.Fprint(output, history.Environment)
	if history.Environment != "" && !strings.HasSuffix(history.Environment, "\n") {
		_, _ = fmt.Fprintln(output)
	}
	return 0
}

func RenderHistoryListJSON(output io.Writer, histories []HistoryEntryView) int {
	if output == nil {
		output = io.Discard
	}
	payload := struct {
		Status    string             `json:"status"`
		Histories []HistoryEntryView `json:"histories"`
	}{
		Status:    "success",
		Histories: histories,
	}
	encoder := json.NewEncoder(output)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(payload); err != nil {
		return 1
	}
	return 0
}

func RenderHistoryVersionJSON(output io.Writer, history HistoryVersionView) int {
	if output == nil {
		output = io.Discard
	}
	payload := struct {
		Status  string             `json:"status"`
		History HistoryVersionView `json:"history"`
	}{
		Status:  "success",
		History: history,
	}
	encoder := json.NewEncoder(output)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(payload); err != nil {
		return 1
	}
	return 0
}

func PromptHistorySelection(input io.Reader, output io.Writer, histories []HistoryEntryView, lang i18n.Language) (string, error) {
	return NewHistoryPromptSession(input, output, lang).PromptSelection(histories)
}

func (p *HistoryPromptSession) PromptSelection(histories []HistoryEntryView) (string, error) {
	if p == nil {
		return "", io.EOF
	}
	RenderHistoryList(p.output, histories, p.lang)
	if len(histories) == 0 {
		return "", io.EOF
	}
	if p.scanner == nil {
		return "", io.EOF
	}
	_, _ = fmt.Fprintln(p.output)
	_, _ = fmt.Fprint(p.output, historySelectionInputPrompt(p.lang))

	if !p.scanner.Scan() {
		if err := p.scanner.Err(); err != nil {
			return "", err
		}
		return "", io.EOF
	}
	choice := strings.TrimSpace(p.scanner.Text())
	if index, err := strconv.Atoi(choice); err == nil && index >= 1 && index <= len(histories) {
		return historyVersionLabel(histories[index-1].VersionID), nil
	}
	return choice, nil
}

func PromptHistoryAction(input io.Reader, output io.Writer, lang i18n.Language) (HistoryAction, error) {
	return NewHistoryPromptSession(input, output, lang).PromptAction()
}

func (p *HistoryPromptSession) PromptAction() (HistoryAction, error) {
	if p == nil || p.scanner == nil {
		return HistoryActionExit, io.EOF
	}
	for {
		_, _ = fmt.Fprintln(p.output)
		_, _ = fmt.Fprintln(p.output, historyActionPrompt(p.lang))
		_, _ = fmt.Fprintf(p.output, "  1. %s\n", historyBackActionLabel(p.lang))
		_, _ = fmt.Fprintf(p.output, "  2. %s\n", historyExitActionLabel(p.lang))
		_, _ = fmt.Fprintln(p.output)
		_, _ = fmt.Fprint(p.output, historySelectionInputPrompt(p.lang))

		if !p.scanner.Scan() {
			if err := p.scanner.Err(); err != nil {
				return HistoryActionExit, err
			}
			return HistoryActionExit, io.EOF
		}
		switch normalizeHistoryActionChoice(p.scanner.Text()) {
		case "1", "back", "b":
			return HistoryActionBack, nil
		case "2", "exit", "quit", "q":
			return HistoryActionExit, nil
		default:
			_, _ = fmt.Fprintln(p.output, historyInvalidActionText(p.lang))
		}
	}
}

func historySelectionPrompt(lang i18n.Language) string {
	if lang == i18n.Korean {
		return "? 확인할 버전을 선택하세요. 번호 또는 버전(v3)을 입력한 뒤 Enter를 누르세요."
	}
	return "? Select a version to inspect. Type a number or version (for example, 1 or v3), then press Enter."
}

func historySelectionInputPrompt(lang i18n.Language) string {
	if lang == i18n.Korean {
		return "선택: "
	}
	return "Selection: "
}

func historyActionPrompt(lang i18n.Language) string {
	if lang == i18n.Korean {
		return "? 작업을 선택하세요."
	}
	return "? Choose the next action."
}

func historyBackActionLabel(lang i18n.Language) string {
	if lang == i18n.Korean {
		return "버전 목록으로 돌아가기"
	}
	return "Back to version list"
}

func historyExitActionLabel(lang i18n.Language) string {
	if lang == i18n.Korean {
		return "종료"
	}
	return "Exit"
}

func historyInvalidActionText(lang i18n.Language) string {
	if lang == i18n.Korean {
		return "1 또는 2를 입력해 주세요."
	}
	return "Enter 1 or 2."
}

func historyNumberHeader(lang i18n.Language) string {
	if lang == i18n.Korean {
		return "번호"
	}
	return "No"
}

func historyVersionHeader(lang i18n.Language) string {
	if lang == i18n.Korean {
		return "버전"
	}
	return "Version"
}

func historyCreatedAtHeader(lang i18n.Language) string {
	if lang == i18n.Korean {
		return "생성 시각"
	}
	return "Created"
}

func historyAuthorHeader(lang i18n.Language) string {
	if lang == i18n.Korean {
		return "작성자"
	}
	return "Author"
}

func historyStatusHeader(lang i18n.Language) string {
	if lang == i18n.Korean {
		return "상태"
	}
	return "Status"
}

func historyEmptyText(lang i18n.Language) string {
	if lang == i18n.Korean {
		return "표시할 버전 이력이 없습니다."
	}
	return "No history versions to display."
}

func historyAuthorLabel(lang i18n.Language) string {
	if lang == i18n.Korean {
		return "작성자"
	}
	return "Author"
}

func historyCreatedAtLabel(lang i18n.Language) string {
	if lang == i18n.Korean {
		return "생성 시각"
	}
	return "Created At"
}

func historyBaseVersionLabel(lang i18n.Language) string {
	if lang == i18n.Korean {
		return "기준 버전"
	}
	return "Base Version"
}

func historyVersionLabel(versionID int64) string {
	if versionID <= 0 {
		return ""
	}
	return "v" + strconv.FormatInt(versionID, 10)
}

func historyVersionLabelOrDash(versionID int64) string {
	if versionID <= 0 {
		return "-"
	}
	return historyVersionLabel(versionID)
}

func shortHistoryTimestamp(value string) string {
	value = strings.TrimSpace(value)
	value = strings.Replace(value, "T", " ", 1)
	if len(value) >= len("2006-01-02 15:04") {
		return value[:len("2006-01-02 15:04")]
	}
	return value
}

func normalizeHistoryActionChoice(choice string) string {
	return strings.ToLower(strings.TrimSpace(choice))
}
