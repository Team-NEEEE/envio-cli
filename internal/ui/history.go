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
	for index, history := range histories {
		latest := ""
		if history.Latest {
			latest = "latest"
		}
		_, _ = fmt.Fprintf(
			output,
			"  %d. %-4s %-16s %-10s %s\n",
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
	RenderHistoryList(output, histories, lang)
	if len(histories) == 0 {
		return "", io.EOF
	}
	if input == nil {
		return "", io.EOF
	}
	if output == nil {
		output = io.Discard
	}
	_, _ = fmt.Fprintln(output)
	_, _ = fmt.Fprint(output, "> ")

	scanner := bufio.NewScanner(input)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return "", err
		}
		return "", io.EOF
	}
	choice := strings.TrimSpace(scanner.Text())
	if index, err := strconv.Atoi(choice); err == nil && index >= 1 && index <= len(histories) {
		return historyVersionLabel(histories[index-1].VersionID), nil
	}
	return choice, nil
}

func historySelectionPrompt(lang i18n.Language) string {
	if lang == i18n.Korean {
		return "? 확인할 버전을 선택하세요."
	}
	return "? Select a version to inspect."
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
	if len(value) >= len("2006-01-02 15:04") {
		return value[:len("2006-01-02 15:04")]
	}
	return value
}
