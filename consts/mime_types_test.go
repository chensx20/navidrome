package consts

import (
	"mime"
	"os"
	"path/filepath"
	"testing"
)

func TestEmbeddedMimeTypes(t *testing.T) {
	expectedLossless := []string{"alac", "ape", "dsf", "flac", "shn", "tak", "wav", "wv", "wvp"}
	if len(LosslessFormats) != len(expectedLossless) {
		t.Fatalf("lossless formats = %v, want %v", LosslessFormats, expectedLossless)
	}
	for index, format := range expectedLossless {
		if LosslessFormats[index] != format {
			t.Fatalf("lossless formats = %v, want %v", LosslessFormats, expectedLossless)
		}
	}

	expected := map[string]string{
		".mp3": "audio/mpeg", ".ogg": "audio/ogg", ".oga": "audio/ogg", ".opus": "audio/ogg",
		".aac": "audio/mp4", ".alac": "audio/mp4", ".m4a": "audio/mp4", ".m4b": "audio/mp4",
		".flac": "audio/flac", ".wav": "audio/x-wav", ".wma": "audio/x-ms-wma", ".ape": "audio/x-monkeys-audio",
		".mpc": "audio/x-musepack", ".shn": "audio/x-shn", ".aif": "audio/x-aiff", ".aiff": "audio/x-aiff",
		".m3u": "audio/x-mpegurl", ".pls": "audio/x-scpls", ".dsf": "audio/dsd", ".wv": "audio/x-wavpack",
		".wvp": "audio/x-wavpack", ".tak": "audio/tak", ".mka": "audio/x-matroska", ".gif": "image/gif",
		".jpg": "image/jpeg", ".jpeg": "image/jpeg", ".webp": "image/webp", ".png": "image/png",
		".bmp": "image/bmp", ".js": "text/javascript; charset=utf-8", ".css": "text/css; charset=utf-8",
	}
	for extension, want := range expected {
		if got := mime.TypeByExtension(extension); got != want {
			t.Errorf("MIME type for %s = %q, want %q", extension, got, want)
		}
	}
}

func TestLoadMimeTypesRejectsInvalidDocumentWithoutChangingState(t *testing.T) {
	before := append([]string(nil), LosslessFormats...)
	err := loadMimeTypes([]byte("audio:\n  .bad: {type: audio/test, lossless: true}\nimage: {}\nother: {}\nunknown: true\n"))
	if err == nil {
		t.Fatal("loadMimeTypes returned nil for an unknown field")
	}
	if len(LosslessFormats) != len(before) {
		t.Fatalf("lossless formats changed after failed validation: %v", LosslessFormats)
	}
	for index := range before {
		if LosslessFormats[index] != before[index] {
			t.Fatalf("lossless formats changed after failed validation: %v", LosslessFormats)
		}
	}
	if got := mime.TypeByExtension(".bad"); got != "" {
		t.Fatalf("invalid document registered .bad as %q", got)
	}
}

func TestLoadMimeTypesCustomDocument(t *testing.T) {
	t.Cleanup(func() {
		if err := loadMimeTypes(embeddedMimeTypes); err != nil {
			t.Fatalf("restore embedded MIME types: %v", err)
		}
	})
	err := loadMimeTypes([]byte("audio:\n  .customaudio: {type: audio/custom, lossless: true}\nimage:\n  .customimage: image/custom\nother:\n  .customother: application/custom\n"))
	if err != nil {
		t.Fatal(err)
	}
	if got := mime.TypeByExtension(".customaudio"); got != "audio/custom" {
		t.Fatalf("custom audio MIME type = %q", got)
	}
	if got := mime.TypeByExtension(".customimage"); got != "image/custom" {
		t.Fatalf("custom image MIME type = %q", got)
	}
	if got := mime.TypeByExtension(".customother"); got != "application/custom" {
		t.Fatalf("custom other MIME type = %q", got)
	}
	if len(LosslessFormats) != 1 || LosslessFormats[0] != "customaudio" {
		t.Fatalf("custom lossless formats = %v", LosslessFormats)
	}
}

func TestLoadMimeTypesOverride(t *testing.T) {
	t.Cleanup(func() {
		if err := loadMimeTypes(embeddedMimeTypes); err != nil {
			t.Fatalf("restore embedded MIME types: %v", err)
		}
	})
	dir := t.TempDir()
	if err := LoadMimeTypesOverride(filepath.Join(dir, "missing.yaml")); err != nil {
		t.Fatalf("missing override returned error: %v", err)
	}
	if err := LoadMimeTypesOverride(dir); err == nil {
		t.Fatal("directory override returned nil")
	}
	filename := filepath.Join(dir, "mime_types.yaml")
	if err := os.WriteFile(filename, []byte("audio: {}\nimage: {}\nother: {}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := LoadMimeTypesOverride(filename); err != nil {
		t.Fatalf("valid override returned error: %v", err)
	}
}
