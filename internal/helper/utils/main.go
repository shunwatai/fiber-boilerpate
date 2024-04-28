package utils

import (
	// "golang-api-starter/internal/config"
	"crypto/rand"
	"fmt"
	"path"
	"path/filepath"
	"runtime"
)

// ToPtr uses to return the pointer of the value without using one more line to declare a variable
// e.g.: helper.ToPtr("some string") returns the address of "some string"
func ToPtr[T any](v T) *T {
	return &v
}

// Deref get the value of pointer. I use it as a helper func in html template.
func Deref[T any](v *T) T {
	return *v
}

// RootDir get the project base path
// ref: https://stackoverflow.com/a/58294680
func RootDir(level int) string {
	parentPath := ""
	for p := 0; p < level; p++ {
		parentPath += "../"
	}
	_, b, _, _ := runtime.Caller(0)
	d := path.Join(path.Dir(b), parentPath)
	return filepath.Dir(d)
}

// GetRandString generate random string by given length
// ref: https://gist.github.com/arxdsilva/8caeca47b126a290c4562a25464895e8
func GetRandString(length int) string {
	if length < 1{
		length = 1
	}
	b := make([]byte, length)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}
