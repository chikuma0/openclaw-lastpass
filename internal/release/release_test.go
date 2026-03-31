package release

import "testing"

func TestNormalizeUname(t *testing.T) {
	t.Parallel()

	platform, err := NormalizeUname("Linux", "x86_64")
	if err != nil {
		t.Fatalf("NormalizeUname() error = %v", err)
	}
	if platform.OS != "linux" || platform.Arch != "amd64" {
		t.Fatalf("NormalizeUname() = %#v, want linux/amd64", platform)
	}
}

func TestNormalizeUnameRejectsUnsupportedValues(t *testing.T) {
	t.Parallel()

	if _, err := NormalizeUname("Windows", "amd64"); err == nil {
		t.Fatal("NormalizeUname() error = nil, want unsupported OS error")
	}
}

func TestAssetName(t *testing.T) {
	t.Parallel()

	asset, err := AssetName("v1.2.3", Platform{OS: "darwin", Arch: "arm64"})
	if err != nil {
		t.Fatalf("AssetName() error = %v", err)
	}
	if asset != "openclaw-lastpass_v1.2.3_darwin_arm64.tar.gz" {
		t.Fatalf("AssetName() = %q, want expected asset name", asset)
	}
}
