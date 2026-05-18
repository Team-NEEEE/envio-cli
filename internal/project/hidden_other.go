//go:build !windows

package project

func hideLocalEnvioDir(string) error {
	return nil
}
