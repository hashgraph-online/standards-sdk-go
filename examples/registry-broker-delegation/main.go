package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/hashgraph-online/standards-sdk-go/pkg/registrybroker"
)

const (
	defaultBaseURL = "https://hol.org/registry/api/v1"
	defaultTask    = "Review an SDK PR and split out docs and verification subtasks."
)

func main() {
	if err := run(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	client, err := registrybroker.NewRegistryBrokerClient(registrybroker.RegistryBrokerClientOptions{
		BaseURL: strings.TrimSpace(getEnvOrDefault("REGISTRY_BROKER_BASE_URL", defaultBaseURL)),
		APIKey:  strings.TrimSpace(os.Getenv("REGISTRY_BROKER_API_KEY")),
	})
	if err != nil {
		return err
	}

	limit, err := strconv.Atoi(strings.TrimSpace(getEnvOrDefault("REGISTRY_BROKER_DELEGATION_LIMIT", "3")))
	if err != nil || limit <= 0 {
		limit = 3
	}

	response, err := client.Delegate(context.Background(), registrybroker.DelegationPlanRequest{
		Task:    strings.TrimSpace(getEnvOrDefault("REGISTRY_BROKER_DELEGATION_TASK", defaultTask)),
		Context: strings.TrimSpace(os.Getenv("REGISTRY_BROKER_DELEGATION_CONTEXT")),
		Limit:   limit,
	})
	if err != nil {
		return err
	}

	fmt.Printf("task=%s\n", response.Task)
	fmt.Printf("shouldDelegate=%t\n", response.ShouldDelegate)
	if response.LocalFirstReason != "" {
		fmt.Printf("localFirstReason=%s\n", response.LocalFirstReason)
	}
	for index := range response.Opportunities {
		opportunity := &response.Opportunities[index]
		fmt.Printf("\nopportunity=%s title=%s\n", opportunity.ID, opportunity.Title)
		fmt.Printf("reason=%s\n", opportunity.Reason)
		if len(opportunity.Candidates) == 0 {
			fmt.Println("topCandidate=<none>")
			continue
		}
		candidate := opportunity.Candidates[0]
		label := candidate.Label
		if label == "" {
			label = "<unknown>"
		}
		fmt.Printf("topCandidate=%s label=%s\n", candidate.UAID, label)
	}
	return nil
}

func getEnvOrDefault(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}
