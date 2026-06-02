package utils

import (
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

var (
	mediaRootOnce sync.Once
	mediaRoot     string
)

// MediaRoot returns the absolute directory used to store and serve uploaded files.
func MediaRoot() string {
	mediaRootOnce.Do(func() {
		mediaRoot = resolveMediaRoot()
	})
	return mediaRoot
}

// MediaPath joins path elements under MediaRoot.
func MediaPath(elem ...string) string {
	parts := append([]string{MediaRoot()}, elem...)
	return filepath.Join(parts...)
}

// EnsureMediaRoot creates the media root if it does not already exist.
func EnsureMediaRoot() error {
	return os.MkdirAll(MediaRoot(), 0755)
}

func resolveMediaRoot() string {
	if root := os.Getenv("MEDIA_ROOT"); root != "" {
		if absRoot, err := filepath.Abs(root); err == nil {
			return absRoot
		}
		return root
	}

	if projectRoot, ok := findProjectRoot(); ok {
		return filepath.Join(projectRoot, "media")
	}

	if cwd, err := os.Getwd(); err == nil {
		return filepath.Join(cwd, "media")
	}

	return "media"
}

func findProjectRoot() (string, bool) {
	candidates := make([]string, 0, 3)

	if cwd, err := os.Getwd(); err == nil {
		candidates = append(candidates, cwd)
	}

	if exe, err := os.Executable(); err == nil {
		candidates = append(candidates, filepath.Dir(exe))
	}

	if _, file, _, ok := runtime.Caller(0); ok {
		candidates = append(candidates, filepath.Dir(file))
	}

	for _, candidate := range candidates {
		if root, ok := searchUpForGoMod(candidate); ok {
			return root, true
		}
	}

	return "", false
}

func searchUpForGoMod(start string) (string, bool) {
	dir, err := filepath.Abs(start)
	if err != nil {
		return "", false
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, true
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false
		}
		dir = parent
	}
}
