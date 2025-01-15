package system

import (
	"fmt"
	"runtime"
	"strings"
)

// Name returns the current OS.
func Name() string {
	return runtime.GOOS
}

// IsMac returns true if the current OS is macOS.
func IsMac() bool {
	return Name() == "darwin"
}

// IsWindows returns true if the current OS is Windows.
func IsWindows() bool {
	return Name() == "windows"
}

// IsLinux returns true if the current OS is Linux.
func IsLinux() bool {
	return Name() == "linux"
}

// IsUnix returns true if the current OS is either macOS or Linux.
func IsUnix() bool {
	return IsMac() || IsLinux()
}

// PlatformAndArch returns the current platform and architecture.
func PlatformAndArch() string {
	platform := strings.Title(Name()) //nolint:staticcheck

	arch := strings.ToLower(runtime.GOARCH)
	if arch == "amd64" {
		arch = "x86_64"
	}

	return fmt.Sprintf("%s_%s", platform, arch)
}
