package i18n

import "testing"

func TestErrorTextUsesEmbeddedCatalog(t *testing.T) {
	t.Parallel()

	korean := ErrorText(Korean, "CONFIG_REQUIRED", "fallback message", "fallback hint")
	if korean.Message != "필수 설정이 누락되었습니다." {
		t.Fatalf("Korean message = %q", korean.Message)
	}
	if korean.Hint != "필수 설정을 확인한 뒤 명령을 다시 실행해 주세요." {
		t.Fatalf("Korean hint = %q", korean.Hint)
	}

	english := ErrorText(English, "CONFIG_REQUIRED", "fallback message", "fallback hint")
	if english.Message != "Required configuration is missing." {
		t.Fatalf("English message = %q", english.Message)
	}
	if english.Hint != "Check the required configuration and run the command again." {
		t.Fatalf("English hint = %q", english.Hint)
	}
}

func TestFallbackMessageWhenCatalogKeyIsMissing(t *testing.T) {
	t.Parallel()

	text := ErrorText(Korean, "NOT_REGISTERED", "fallback message", "fallback hint")
	if text.Message != "fallback message" {
		t.Fatalf("message fallback = %q", text.Message)
	}
	if text.Hint != "fallback hint" {
		t.Fatalf("hint fallback = %q", text.Hint)
	}
}

func TestLanguageResolutionSupportsBCP47Variants(t *testing.T) {
	t.Parallel()

	if lang, ok := Resolve(""); !ok || lang != English {
		t.Fatalf("Resolve(empty) = %q, %v", lang, ok)
	}
	if lang, ok := Resolve("ko-KR"); !ok || lang != Korean {
		t.Fatalf("Resolve(ko-KR) = %q, %v", lang, ok)
	}
	if lang, ok := Resolve("en-US"); !ok || lang != English {
		t.Fatalf("Resolve(en-US) = %q, %v", lang, ok)
	}
	if lang, ok := Resolve("ja"); ok || lang != English {
		t.Fatalf("Resolve(ja) = %q, %v, want English fallback with unsupported=false", lang, ok)
	}
}
