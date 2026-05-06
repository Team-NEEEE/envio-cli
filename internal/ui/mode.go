package ui

type RequestedMode string

const (
	ModeAuto  RequestedMode = "auto"
	ModePlain RequestedMode = "plain"
	ModeJSON  RequestedMode = "json"
	ModeTUI   RequestedMode = "tui"
)

func SelectMode(requested RequestedMode, isTerminal bool, env map[string]string) RequestedMode {
	switch requested {
	case ModePlain, ModeJSON, ModeTUI:
		return requested
	default:
		if env != nil && env["CI"] != "" {
			return ModePlain
		}
		if isTerminal {
			return ModeTUI
		}
		return ModePlain
	}
}
