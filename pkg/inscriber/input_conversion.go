package inscriber

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func buildStartInscriptionRequest(
	input InscriptionInput,
	accountID string,
	network Network,
	options InscriptionOptions,
) (StartInscriptionRequest, error) {
	mode := options.Mode
	if mode == "" {
		mode = ModeFile
	}

	request := StartInscriptionRequest{
		HolderID:           strings.TrimSpace(accountID),
		Mode:               mode,
		Network:            network,
		Metadata:           options.Metadata,
		Tags:               options.Tags,
		Creator:            strings.TrimSpace(stringOrDefault(options.Metadata, "creator", "")),
		Description:        strings.TrimSpace(stringOrDefault(options.Metadata, "description", "")),
		FileStandard:       strings.TrimSpace(options.FileStandard),
		ChunkSize:          options.ChunkSize,
		OnlyJSONCollection: false,
		JSONFileURL:        strings.TrimSpace(options.JSONFileURL),
		MetadataObject:     options.Metadata,
	}

	if request.HolderID == "" {
		return StartInscriptionRequest{}, fmt.Errorf("holder ID is required")
	}

	switch input.Type {
	case InscriptionInputTypeURL:
		if strings.TrimSpace(input.URL) == "" {
			return StartInscriptionRequest{}, fmt.Errorf("input.url is required for url input type")
		}
		request.File = FileInput{
			Type: "url",
			URL:  strings.TrimSpace(input.URL),
		}
	case InscriptionInputTypeFile:
		base64Value, fileName, mimeType, err := convertFilePathToBase64(input.Path)
		if err != nil {
			return StartInscriptionRequest{}, err
		}
		request.File = FileInput{
			Type:     "base64",
			Base64:   base64Value,
			FileName: fileName,
			MimeType: mimeType,
		}
	case InscriptionInputTypeBuffer:
		if len(input.Buffer) == 0 {
			return StartInscriptionRequest{}, fmt.Errorf("input.buffer is required for buffer input type")
		}
		fileName := strings.TrimSpace(input.FileName)
		if fileName == "" {
			return StartInscriptionRequest{}, fmt.Errorf("input.fileName is required for buffer input type")
		}

		mimeType := strings.TrimSpace(input.MimeType)
		if mimeType == "" {
			mimeType = guessMimeTypeFromName(fileName)
		}

		request.File = FileInput{
			Type:     "base64",
			Base64:   base64.StdEncoding.EncodeToString(input.Buffer),
			FileName: fileName,
			MimeType: mimeType,
		}
	default:
		return StartInscriptionRequest{}, fmt.Errorf("input.type must be one of: url, file, buffer")
	}

	if mode == ModeHashinal {
		if request.MetadataObject == nil {
			return StartInscriptionRequest{}, fmt.Errorf("hashinal mode requires metadataObject")
		}
		required := []string{"name", "creator", "description", "type"}
		for _, key := range required {
			value := strings.TrimSpace(stringOrDefault(request.MetadataObject, key, ""))
			if value == "" {
				return StartInscriptionRequest{}, fmt.Errorf("hashinal mode requires metadataObject.%s", key)
			}
		}
		if strings.TrimSpace(request.Creator) == "" {
			request.Creator = strings.TrimSpace(stringOrDefault(request.MetadataObject, "creator", ""))
		}
		if strings.TrimSpace(request.Description) == "" {
			request.Description = strings.TrimSpace(stringOrDefault(request.MetadataObject, "description", ""))
		}
	}

	return request, nil
}

func convertFilePathToBase64(path string) (string, string, string, error) {
	trimmedPath := strings.TrimSpace(path)
	if trimmedPath == "" {
		return "", "", "", fmt.Errorf("input.path is required for file input type")
	}

	bytes, err := os.ReadFile(trimmedPath)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to read file %s: %w", trimmedPath, err)
	}

	fileName := filepath.Base(trimmedPath)
	mimeType := guessMimeTypeFromName(fileName)

	return base64.StdEncoding.EncodeToString(bytes), fileName, mimeType, nil
}

func encodeBufferToBase64(buffer []byte) string {
	return base64.StdEncoding.EncodeToString(buffer)
}

func stringOrDefault(input map[string]any, key string, fallback string) string {
	if input == nil {
		return fallback
	}
	raw, ok := input[key]
	if !ok {
		return fallback
	}
	value, ok := raw.(string)
	if !ok {
		return fallback
	}
	return value
}

func guessMimeTypeFromName(fileName string) string {
	extension := strings.ToLower(strings.TrimSpace(filepath.Ext(fileName)))
	switch extension {
	case ".txt":
		return "text/plain"
	case ".json":
		return "application/json"
	case ".html", ".htm":
		return "text/html"
	case ".css":
		return "text/css"
	case ".js", ".mjs":
		return "application/javascript"
	case ".ts":
		return "application/typescript"
	case ".tsx":
		return "text/tsx"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".svg":
		return "image/svg+xml"
	case ".webp":
		return "image/webp"
	case ".avif":
		return "image/avif"
	case ".mp4":
		return "video/mp4"
	case ".webm":
		return "video/webm"
	case ".mp3":
		return "audio/mpeg"
	case ".wav":
		return "audio/wav"
	case ".ogg":
		return "audio/ogg"
	case ".pdf":
		return "application/pdf"
	case ".zip":
		return "application/zip"
	case ".wasm":
		return "application/wasm"
	default:
		return "application/octet-stream"
	}
}
