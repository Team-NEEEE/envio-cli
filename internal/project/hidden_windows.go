//go:build windows

package project

import "syscall"

func hideLocalEnvioDir(path string) error {
	name, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return err
	}
	attrs, err := syscall.GetFileAttributes(name)
	if err != nil {
		return err
	}
	return syscall.SetFileAttributes(name, attrs|syscall.FILE_ATTRIBUTE_HIDDEN)
}
