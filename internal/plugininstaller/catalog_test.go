package plugininstaller

import "testing"

func TestAssetPattern(t *testing.T) {
	tests := []struct {
		id, goos, goarch, prefix, extension string
	}{
		{"ytmusic", "windows", "amd64", "gozik-ytmusic-plugin-windows-amd64", ".exe"},
		{"ytmusic", "darwin", "arm64", "gozik-ytmusic-plugin-macos-arm64-", ".dmg"},
		{"spotify", "linux", "amd64", "gozik-spotify-plugin-linux-amd64-", ".tar.gz"},
		{"spotify", "windows", "arm64", "gozik-spotify-plugin-windows-arm64-", ".zip"},
	}
	for _, test := range tests {
		prefix, extension := assetPattern(test.id, test.goos, test.goarch)
		if prefix != test.prefix || extension != test.extension {
			t.Errorf("assetPattern(%q, %q, %q) = %q, %q; want %q, %q", test.id, test.goos, test.goarch, prefix, extension, test.prefix, test.extension)
		}
	}
}

func TestArchivePathRejectsTraversal(t *testing.T) {
	if _, err := archivePathFor(t.TempDir(), "../outside"); err == nil {
		t.Fatal("archivePathFor accepted a traversal path")
	}
}
