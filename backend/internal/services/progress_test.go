package services

import "testing"

func TestParseDownloadProgressPercentOnly(t *testing.T) {
	info, ok := ParseDownloadProgress("[download] 10% of 100MiB")
	if !ok {
		t.Fatal("expected progress line to be parsed")
	}
	if info.Progress != 10 {
		t.Fatalf("Progress = %v, want 10", info.Progress)
	}
	if info.Speed != "" {
		t.Fatalf("Speed = %q, want empty", info.Speed)
	}
	if info.ETA != "" {
		t.Fatalf("ETA = %q, want empty", info.ETA)
	}
}

func TestParseDownloadProgressWithSpeedAndETA(t *testing.T) {
	info, ok := ParseDownloadProgress("[download] 50.5% of 100MiB at 5MiB/s ETA 00:10")
	if !ok {
		t.Fatal("expected progress line to be parsed")
	}
	if info.Progress != 50.5 {
		t.Fatalf("Progress = %v, want 50.5", info.Progress)
	}
	if info.Speed != "5MiB/s" {
		t.Fatalf("Speed = %q, want 5MiB/s", info.Speed)
	}
	if info.ETA != "00:10" {
		t.Fatalf("ETA = %q, want 00:10", info.ETA)
	}
}

func TestParseDownloadProgressCompleted(t *testing.T) {
	info, ok := ParseDownloadProgress("[download] 100% of 100MiB")
	if !ok {
		t.Fatal("expected progress line to be parsed")
	}
	if info.Progress != 100 {
		t.Fatalf("Progress = %v, want 100", info.Progress)
	}
}
