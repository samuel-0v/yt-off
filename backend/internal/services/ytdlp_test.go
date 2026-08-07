package services

import (
	"os"
	"path/filepath"
	"testing"
)

func TestYTDLPOptionsCommandPrefix(t *testing.T) {
	cookiesFile := filepath.Join(t.TempDir(), "youtube.txt")
	if err := os.WriteFile(cookiesFile, []byte("cookies"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	command := YTDLPOptions{
		JSRuntime:   "node",
		CookiesFile: cookiesFile,
	}.commandPrefix()

	want := []string{"yt-dlp", "--js-runtimes", "node", "--cookies", cookiesFile}
	if len(command) != len(want) {
		t.Fatalf("command = %#v, want %#v", command, want)
	}
	for index := range want {
		if command[index] != want[index] {
			t.Fatalf("command[%d] = %q, want %q", index, command[index], want[index])
		}
	}
}

func TestYTDLPOptionsSkipsMissingCookiesFile(t *testing.T) {
	command := YTDLPOptions{
		JSRuntime:   "node",
		CookiesFile: filepath.Join(t.TempDir(), "missing.txt"),
	}.commandPrefix()

	want := []string{"yt-dlp", "--js-runtimes", "node"}
	if len(command) != len(want) {
		t.Fatalf("command = %#v, want %#v", command, want)
	}
	for index := range want {
		if command[index] != want[index] {
			t.Fatalf("command[%d] = %q, want %q", index, command[index], want[index])
		}
	}
}
