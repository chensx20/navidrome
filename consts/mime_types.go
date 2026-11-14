package consts

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"mime"
	"os"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

type mimeType struct {
	Type     string `yaml:"type"`
	Lossless *bool  `yaml:"lossless"`
}

type mimeTypes struct {
	Audio map[string]mimeType `yaml:"audio"`
	Image map[string]string   `yaml:"image"`
	Other map[string]string   `yaml:"other"`
}

//go:embed mime_types.yaml
var embeddedMimeTypes []byte

var extensionPattern = regexp.MustCompile(`^\.[a-z0-9]+(?:[.-][a-z0-9]+)*$`)

var LosslessFormats []string

var mimeTypesByExtension map[string]string

// MimeTypeByExtension returns the configured MIME type for an extension.
func MimeTypeByExtension(extension string) string {
	return mimeTypesByExtension[strings.ToLower(extension)]
}

func init() {
	if err := loadMimeTypes(embeddedMimeTypes); err != nil {
		panic(fmt.Errorf("load embedded MIME types: %w", err))
	}
}

func LoadMimeTypesOverride(filename string) error {
	data, err := os.ReadFile(filename)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read MIME types override %q: %w", filename, err)
	}
	if err := loadMimeTypes(data); err != nil {
		return fmt.Errorf("load MIME types override %q: %w", filename, err)
	}
	return nil
}

func loadMimeTypes(data []byte) error {
	var types mimeTypes
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&types); err != nil {
		return fmt.Errorf("parse YAML: %w", err)
	}
	var extra interface{}
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return fmt.Errorf("document must contain exactly one YAML value")
	}
	if types.Audio == nil || types.Image == nil || types.Other == nil {
		return fmt.Errorf("audio, image, and other sections are required")
	}

	seen := make(map[string]struct{}, len(types.Audio)+len(types.Image)+len(types.Other))
	for extension, value := range types.Audio {
		if err := validateExtension(extension, seen); err != nil {
			return err
		}
		if err := validateMIME(extension, value.Type); err != nil {
			return err
		}
		if value.Lossless == nil {
			return fmt.Errorf("missing lossless field for %s", extension)
		}
	}
	for extension, value := range types.Image {
		if err := validateExtension(extension, seen); err != nil {
			return err
		}
		if err := validateMIME(extension, value); err != nil {
			return err
		}
	}
	for extension, value := range types.Other {
		if err := validateExtension(extension, seen); err != nil {
			return err
		}
		if err := validateMIME(extension, value); err != nil {
			return err
		}
	}

	registered := make(map[string]string, len(seen))
	for extension, value := range types.Audio {
		registered[extension] = formatMIME(value.Type)
	}
	for extension, value := range types.Image {
		registered[extension] = formatMIME(value)
	}
	for extension, value := range types.Other {
		registered[extension] = formatMIME(value)
	}

	lossless := make([]string, 0)
	for extension, value := range types.Audio {
		if *value.Lossless {
			lossless = append(lossless, strings.TrimPrefix(extension, "."))
		}
	}
	sort.Strings(lossless)
	mimeTypesByExtension = registered
	LosslessFormats = lossless
	return nil
}

func formatMIME(value string) string {
	if strings.HasPrefix(value, "text/") {
		return mime.FormatMediaType(value, map[string]string{"charset": "utf-8"})
	}
	return value
}

func validateExtension(extension string, seen map[string]struct{}) error {
	if !extensionPattern.MatchString(extension) {
		return fmt.Errorf("invalid extension %q", extension)
	}
	if _, exists := seen[extension]; exists {
		return fmt.Errorf("duplicate extension %q", extension)
	}
	seen[extension] = struct{}{}
	return nil
}

func validateMIME(extension, value string) error {
	if value == "" || strings.TrimSpace(value) != value {
		return fmt.Errorf("invalid MIME type for %s: %q", extension, value)
	}
	parsed, _, err := mime.ParseMediaType(value)
	mediaType, mediaSubtype, hasSubtype := strings.Cut(parsed, "/")
	if err != nil || parsed != value || !hasSubtype || mediaType == "" || mediaSubtype == "" {
		return fmt.Errorf("invalid MIME type for %s: %q", extension, value)
	}
	return nil
}
