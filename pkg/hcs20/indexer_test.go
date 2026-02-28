package hcs20

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

type topicMessagesFixture map[string][]map[string]any

func TestPointsIndexerPrivateTopicFlow(t *testing.T) {
	privateTopicID := "0.0.5001"
	server := newMirrorFixtureServer(t, topicMessagesFixture{
		privateTopicID: {
			newFixtureMessage(t, privateTopicID, 1, "0.0.1001", Message{
				Protocol:  "hcs-20",
				Operation: "deploy",
				Name:      "Loyalty",
				Tick:      "loyal",
				Max:       "1000",
			}),
			newFixtureMessage(t, privateTopicID, 2, "0.0.1001", Message{
				Protocol:  "hcs-20",
				Operation: "mint",
				Tick:      "loyal",
				Amount:    "100",
				To:        "0.0.1001",
			}),
			newFixtureMessage(t, privateTopicID, 3, "0.0.9999", Message{
				Protocol:  "hcs-20",
				Operation: "transfer",
				Tick:      "loyal",
				Amount:    "10",
				From:      "0.0.1001",
				To:        "0.0.1002",
			}),
			newFixtureMessage(t, privateTopicID, 4, "0.0.7777", Message{
				Protocol:  "hcs-20",
				Operation: "burn",
				Tick:      "loyal",
				Amount:    "30",
				From:      "0.0.1001",
			}),
		},
	})
	defer server.Close()

	indexer, err := NewPointsIndexer(IndexerConfig{
		Network:       "testnet",
		MirrorBaseURL: server.URL,
	})
	if err != nil {
		t.Fatalf("unexpected indexer error: %v", err)
	}

	if err := indexer.IndexOnce(t.Context(), IndexOptions{
		IncludePublicTopic:   false,
		IncludeRegistryTopic: false,
		PrivateTopics:        []string{privateTopicID},
	}); err != nil {
		t.Fatalf("unexpected indexer run error: %v", err)
	}

	pointsInfo, exists := indexer.GetPointsInfo("LOYAL")
	if !exists {
		t.Fatal("expected deployed points info")
	}
	if pointsInfo.CurrentSupply != "70" {
		t.Fatalf("expected supply 70, got %s", pointsInfo.CurrentSupply)
	}

	if gotBalance := indexer.GetBalance("loyal", "0.0.1001"); gotBalance != "60" {
		t.Fatalf("expected sender balance 60, got %s", gotBalance)
	}
	if gotBalance := indexer.GetBalance("loyal", "0.0.1002"); gotBalance != "10" {
		t.Fatalf("expected recipient balance 10, got %s", gotBalance)
	}
}

func TestPointsIndexerPublicTopicRejectsPayerMismatch(t *testing.T) {
	publicTopicID := DefaultPublicTopicID
	server := newMirrorFixtureServer(t, topicMessagesFixture{
		publicTopicID: {
			newFixtureMessage(t, publicTopicID, 1, "0.0.1001", Message{
				Protocol:  "hcs-20",
				Operation: "deploy",
				Name:      "Public Loyalty",
				Tick:      "pub",
				Max:       "1000",
			}),
			newFixtureMessage(t, publicTopicID, 2, "0.0.1001", Message{
				Protocol:  "hcs-20",
				Operation: "mint",
				Tick:      "pub",
				Amount:    "50",
				To:        "0.0.1001",
			}),
			newFixtureMessage(t, publicTopicID, 3, "0.0.9999", Message{
				Protocol:  "hcs-20",
				Operation: "transfer",
				Tick:      "pub",
				Amount:    "20",
				From:      "0.0.1001",
				To:        "0.0.1002",
			}),
		},
	})
	defer server.Close()

	indexer, err := NewPointsIndexer(IndexerConfig{
		Network:       "testnet",
		MirrorBaseURL: server.URL,
	})
	if err != nil {
		t.Fatalf("unexpected indexer error: %v", err)
	}

	if err := indexer.IndexOnce(t.Context(), IndexOptions{
		IncludePublicTopic:   true,
		IncludeRegistryTopic: false,
		PublicTopicID:        publicTopicID,
	}); err != nil {
		t.Fatalf("unexpected indexer run error: %v", err)
	}

	if gotBalance := indexer.GetBalance("pub", "0.0.1001"); gotBalance != "50" {
		t.Fatalf("expected sender balance to remain 50, got %s", gotBalance)
	}
	if gotBalance := indexer.GetBalance("pub", "0.0.1002"); gotBalance != "0" {
		t.Fatalf("expected recipient balance to remain 0, got %s", gotBalance)
	}
}

