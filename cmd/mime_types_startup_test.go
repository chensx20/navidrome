package cmd

import (
	"mime"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestMimeTypesStartupHelper(t *testing.T) {
	if os.Getenv("NAVIDROME_MIME_STARTUP_HELPER") != "1" {
		return
	}
	inspectCmd.Run = func(*cobra.Command, []string) {
		_, _ = os.Stdout.WriteString("inspect-ran\n")
	}
	rootCmd.SetArgs([]string{"--nobanner", "--datafolder", os.Getenv("NAVIDROME_TEST_DATA_FOLDER"), "inspect", "placeholder.mp3"})
	err := rootCmd.Execute()
	if err != nil {
		_, _ = os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(3)
	}
	_, _ = os.Stdout.WriteString("mime=" + mime.TypeByExtension(".startup") + "\n")
}

func TestMimeTypesStartupOverrides(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(string) error
		wantOutput string
		wantError  string
	}{
		{name: "absent", wantOutput: "inspect-ran\nmime=\n"},
		{
			name: "valid configured folder replacement",
			setup: func(filename string) error {
				return os.WriteFile(filename, []byte("audio: { .startup: {type: audio/startup, lossless: false} }\nimage: {}\nother: {}\n"), 0o600)
			},
			wantOutput: "inspect-ran\nmime=audio/startup",
		},
		{
			name: "malformed",
			setup: func(filename string) error {
				return os.WriteFile(filename, []byte("audio: ["), 0o600)
			},
			wantError: "error loading MIME types",
		},
		{
			name:      "unreadable path",
			setup:     func(filename string) error { return os.Mkdir(filename, 0o700) },
			wantError: "read MIME types override",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dataFolder := t.TempDir()
			if test.setup != nil {
				resourceFolder := filepath.Join(dataFolder, "resources")
				if err := os.Mkdir(resourceFolder, 0o700); err != nil {
					t.Fatal(err)
				}
				if err := test.setup(filepath.Join(resourceFolder, "mime_types.yaml")); err != nil {
					t.Fatal(err)
				}
			}
			command := exec.Command("go", "test", "-run=^TestMimeTypesStartupHelper$", "-v", ".")
			command.Env = append(os.Environ(), "NAVIDROME_MIME_STARTUP_HELPER=1", "NAVIDROME_TEST_DATA_FOLDER="+dataFolder)
			output, err := command.CombinedOutput()
			if test.wantError == "" && err != nil {
				t.Fatalf("startup failed: %v\n%s", err, output)
			}
			if test.wantError != "" && (err == nil || !strings.Contains(string(output), test.wantError)) {
				t.Fatalf("startup error = %v, output %q; want %q", err, output, test.wantError)
			}
			if test.wantError != "" && strings.Contains(string(output), "inspect-ran") {
				t.Fatalf("inspect ran after startup error: %s", output)
			}
			if test.wantOutput != "" && !strings.Contains(string(output), test.wantOutput) {
				t.Fatalf("startup output %q does not contain %q", output, test.wantOutput)
			}
		})
	}
}
