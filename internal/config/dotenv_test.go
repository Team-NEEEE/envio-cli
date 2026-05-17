package config

import "testing"

func TestParseDotenvParsesCommonValues(t *testing.T) {
	t.Parallel()

	got, err := ParseDotenv([]byte(`
# comment
DATABASE_URL=postgres://localhost/db
EMPTY=
export API_KEY="line\nvalue"
PASSWORD='literal#value'
INLINE=value # comment
HASH=value#kept
`))
	if err != nil {
		t.Fatalf("ParseDotenv() error = %v", err)
	}

	if got["DATABASE_URL"] != "postgres://localhost/db" ||
		got["EMPTY"] != "" ||
		got["API_KEY"] != "line\nvalue" ||
		got["PASSWORD"] != "literal#value" ||
		got["INLINE"] != "value" ||
		got["HASH"] != "value#kept" {
		t.Fatalf("ParseDotenv() = %#v", got)
	}
}

func TestParseDotenvRejectsInvalidLine(t *testing.T) {
	t.Parallel()

	if _, err := ParseDotenv([]byte("1BAD=value\n")); err == nil {
		t.Fatal("ParseDotenv() error = nil, want invalid key error")
	}
}
