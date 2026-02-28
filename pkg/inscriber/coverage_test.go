package inscriber

import (
	"testing"
)

func TestCovGuessMimeTypeFromName(t *testing.T) {
	cases := map[string]string{
		"file.txt":  "text/plain",
		"file.json": "application/json",
		"file.html": "text/html",
		"file.htm":  "text/html",
		"file.css":  "text/css",
		"file.js":   "application/javascript",
		"file.mjs":  "application/javascript",
		"file.ts":   "application/typescript",
		"file.tsx":  "text/tsx",
		"file.jpg":  "image/jpeg",
		"file.jpeg": "image/jpeg",
		"file.png":  "image/png",
		"file.gif":  "image/gif",
		"file.svg":  "image/svg+xml",
		"file.webp": "image/webp",
		"file.avif": "image/avif",
		"file.mp4":  "video/mp4",
		"file.webm": "video/webm",
		"file.mp3":  "audio/mpeg",
		"file.wav":  "audio/wav",
		"file.ogg":  "audio/ogg",
		"file.pdf":  "application/pdf",
		"file.zip":  "application/zip",
		"file.wasm": "application/wasm",
		"file.xyz":  "application/octet-stream",
		"":          "application/octet-stream",
	}
	for name, expected := range cases {
		got := guessMimeTypeFromName(name)
		if got != expected {
			t.Errorf("guessMimeTypeFromName(%q) = %q, want %q", name, got, expected)
		}
	}
}
