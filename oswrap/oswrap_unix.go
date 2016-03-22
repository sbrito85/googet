//+build linux darwin

package oswrap

import (
	"path/filepath"
	"os"
)

func Open(name string) (*os.File, error) {
	return os.Open(name)
}

func Create(name string) (*os.File, error) {
	return os.Create(name)
}

func OpenFile(name string, flag int, perm os.FileMode) (*os.File, error) {
	return os.OpenFile(name, flag, perm)
}

func Remove(name string) error {
	return os.Remove(name)
}

func RemoveAll(name string) error {
	return os.RemoveAll(name)
}

func Mkdir(name string, mode os.FileMode) error {
	return os.Mkdir(name, mode)
}

func MkdirAll(name string, mode os.FileMode) error {
	return os.MkdirAll(name, mode)
}

func Rename(oldpath, newpath string) error {
	return os.Rename(oldpath, newpath)
}


func Lstat(name string) (os.FileInfo, error) {
	return os.Lstat(name)
}

func Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

func Walk(root string, walkFn filepath.WalkFunc) error {
	return filepath.Walk(root, walkFn)
}
