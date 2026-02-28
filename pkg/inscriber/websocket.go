package inscriber

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	socketio "github.com/zhouhui8915/go-socket.io-client"
)

type webSocketServer struct {
	URL    string `json:"url"`
	Status string `json:"status"`
}

type webSocketServersResponse struct {
	Servers     []webSocketServer `json:"servers"`
	Recommended string            `json:"recommended"`
}

func (c *Client) resolveWebSocketBaseURL(ctx context.Context) (string, error) {
	if strings.TrimSpace(c.webSocketBaseURL) != "" {
		return strings.TrimSpace(c.webSocketBaseURL), nil
	}

	var response webSocketServersResponse
	if err := c.getJSON(ctx, "/inscriptions/websocket-servers", &response); err != nil {
		return "", err
	}

	if strings.TrimSpace(response.Recommended) != "" {
		return strings.TrimSpace(response.Recommended), nil
	}

	for _, server := range response.Servers {
		if strings.EqualFold(strings.TrimSpace(server.Status), "active") &&
			strings.TrimSpace(server.URL) != "" {
			return strings.TrimSpace(server.URL), nil
		}
	}

	for _, server := range response.Servers {
		if strings.TrimSpace(server.URL) != "" {
			return strings.TrimSpace(server.URL), nil
		}
	}

	return "", fmt.Errorf("no websocket servers available")
}

func (c *Client) waitForInscriptionWebSocket(
	ctx context.Context,
	transactionID string,
	progressCallback RegistrationProgressCallback,
) (InscriptionJob, error) {
	wsURL, err := c.resolveWebSocketBaseURL(ctx)
	if err != nil {
		return InscriptionJob{}, err
	}

	options := &socketio.Options{
		Transport: "websocket",
		Query: map[string]string{
			"apiKey": c.apiKey,
		},
		Header: map[string][]string{
			"x-api-key": {c.apiKey},
		},
	}

	client, err := socketio.NewClient(wsURL, options)
	if err != nil {
		return InscriptionJob{}, err
	}

	normalizedTransactionID := normalizeTransactionID(transactionID)
	progressChannel := make(chan map[string]any, 4)
	completeChannel := make(chan map[string]any, 2)
	errorChannel := make(chan string, 2)

	_ = client.On("error", func(message any) {
		errorChannel <- fmt.Sprintf("%v", message)
	})
	_ = client.On("inscription-error", func(payload map[string]any) {
		errorMessage := fmt.Sprintf("%v", payload["error"])
		if strings.TrimSpace(errorMessage) == "" {
			errorMessage = "websocket inscription error"
		}
		errorChannel <- errorMessage
	})
	_ = client.On("inscription-progress", func(payload map[string]any) {
		progressChannel <- payload
	})
	_ = client.On("inscription-complete", func(payload map[string]any) {
		completeChannel <- payload
	})

	inactivityTimeout := c.webSocketInactivityTimeoutMs
	if inactivityTimeout <= 0 {
		inactivityTimeout = 30000
	}
	timer := time.NewTimer(time.Duration(inactivityTimeout) * time.Millisecond)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return InscriptionJob{}, ctx.Err()
		case <-timer.C:
			return InscriptionJob{}, fmt.Errorf("websocket inscription timeout")
		case message := <-errorChannel:
			return InscriptionJob{}, fmt.Errorf("%s", message)
		case payload := <-progressChannel:
			timer.Reset(time.Duration(inactivityTimeout) * time.Millisecond)

			if !matchesInscriptionEvent(normalizedTransactionID, payload) {
				continue
			}

			if progressCallback != nil {
				progressCallback(RegistrationProgressData{
					Stage:           RegistrationStageConfirming,
					Message:         "Processing inscription",
					ProgressPercent: parseFloat(payload["progress"]),
					Details:         payload,
				})
			}

			status := strings.ToLower(parseString(payload["status"]))
			progress := parseFloat(payload["progress"])
			if status == "completed" || progress >= 100 {
				job := parseInscriptionEvent(payload)
				if strings.TrimSpace(job.Status) == "" {
					job.Status = "completed"
				}
				job.Completed = true
				return job, nil
			}
		case payload := <-completeChannel:
			timer.Reset(time.Duration(inactivityTimeout) * time.Millisecond)
			if !matchesInscriptionEvent(normalizedTransactionID, payload) {
				continue
			}

			job := parseInscriptionEvent(payload)
			job.Completed = true
			if strings.TrimSpace(job.Status) == "" {
				job.Status = "completed"
			}
			return job, nil
		}
	}
}

func matchesInscriptionEvent(normalizedTransactionID string, payload map[string]any) bool {
	if normalizedTransactionID == "" {
		return true
	}

	for _, key := range []string{"jobId", "tx_id", "transactionId"} {
		value := normalizeTransactionID(parseString(payload[key]))
		if strings.TrimSpace(value) == "" {
			continue
		}
		if value == normalizedTransactionID {
			return true
		}
	}
	return false
}

func parseInscriptionEvent(payload map[string]any) InscriptionJob {
	return InscriptionJob{
		ID:            parseString(payload["id"]),
		Status:        parseString(payload["status"]),
		Completed:     strings.EqualFold(parseString(payload["status"]), "completed"),
		TxID:          parseString(payload["tx_id"]),
		TransactionID: parseString(payload["transactionId"]),
		TopicID:       firstNonEmptyString(payload, "topicId", "topic_id"),
		Error:         parseString(payload["error"]),
	}
}

func parseString(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case fmt.Stringer:
		return typed.String()
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64)
	case int64:
		return strconv.FormatInt(typed, 10)
	case int:
		return strconv.Itoa(typed)
	default:
		return ""
	}
}

func parseFloat(value any) float64 {
	switch typed := value.(type) {
	case float64:
		return typed
	case float32:
		return float64(typed)
	case int:
		return float64(typed)
	case int64:
		return float64(typed)
	case string:
		parsed, err := strconv.ParseFloat(strings.TrimSpace(typed), 64)
		if err != nil {
			return 0
		}
		return parsed
	default:
		return 0
	}
}

func firstNonEmptyString(payload map[string]any, keys ...string) string {
	for _, key := range keys {
		value := strings.TrimSpace(parseString(payload[key]))
		if value != "" {
			return value
		}
	}
	return ""
}

func normalizeWebSocketURL(input string) string {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return ""
	}
	parsed, err := url.Parse(trimmed)
	if err != nil {
		return trimmed
	}
	return parsed.String()
}
