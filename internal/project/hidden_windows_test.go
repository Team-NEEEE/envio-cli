//go:build windows

package project

import (
	"path/filepath"
	"syscall"
	"testing"
)

func TestSaveLocalLinkConfigHidesEnvioDirectory(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if err := saveLocalLinkConfig(root, validLocalLinkConfigForTest()); err != nil {
		t.Fatalf("saveLocalLinkConfig() error = %v", err)
	}

	path, err := syscall.UTF16PtrFromString(filepath.Join(root, localLinkConfigDir))
	if err != nil {
		t.Fatalf("UTF16PtrFromString() error = %v", err)
	}
	attrs, err := syscall.GetFileAttributes(path)
	if err != nil {
		t.Fatalf("GetFileAttributes(.envio) error = %v", err)
	}
	if attrs&syscall.FILE_ATTRIBUTE_HIDDEN == 0 {
		t.Fatalf(".envio attributes = %#x, want hidden", attrs)
	}
}
