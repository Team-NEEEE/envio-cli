package config

import (
	"bufio"
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

func ParseDotenv(raw []byte) (map[string]string, error) {
	values := map[string]string{}
	scanner := bufio.NewScanner(bytes.NewReader(raw))
	scanner.Buffer(make([]byte, 1024), 1024*1024)

	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(strings.TrimSuffix(scanner.Text(), "\r"))
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "export ") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
		}

		separator := strings.Index(line, "=")
		if separator < 0 {
			return nil, fmt.Errorf("line %d: expected KEY=VALUE", lineNumber)
		}

		key := strings.TrimSpace(line[:separator])
		if !validEnvKey(key) {
			return nil, fmt.Errorf("line %d: invalid environment variable name %q", lineNumber, key)
		}

		value, err := parseDotenvValue(strings.TrimSpace(line[separator+1:]))
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", lineNumber, err)
		}
		values[key] = value
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read dotenv content: %w", err)
	}
	return values, nil
}

func validEnvKey(key string) bool {
	if key == "" {
		return false
	}
	for index, r := range key {
		if index == 0 {
			if r != '_' && !unicode.IsLetter(r) {
				return false
			}
			continue
		}
		if r != '_' && !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

func parseDotenvValue(raw string) (string, error) {
	if raw == "" {
		return "", nil
	}

	switch raw[0] {
	case '"':
		if !strings.HasSuffix(raw, "\"") {
			return "", fmt.Errorf("unterminated double-quoted value")
		}
		value, err := strconv.Unquote(raw)
		if err != nil {
			return "", fmt.Errorf("invalid double-quoted value: %w", err)
		}
		return value, nil
	case '\'':
		if !strings.HasSuffix(raw, "'") || len(raw) == 1 {
			return "", fmt.Errorf("unterminated single-quoted value")
		}
		return raw[1 : len(raw)-1], nil
	default:
		return strings.TrimSpace(stripInlineComment(raw)), nil
	}
}

func stripInlineComment(raw string) string {
	for index, r := range raw {
		if r == '#' && index > 0 {
			previous, _ := utf8LastRune(raw[:index])
			if unicode.IsSpace(previous) {
				return raw[:index]
			}
		}
	}
	return raw
}

func utf8LastRune(raw string) (rune, bool) {
	if raw == "" {
		return 0, false
	}
	r, _ := utf8.DecodeLastRuneInString(raw)
	if r == utf8.RuneError {
		return 0, false
	}
	return r, true
}
