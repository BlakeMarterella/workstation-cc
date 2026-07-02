// Package osdetect is the single source of truth for operating system and
// architecture detection. Callers should never branch on runtime.GOOS
// directly; they ask osdetect once and pass the result down.
package osdetect

import (
	"fmt"
	"os"
	"runtime"
	"strings"
)

// OS is a supported operating system.
type OS string

const (
	Darwin  OS = "darwin"
	Linux   OS = "linux"
	Windows OS = "windows"
)

// Arch is a supported CPU architecture.
type Arch string

const (
	AMD64 Arch = "amd64"
	ARM64 Arch = "arm64"
)

// Info captures everything the rest of the program needs to know about the
// host platform.
type Info struct {
	OS   OS
	Arch Arch
	// LinuxID is the /etc/os-release ID field (e.g. "ubuntu", "fedora") on
	// Linux; empty on other platforms. Used to pick apt vs dnf.
	LinuxID string
}

// Detect inspects the running host and returns its Info, or an error if the
// platform is unsupported.
func Detect() (Info, error) {
	osys, err := parseOS(runtime.GOOS)
	if err != nil {
		return Info{}, err
	}
	arch, err := parseArch(runtime.GOARCH)
	if err != nil {
		return Info{}, err
	}

	info := Info{OS: osys, Arch: arch}
	if osys == Linux {
		if content, err := os.ReadFile("/etc/os-release"); err == nil {
			info.LinuxID = parseOSReleaseID(string(content))
		}
	}
	return info, nil
}

// parseOS maps a runtime.GOOS string to a supported OS.
func parseOS(goos string) (OS, error) {
	switch goos {
	case "darwin":
		return Darwin, nil
	case "linux":
		return Linux, nil
	case "windows":
		return Windows, nil
	default:
		return "", fmt.Errorf("unsupported operating system: %q", goos)
	}
}

// parseArch maps a runtime.GOARCH string to a supported Arch.
func parseArch(goarch string) (Arch, error) {
	switch goarch {
	case "amd64":
		return AMD64, nil
	case "arm64":
		return ARM64, nil
	default:
		return "", fmt.Errorf("unsupported architecture: %q", goarch)
	}
}

// parseOSReleaseID extracts the ID= value from /etc/os-release content,
// stripping surrounding quotes. Returns "" when absent.
func parseOSReleaseID(content string) string {
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "ID=") {
			continue
		}
		val := strings.TrimPrefix(line, "ID=")
		return strings.Trim(val, "\"'")
	}
	return ""
}
