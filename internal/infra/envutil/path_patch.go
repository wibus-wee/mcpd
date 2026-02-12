package envutil

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

const (
	skipPathPatchEnv = "MCPV_SKIP_PATH_PATCH"
	debugPathPatch   = "MCPV_DEBUG_PATH_PATCH"
	termEnv          = "TERM"
	shellEnv         = "SHELL"
	pathEnv          = "PATH"
	pathMarker       = "__MCPV_PATH__"
	pathCommand      = "printf '" + pathMarker + "'; /usr/bin/printenv PATH; printf '" + pathMarker + "'"
)

type pathCacheEntry struct {
	path string
	err  error
}

var loginPathCache sync.Map
var userShellCache sync.Map

// PatchProcessPATHIfNeeded updates the current process PATH for GUI-launched processes on macOS.
func PatchProcessPATHIfNeeded() {
	if runtime.GOOS != "darwin" {
		return
	}
	env := os.Environ()
	patched := patchPATH(env, true)
	currentPath := envVarValue(env, pathEnv)
	patchedPath := envVarValue(patched, pathEnv)
	if patchedPath == "" || patchedPath == currentPath {
		debugPathPatchf("process PATH unchanged, current=%q patched=%q", currentPath, patchedPath)
		return
	}
	_ = os.Setenv(pathEnv, patchedPath)
	debugPathPatchf("process PATH updated, length=%d", len(patchedPath))
}

// PatchPATHIfNeeded updates PATH for GUI-launched processes on macOS.
func PatchPATHIfNeeded(env []string) []string {
	return patchPATH(env, false)
}

func patchPATH(env []string, ignoreTerm bool) []string {
	if runtime.GOOS != "darwin" {
		return env
	}
	if strings.TrimSpace(envVarValue(env, skipPathPatchEnv)) != "" {
		return env
	}
	if !ignoreTerm && strings.TrimSpace(envVarValue(env, termEnv)) != "" {
		return env
	}
	shellPath := resolveShellPath(env)
	debugPathPatchf("shell path resolved: %q", shellPath)
	loginPath, err := loginShellPATH(shellPath)
	if err != nil || strings.TrimSpace(loginPath) == "" {
		debugPathPatchf("shell path resolve failed: %v", err)
		return env
	}
	currentPath := envVarValue(env, pathEnv)
	mergedPath := mergePATH(loginPath, currentPath)
	if mergedPath == "" || mergedPath == currentPath {
		debugPathPatchf("merged PATH unchanged, current=%q", currentPath)
		return env
	}
	debugPathPatchf("merged PATH applied, length=%d", len(mergedPath))
	return setEnvValue(env, pathEnv, mergedPath)
}

func envVarValue(env []string, key string) string {
	if key == "" {
		return ""
	}
	prefix := key + "="
	var value string
	for _, entry := range env {
		if strings.HasPrefix(entry, prefix) {
			value = strings.TrimPrefix(entry, prefix)
		}
	}
	return value
}

func setEnvValue(env []string, key, value string) []string {
	if key == "" {
		return env
	}
	prefix := key + "="
	out := make([]string, 0, len(env)+1)
	for _, entry := range env {
		if strings.HasPrefix(entry, prefix) {
			continue
		}
		out = append(out, entry)
	}
	out = append(out, prefix+value)
	return out
}

func resolveShellPath(env []string) string {
	shellPath, err := userShellPath()
	if err == nil && strings.TrimSpace(shellPath) != "" {
		return shellPath
	}
	shellPath = strings.TrimSpace(envVarValue(env, shellEnv))
	if shellPath != "" {
		return shellPath
	}
	return "/bin/zsh"
}

func userShellPath() (string, error) {
	if cached, ok := userShellCache.Load("user"); ok {
		entry := cached.(pathCacheEntry)
		return entry.path, entry.err
	}
	path, err := lookupUserShellPath()
	userShellCache.Store("user", pathCacheEntry{path: path, err: err})
	return path, err
}

func lookupUserShellPath() (string, error) {
	userName := strings.TrimSpace(os.Getenv("USER"))
	if userName == "" {
		current, err := user.Current()
		if err == nil && current != nil {
			userName = current.Username
		}
	}
	if userName == "" {
		return "", fmt.Errorf("user not found")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	output, err := exec.CommandContext(ctx, "dscl", ".", "-read", "/Users/"+userName, "UserShell").Output()
	if err != nil {
		return "", err
	}
	line := strings.TrimSpace(string(output))
	if line == "" {
		return "", fmt.Errorf("user shell not found")
	}
	idx := strings.LastIndex(line, ":")
	if idx == -1 {
		return "", fmt.Errorf("unexpected dscl output: %s", line)
	}
	return strings.TrimSpace(line[idx+1:]), nil
}

func loginShellPATH(shellPath string) (string, error) {
	if cached, ok := loginPathCache.Load(shellPath); ok {
		entry := cached.(pathCacheEntry)
		return entry.path, entry.err
	}
	path, err := resolveLoginShellPATH(shellPath)
	loginPathCache.Store(shellPath, pathCacheEntry{path: path, err: err})
	return path, err
}

func resolveLoginShellPATH(shellPath string) (string, error) {
	path, err := resolveShellPATH(shellPath, true)
	if err == nil && strings.TrimSpace(path) != "" {
		return path, nil
	}
	fallback, fallbackErr := resolveShellPATH(shellPath, false)
	if fallbackErr == nil && strings.TrimSpace(fallback) != "" {
		return fallback, nil
	}
	if err == nil {
		err = fallbackErr
	}
	return path, err
}

func resolveShellPATH(shellPath string, interactive bool) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	args := shellArgs(shellPath, interactive)
	cmd := exec.CommandContext(ctx, shellPath, append(args, pathCommand)...)
	cmd.Env = append(os.Environ(), "LANG=C", "LC_ALL=C")
	output, err := cmd.Output()
	if err != nil {
		debugPathPatchf("shell PATH command failed: %v", err)
		return "", err
	}
	return strings.TrimSpace(extractMarkedPath(string(output))), nil
}

func shellArgs(shellPath string, interactive bool) []string {
	if interactive && supportsInteractiveShell(shellPath) {
		return []string{"-i", "-l", "-c"}
	}
	return []string{"-l", "-c"}
}

func supportsInteractiveShell(shellPath string) bool {
	name := strings.ToLower(filepath.Base(shellPath))
	return name == "zsh" || name == "bash" || name == "fish"
}

func mergePATH(primary, fallback string) string {
	separator := string(os.PathListSeparator)
	seen := map[string]struct{}{}
	out := make([]string, 0, 8)

	appendPath := func(path string) {
		if strings.TrimSpace(path) == "" {
			return
		}
		for _, entry := range strings.Split(path, separator) {
			entry = strings.TrimSpace(entry)
			if entry == "" {
				continue
			}
			if _, exists := seen[entry]; exists {
				continue
			}
			seen[entry] = struct{}{}
			out = append(out, entry)
		}
	}

	appendPath(primary)
	appendPath(fallback)

	if len(out) == 0 {
		return ""
	}
	return strings.Join(out, separator)
}

func extractMarkedPath(output string) string {
	start := strings.Index(output, pathMarker)
	if start == -1 {
		return ""
	}
	rest := output[start+len(pathMarker):]
	end := strings.Index(rest, pathMarker)
	if end == -1 {
		return ""
	}
	return rest[:end]
}

func debugPathPatchf(format string, args ...any) {
	if strings.TrimSpace(os.Getenv(debugPathPatch)) == "" {
		return
	}
	_, _ = fmt.Fprintf(os.Stderr, "path_patch: "+format+"\n", args...)
}
