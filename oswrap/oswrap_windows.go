//+build windows

package oswrap

import (
	"path/filepath"
	"os"
	"strings"
)

func NormPath(path string) (string, error) {
	if strings.HasPrefix(path, "\\\\?\\") {
		return path, nil
	}

	path, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	path = filepath.Clean(path)
	return "\\\\?\\" + path, nil
}

func Open(name string) (*os.File, error) {
	name, err := NormPath(name)
	if err != nil {
		return nil, err
	}
	return os.Open(name)
}

func Create(name string) (*os.File, error) {
	name, err := NormPath(name)
	if err != nil {
		return nil, err
	}
	return os.Create(name)
}

func OpenFile(name string, flag int, perm os.FileMode) (*os.File, error) {
	name, err := NormPath(name)
	if err != nil {
		return nil, err
	}
	return os.OpenFile(name, flag, perm)
}

func Remove(name string) error {
	name, err := NormPath(name)
	if err != nil {
		return nil
	}
	return os.Remove(name)
}

func RemoveAll(name string) error {
	name, err := NormPath(name)
	if err != nil {
		return nil
	}
	return os.RemoveAll(name)
}

func Mkdir(name string, mode os.FileMode) error {
	name, err := NormPath(name)
	if err != nil {
		return err
	}
	return os.Mkdir(name, mode)
}

func MkdirAll(name string, mode os.FileMode) error {
	name, err := NormPath(name)
	if err != nil {
		return err
	}
	return os.MkdirAll(name, mode)
}

func Rename(oldpath, newpath string) error {
	oldpath, err := NormPath(oldpath)
	if err != nil {
		return err
	}
	newpath, err = NormPath(newpath)
	if err != nil {
		return err
	}
	return os.Rename(oldpath, newpath)
}


func Lstat(name string) (os.FileInfo, error) {
	name, err := NormPath(name)
	if err != nil {
		return nil, err
	}
	return os.Lstat(name)
}

func Stat(name string) (os.FileInfo, error) {
	name, err := NormPath(name)
	if err != nil {
		return nil, err
	}
	return os.Stat(name)
}

func Walk(root string, walkFn filepath.WalkFunc) error {
	newroot, err := NormPath(root)
	if err != nil {
		return err
	}
	return filepath.Walk(newroot, func(path string, info os.FileInfo, err error) error {
		oldpath := root + strings.TrimPrefix(path, newroot)
		return walkFn(oldpath, info, err)
	})
}
