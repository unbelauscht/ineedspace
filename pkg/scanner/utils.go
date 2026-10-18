package scanner

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// UserHomeDir returns the current user's home directory.
func UserHomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return os.Getenv("HOME")
	}
	return home
}

// ExpandPath expands ~ to the home directory.
func ExpandPath(p string) string {
	if strings.HasPrefix(p, "~/") {
		return filepath.Join(UserHomeDir(), p[2:])
	}
	return p
}

// DirExists returns true if the path exists and is a directory.
func DirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// PathExists returns true if the path exists.
func PathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// DirSize calculates the total size and file count of a directory.
func DirSize(path string) (int64, int, error) {
	var size int64
	var count int

	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // ignore permission errors on individual files
		}
		if info != nil && !info.IsDir() {
			size += info.Size()
			count++
		}
		return nil
	})

	return size, count, err
}

// DirSizeParallel calculates directory sizes in parallel for a slice of subdirectories.
func DirSizeParallel(paths []string) map[string]int64 {
	results := make(map[string]int64)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, p := range paths {
		wg.Add(1)
		go func(target string) {
			defer wg.Done()
			sz, _, _ := DirSize(target)
			mu.Lock()
			results[target] = sz
			mu.Unlock()
		}(p)
	}

	wg.Wait()
	return results
}

// CommandExists checks if a binary is available in PATH.
func CommandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

// RunCmdWithTimeout runs an external command with a timeout.
func RunCmdWithTimeout(timeout time.Duration, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Env = os.Environ()

	done := make(chan error, 1)
	var out []byte

	go func() {
		var err error
		out, err = cmd.CombinedOutput()
		done <- err
	}()

	select {
	case <-time.After(timeout):
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		return "", os.ErrDeadlineExceeded
	case err := <-done:
		return strings.TrimSpace(string(out)), err
	}
}
