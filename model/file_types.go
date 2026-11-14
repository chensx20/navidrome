package model

import (
	"path/filepath"
	"slices"
	"strings"

	"github.com/navidrome/navidrome/consts"
)

var excludeAudioType = []string{
	"audio/x-mpegurl",
	"audio/x-scpls",
}

func IsAudioFile(filePath string) bool {
	extension := filepath.Ext(filePath)
	mimeType := consts.MimeTypeByExtension(extension)
	return !slices.Contains(excludeAudioType, mimeType) && strings.HasPrefix(mimeType, "audio/")
}

func IsImageFile(filePath string) bool {
	extension := filepath.Ext(filePath)
	return strings.HasPrefix(consts.MimeTypeByExtension(extension), "image/")
}

func IsValidPlaylist(filePath string) bool {
	extension := strings.ToLower(filepath.Ext(filePath))
	return extension == ".m3u" || extension == ".m3u8" || extension == ".nsp"
}
