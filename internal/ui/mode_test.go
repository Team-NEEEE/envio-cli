package ui

import "testing"

func TestSelectMode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		requested  RequestedMode
		isTerminal bool
		env        map[string]string
		want       RequestedMode
	}{
		{name: "auto terminal uses tui", requested: ModeAuto, isTerminal: true, want: ModeTUI},
		{name: "auto pipe uses plain", requested: ModeAuto, isTerminal: false, want: ModePlain},
		{name: "ci uses plain", requested: ModeAuto, isTerminal: true, env: map[string]string{"CI": "true"}, want: ModePlain},
		{name: "plain forced", requested: ModePlain, isTerminal: true, want: ModePlain},
		{name: "json forced", requested: ModeJSON, isTerminal: true, want: ModeJSON},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := SelectMode(tt.requested, tt.isTerminal, tt.env); got != tt.want {
				t.Fatalf("SelectMode() = %q, want %q", got, tt.want)
			}
		})
	}
}