func TestPointsIndexerRegistryDiscovery(t *testing.T) {
	registryTopicID := "0.0.7001"
	privateTopicID := "0.0.7002"
	isPrivate := true

	server := newMirrorFixtureServer(t, topicMessagesFixture{
		registryTopicID: {
			newFixtureMessage(t, registryTopicID, 1, "0.0.1001", Message{
				Protocol:  "hcs-20",
				Operation: "register",
				Name:      "Private Loyalty",
				TopicID:   privateTopicID,
				Private:   &isPrivate,
			}),
		},
		privateTopicID: {
			newFixtureMessage(t, privateTopicID, 1, "0.0.1001", Message{
				Protocol:  "hcs-20",
				Operation: "deploy",
				Name:      "Private Loyalty",
				Tick:      "priv",
				Max:       "1000",
			}),
			newFixtureMessage(t, privateTopicID, 2, "0.0.1001", Message{
				Protocol:  "hcs-20",
				Operation: "mint",
				Tick:      "priv",
				Amount:    "80",
				To:        "0.0.1001",
			}),
		},
	})
	defer server.Close()

	indexer, err := NewPointsIndexer(IndexerConfig{
		Network:       "testnet",
		MirrorBaseURL: server.URL,
	})
	if err != nil {
		t.Fatalf("unexpected indexer error: %v", err)
	}

	if err := indexer.IndexOnce(t.Context(), IndexOptions{
		IncludePublicTopic:   false,
		IncludeRegistryTopic: true,
		RegistryTopicID:      registryTopicID,
	}); err != nil {
		t.Fatalf("unexpected indexer run error: %v", err)
	}

	if gotBalance := indexer.GetBalance("priv", "0.0.1001"); gotBalance != "80" {
		t.Fatalf("expected discovered private topic balance 80, got %s", gotBalance)
	}
}

func newFixtureMessage(
	t *testing.T,
	topicID string,
	sequenceNumber int64,
	payerAccountID string,
	message Message,
) map[string]any {
	t.Helper()
	payload, _, err := BuildMessagePayload(message)
	if err != nil {
		t.Fatalf("failed to build fixture message: %v", err)
	}
	return map[string]any{
		"consensus_timestamp":  "1234567890." + strconv.FormatInt(sequenceNumber, 10),
		"message":              base64.StdEncoding.EncodeToString(payload),
		"payer_account_id":     payerAccountID,
		"running_hash":         "",
		"running_hash_version": 3,
		"sequence_number":      sequenceNumber,
		"topic_id":             topicID,
	}
}

func newMirrorFixtureServer(t *testing.T, fixtures topicMessagesFixture) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		responseWriter.Header().Set("Content-Type", "application/json")

		pathParts := strings.Split(strings.Trim(request.URL.Path, "/"), "/")
		if len(pathParts) != 5 || pathParts[0] != "api" || pathParts[1] != "v1" || pathParts[2] != "topics" || pathParts[4] != "messages" {
			responseWriter.WriteHeader(http.StatusNotFound)
			_, _ = responseWriter.Write([]byte(`{"error":"not found"}`))
			return
		}

		topicID := pathParts[3]
		messages, exists := fixtures[topicID]
		if !exists {
			messages = []map[string]any{}
		}

		sequenceFilter := request.URL.Query().Get("sequencenumber")
		filtered := applySequenceFilter(messages, sequenceFilter)
		if order := request.URL.Query().Get("order"); strings.EqualFold(order, "desc") {
			reversed := make([]map[string]any, 0, len(filtered))
			for index := len(filtered) - 1; index >= 0; index-- {
				reversed = append(reversed, filtered[index])
			}
			filtered = reversed
		}

		body := map[string]any{
			"links": map[string]any{
				"next": "",
			},
			"messages": filtered,
		}
		encoded, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("failed to marshal mirror fixture response: %v", err)
		}
		_, _ = responseWriter.Write(encoded)
	}))
}

func applySequenceFilter(messages []map[string]any, sequenceFilter string) []map[string]any {
	trimmedFilter := strings.TrimSpace(sequenceFilter)
	if trimmedFilter == "" {
		return append([]map[string]any{}, messages...)
	}

	filterParts := strings.SplitN(trimmedFilter, ":", 2)
	if len(filterParts) != 2 {
		return append([]map[string]any{}, messages...)
	}

	value, err := strconv.ParseInt(filterParts[1], 10, 64)
	if err != nil {
		return append([]map[string]any{}, messages...)
	}

	filtered := make([]map[string]any, 0, len(messages))
	for _, message := range messages {
		sequenceAny, exists := message["sequence_number"]
		if !exists {
			continue
		}

		sequenceNumber, ok := sequenceAny.(int64)
		if !ok {
			if number, castOK := sequenceAny.(float64); castOK {
				sequenceNumber = int64(number)
			} else {
				continue
			}
		}

		switch filterParts[0] {
		case "gt":
			if sequenceNumber > value {
				filtered = append(filtered, message)
			}
		case "eq":
			if sequenceNumber == value {
				filtered = append(filtered, message)
			}
		default:
			filtered = append(filtered, message)
		}
	}

	return filtered
}
