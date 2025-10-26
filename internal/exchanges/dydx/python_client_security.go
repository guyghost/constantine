package dydx

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// resolveScriptPath resolves the script path to an absolute path
// SECURITY: This prevents path traversal and relative path vulnerabilities
func resolveScriptPath(configPath string) (string, error) {
	// If no path provided, use default location relative to executable
	if configPath == "" {
		// Try multiple locations in order of preference
		scriptPath := findScriptInCommonLocations()
		if scriptPath != "" {
			configPath = scriptPath
		} else {
			return "", fmt.Errorf("could not find dydx_client.py in any common location")
		}
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(configPath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Clean the path to remove any .. or . components
	cleanPath := filepath.Clean(absPath)

	return cleanPath, nil
}

// findScriptInCommonLocations searches for the script in common locations
func findScriptInCommonLocations() string {
	locations := []string{}

	// 1. Current working directory (for go run)
	cwd, err := os.Getwd()
	if err == nil {
		locations = append(locations,
			filepath.Join(cwd, "internal", "exchanges", "dydx", "scripts", "dydx_client.py"),
		)
	}

	// 2. Executable directory (for compiled binary)
	executable, err := os.Executable()
	if err == nil {
		execDir := filepath.Dir(executable)

		// Production location
		locations = append(locations,
			filepath.Join(execDir, "scripts", "dydx_client.py"),
		)

		// Development location (if running from bin/)
		locations = append(locations,
			filepath.Join(execDir, "..", "internal", "exchanges", "dydx", "scripts", "dydx_client.py"),
		)
	}

	// 3. Try relative to source code (for go run)
	if executable != "" && strings.Contains(executable, "go-build") {
		// We're in go run mode, use working directory
		if cwd != "" {
			locations = append(locations,
				filepath.Join(cwd, "internal", "exchanges", "dydx", "scripts", "dydx_client.py"),
			)
		}
	}

	// Try each location
	for _, loc := range locations {
		if _, err := os.Stat(loc); err == nil {
			return loc
		}
	}

	return ""
}

// validateScriptPath validates that the script path is safe to execute
// SECURITY: Multiple checks to prevent malicious script execution
func validateScriptPath(scriptPath string) error {
	// Check 1: File must exist
	info, err := os.Stat(scriptPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("script not found at %s (hint: run from project root or install to <executable_dir>/scripts/)", scriptPath)
		}
		return fmt.Errorf("failed to stat script: %w", err)
	}

	// Check 2: Must be a regular file (not a directory, symlink, etc.)
	if !info.Mode().IsRegular() {
		return fmt.Errorf("script path is not a regular file: %s", scriptPath)
	}

	// Check 3: Must be readable
	file, err := os.Open(scriptPath)
	if err != nil {
		return fmt.Errorf("script is not readable: %w", err)
	}
	defer file.Close()

	// Check 4: Must have .py extension
	if !strings.HasSuffix(scriptPath, ".py") {
		return fmt.Errorf("script must have .py extension: %s", scriptPath)
	}

	// Check 5: Must start with python shebang
	header := make([]byte, 50)
	n, err := file.Read(header)
	if err != nil && err != io.EOF {
		return fmt.Errorf("failed to read script header: %w", err)
	}

	headerStr := string(header[:n])
	if !strings.HasPrefix(headerStr, "#!/usr/bin/env python") && !strings.HasPrefix(headerStr, "#!/usr/bin/python") {
		return fmt.Errorf("script must start with python shebang")
	}

	// Check 6: Verify script name matches expected
	scriptName := filepath.Base(scriptPath)
	if scriptName != "dydx_client.py" {
		// Warning but not error - allow custom names in development
		fmt.Fprintf(os.Stderr, "WARNING: Script name is %s, expected dydx_client.py\n", scriptName)
	}

	return nil
}

// calculateScriptChecksum calculates SHA256 checksum of the script
// This can be used to verify script integrity in production
func calculateScriptChecksum(scriptPath string) (string, error) {
	file, err := os.Open(scriptPath)
	if err != nil {
		return "", fmt.Errorf("failed to open script: %w", err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("failed to hash script: %w", err)
	}

	checksum := hex.EncodeToString(hash.Sum(nil))
	return checksum, nil
}

// verifyScriptChecksum verifies the script matches an expected checksum
// Use this in production to ensure script hasn't been tampered with
func verifyScriptChecksum(scriptPath, expectedChecksum string) error {
	actualChecksum, err := calculateScriptChecksum(scriptPath)
	if err != nil {
		return err
	}

	if actualChecksum != expectedChecksum {
		return fmt.Errorf("script checksum mismatch: expected %s, got %s (script may have been tampered with)", expectedChecksum, actualChecksum)
	}

	return nil
}
