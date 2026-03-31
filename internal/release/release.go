package release

import (
	"fmt"
	"strings"
)

type Platform struct {
	OS   string
	Arch string
}

func NormalizeUname(unameOS, unameArch string) (Platform, error) {
	var platform Platform

	switch strings.ToLower(strings.TrimSpace(unameOS)) {
	case "linux":
		platform.OS = "linux"
	case "darwin":
		platform.OS = "darwin"
	default:
		return Platform{}, fmt.Errorf("unsupported OS %q", unameOS)
	}

	switch strings.ToLower(strings.TrimSpace(unameArch)) {
	case "x86_64", "amd64":
		platform.Arch = "amd64"
	case "arm64", "aarch64":
		platform.Arch = "arm64"
	default:
		return Platform{}, fmt.Errorf("unsupported architecture %q", unameArch)
	}

	return platform, nil
}

func AssetName(version string, platform Platform) (string, error) {
	if strings.TrimSpace(version) == "" {
		return "", fmt.Errorf("version must not be empty")
	}
	if platform.OS == "" || platform.Arch == "" {
		return "", fmt.Errorf("platform must include OS and architecture")
	}

	return fmt.Sprintf("openclaw-lastpass_%s_%s_%s.tar.gz", version, platform.OS, platform.Arch), nil
}
