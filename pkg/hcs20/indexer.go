package hcs20

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/hashgraph-online/standards-sdk-go/pkg/mirror"
	"github.com/hashgraph-online/standards-sdk-go/pkg/shared"
)

type registeredTopic struct {
	TopicID   string
	IsPrivate bool
}

type PointsIndexer struct {
	mirrorClient *mirror.Client

	mutex               sync.RWMutex
	state               PointsState
	lastIndexedSequence map[string]int64

	pollStopChannel chan struct{}
	pollDoneChannel chan struct{}
}

// NewPointsIndexer creates a mirror-backed HCS-20 points indexer.
func NewPointsIndexer(config IndexerConfig) (*PointsIndexer, error) {
	network, err := shared.NormalizeNetwork(config.Network)
	if err != nil {
		return nil, err
	}

	mirrorClient, err := mirror.NewClient(mirror.Config{
		Network: network,
		BaseURL: config.MirrorBaseURL,
		APIKey:  config.MirrorAPIKey,
	})
	if err != nil {
		return nil, err
	}

	return &PointsIndexer{
		mirrorClient:        mirrorClient,
		state:               newEmptyState(),
		lastIndexedSequence: map[string]int64{},
		pollStopChannel:     nil,
		pollDoneChannel:     nil,
	}, nil
}

// StateSnapshot returns a deep copy of the current index state.
func (indexer *PointsIndexer) StateSnapshot() PointsState {
	indexer.mutex.RLock()
	defer indexer.mutex.RUnlock()

	deployedPoints := make(map[string]PointsInfo, len(indexer.state.DeployedPoints))
	for key, value := range indexer.state.DeployedPoints {
		deployedPoints[key] = value
	}

	balances := make(map[string]map[string]PointsBalance, len(indexer.state.Balances))
	for tick, tickBalances := range indexer.state.Balances {
		clone := make(map[string]PointsBalance, len(tickBalances))
		for accountID, balance := range tickBalances {
			clone[accountID] = balance
		}
		balances[tick] = clone
	}

	transactions := make([]PointsTransaction, len(indexer.state.Transactions))
	copy(transactions, indexer.state.Transactions)

	return PointsState{
		DeployedPoints:         deployedPoints,
		Balances:               balances,
		Transactions:           transactions,
		LastProcessedSequence:  indexer.state.LastProcessedSequence,
		LastProcessedTimestamp: indexer.state.LastProcessedTimestamp,
	}
}

// GetPointsInfo returns deployed points metadata for the given tick.
func (indexer *PointsIndexer) GetPointsInfo(tick string) (PointsInfo, bool) {
	normalizedTick := NormalizeTick(tick)
	indexer.mutex.RLock()
	defer indexer.mutex.RUnlock()
	info, ok := indexer.state.DeployedPoints[normalizedTick]
	return info, ok
}

// GetBalance returns the indexed balance for tick/account.
func (indexer *PointsIndexer) GetBalance(tick string, accountID string) string {
	normalizedTick := NormalizeTick(tick)
	indexer.mutex.RLock()
	defer indexer.mutex.RUnlock()

	tickBalances, exists := indexer.state.Balances[normalizedTick]
	if !exists {
		return "0"
	}

	normalizedAccountID, err := NormalizeAccountID(accountID)
	if err != nil {
		return "0"
	}

	balance, exists := tickBalances[normalizedAccountID]
	if !exists {
		return "0"
	}

	return balance.Balance
}

// IndexOnce runs one indexing cycle.
func (indexer *PointsIndexer) IndexOnce(
	ctx context.Context,
	options IndexOptions,
) error {
	includePublic := options.IncludePublicTopic
	includeRegistry := options.IncludeRegistryTopic
	if !options.IncludePublicTopic && !options.IncludeRegistryTopic && len(options.PrivateTopics) == 0 {
		includePublic = true
		includeRegistry = true
	}

	publicTopicID := options.PublicTopicID
	if publicTopicID == "" {
		publicTopicID = DefaultPublicTopicID
	}
	registryTopicID := options.RegistryTopicID
	if registryTopicID == "" {
		registryTopicID = DefaultRegistryTopicID
	}

	normalizedPublicTopicID, err := NormalizeAccountID(publicTopicID)
	if err != nil {
		return err
	}
	normalizedRegistryTopicID, err := NormalizeAccountID(registryTopicID)
	if err != nil {
		return err
	}

	if includePublic {
		if err := indexer.indexTopic(ctx, normalizedPublicTopicID, false); err != nil {
			return err
		}
	}

	seenTopics := map[string]bool{}
	registeredTopics := make([]registeredTopic, 0)
	if includeRegistry {
		topics, topicErr := indexer.fetchRegisteredTopics(ctx, normalizedRegistryTopicID)
		if topicErr != nil {
			return topicErr
		}
		registeredTopics = append(registeredTopics, topics...)
	}

	for _, topicID := range options.PrivateTopics {
		normalizedTopicID, normalizeErr := NormalizeAccountID(topicID)
		if normalizeErr != nil {
			return normalizeErr
		}
		registeredTopics = append(registeredTopics, registeredTopic{
			TopicID:   normalizedTopicID,
			IsPrivate: true,
		})
	}

	for _, topic := range registeredTopics {
		if topic.TopicID == normalizedPublicTopicID {
			continue
		}
		if seenTopics[topic.TopicID] {
			continue
		}
		seenTopics[topic.TopicID] = true
		if err := indexer.indexTopic(ctx, topic.TopicID, topic.IsPrivate); err != nil {
			return err
		}
	}

	return nil
}

// StartPolling starts periodic indexing until StopPolling is called.
func (indexer *PointsIndexer) StartPolling(
	ctx context.Context,
	options IndexOptions,
	interval time.Duration,
) error {
	if interval <= 0 {
		interval = 30 * time.Second
	}

	indexer.mutex.Lock()
	if indexer.pollStopChannel != nil {
		indexer.mutex.Unlock()
		return fmt.Errorf("indexer polling already running")
	}
	stopChannel := make(chan struct{})
	doneChannel := make(chan struct{})
	indexer.pollStopChannel = stopChannel
	indexer.pollDoneChannel = doneChannel
	indexer.mutex.Unlock()

	go func() {
		defer close(doneChannel)

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		_ = indexer.IndexOnce(ctx, options)

		for {
			select {
			case <-ctx.Done():
				return
			case <-stopChannel:
				return
			case <-ticker.C:
				_ = indexer.IndexOnce(ctx, options)
			}
		}
	}()

	return nil
}

// StopPolling stops an active polling loop.
func (indexer *PointsIndexer) StopPolling() {
	indexer.mutex.Lock()
	stopChannel := indexer.pollStopChannel
	doneChannel := indexer.pollDoneChannel
	indexer.pollStopChannel = nil
	indexer.pollDoneChannel = nil
	indexer.mutex.Unlock()

	if stopChannel != nil {
		close(stopChannel)
	}
	if doneChannel != nil {
		<-doneChannel
	}
}

func (indexer *PointsIndexer) fetchRegisteredTopics(
	ctx context.Context,
	registryTopicID string,
) ([]registeredTopic, error) {
	messages, err := indexer.mirrorClient.GetTopicMessages(ctx, registryTopicID, mirror.MessageQueryOptions{
		Order: "asc",
		Limit: 1000,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query HCS-20 registry topic %s: %w", registryTopicID, err)
	}

	topics := make([]registeredTopic, 0)
	for _, topicMessage := range messages {
		payload, decodeErr := mirror.DecodeMessageData(topicMessage)
		if decodeErr != nil {
			continue
		}

		message, parseErr := ParseMessageBytes(payload)
		if parseErr != nil {
			continue
		}
		if message.Operation != OperationRegister {
			continue
		}
		if message.TopicID == "" || message.Private == nil {
			continue
		}

		topics = append(topics, registeredTopic{
			TopicID:   message.TopicID,
			IsPrivate: *message.Private,
		})
	}

	return topics, nil
}

func (indexer *PointsIndexer) indexTopic(
	ctx context.Context,
	topicID string,
	isPrivate bool,
) error {
	indexer.mutex.RLock()
	lastSequence := indexer.lastIndexedSequence[topicID]
	indexer.mutex.RUnlock()

	sequenceFilter := ""
	if lastSequence > 0 {
		sequenceFilter = fmt.Sprintf("gt:%d", lastSequence)
	}

	messages, err := indexer.mirrorClient.GetTopicMessages(ctx, topicID, mirror.MessageQueryOptions{
		SequenceNumber: sequenceFilter,
		Order:          "asc",
		Limit:          1000,
	})
	if err != nil {
		return fmt.Errorf("failed to fetch topic %s messages: %w", topicID, err)
	}

	maxSequence := lastSequence
	for _, topicMessage := range messages {
		payload, decodeErr := mirror.DecodeMessageData(topicMessage)
		if decodeErr != nil {
			continue
		}

		message, parseErr := ParseMessageBytes(payload)
		if parseErr != nil {
			continue
		}

		if topicMessage.SequenceNumber > maxSequence {
			maxSequence = topicMessage.SequenceNumber
		}

		indexer.processMessage(topicID, topicMessage, message, isPrivate)
	}

	if maxSequence > lastSequence {
		indexer.mutex.Lock()
		indexer.lastIndexedSequence[topicID] = maxSequence
		indexer.mutex.Unlock()
	}

	return nil
}

func (indexer *PointsIndexer) processMessage(
	topicID string,
	topicMessage mirror.TopicMessage,
	message Message,
	isPrivate bool,
) {
	switch message.Operation {
	case OperationDeploy:
		indexer.processDeploy(topicID, topicMessage, message, isPrivate)
	case OperationMint:
		indexer.processMint(topicID, topicMessage, message)
	case OperationTransfer:
		indexer.processTransfer(topicID, topicMessage, message, isPrivate)
	case OperationBurn:
		indexer.processBurn(topicID, topicMessage, message, isPrivate)
	}
}

func (indexer *PointsIndexer) processDeploy(
	topicID string,
	topicMessage mirror.TopicMessage,
	message Message,
	isPrivate bool,
) {
	normalizedTick := NormalizeTick(message.Tick)

	indexer.mutex.Lock()
	defer indexer.mutex.Unlock()

	if _, exists := indexer.state.DeployedPoints[normalizedTick]; exists {
		return
	}

	indexer.state.DeployedPoints[normalizedTick] = PointsInfo{
		Name:                message.Name,
		Tick:                normalizedTick,
		MaxSupply:           message.Max,
		LimitPerMint:        message.Limit,
		Metadata:            message.Metadata,
		TopicID:             topicID,
		DeployerAccountID:   topicMessage.PayerAccountID,
		CurrentSupply:       "0",
		DeploymentTimestamp: topicMessage.ConsensusTimestamp,
		IsPrivate:           isPrivate,
	}

	indexer.state.LastProcessedSequence++
	indexer.state.LastProcessedTimestamp = topicMessage.ConsensusTimestamp
}

func (indexer *PointsIndexer) processMint(
	topicID string,
	topicMessage mirror.TopicMessage,
	message Message,
) {
	normalizedTick := NormalizeTick(message.Tick)

	indexer.mutex.Lock()
	defer indexer.mutex.Unlock()

	pointsInfo, exists := indexer.state.DeployedPoints[normalizedTick]
	if !exists {
		return
	}

	mintAmount, ok := parseAmount(message.Amount)
	if !ok {
		return
	}
	currentSupply, ok := parseAmount(pointsInfo.CurrentSupply)
	if !ok {
		return
	}
	maxSupply, ok := parseAmount(pointsInfo.MaxSupply)
	if !ok {
		return
	}
	if new(big.Int).Add(currentSupply, mintAmount).Cmp(maxSupply) > 0 {
		return
	}

	if pointsInfo.LimitPerMint != "" {
		limit, limitOK := parseAmount(pointsInfo.LimitPerMint)
		if !limitOK {
			return
		}
		if mintAmount.Cmp(limit) > 0 {
			return
		}
	}

	updatedSupply := new(big.Int).Add(currentSupply, mintAmount)
	pointsInfo.CurrentSupply = updatedSupply.String()
	indexer.state.DeployedPoints[normalizedTick] = pointsInfo

	tickBalances := indexer.getOrCreateTickBalancesLocked(normalizedTick)
	recipientBalance, recipientExists := tickBalances[message.To]
	if recipientExists {
		existingBalance, balanceOK := parseAmount(recipientBalance.Balance)
		if !balanceOK {
			return
		}
		recipientBalance.Balance = new(big.Int).Add(existingBalance, mintAmount).String()
		recipientBalance.LastUpdated = topicMessage.ConsensusTimestamp
		tickBalances[message.To] = recipientBalance
	} else {
		tickBalances[message.To] = PointsBalance{
			Tick:        normalizedTick,
			AccountID:   message.To,
			Balance:     mintAmount.String(),
			LastUpdated: topicMessage.ConsensusTimestamp,
		}
	}

	indexer.state.Transactions = append(indexer.state.Transactions, PointsTransaction{
		ID:             transactionIDOrFallback(topicID, topicMessage),
		Operation:      OperationMint,
		Tick:           normalizedTick,
		Amount:         message.Amount,
		To:             message.To,
		Timestamp:      topicMessage.ConsensusTimestamp,
		SequenceNumber: topicMessage.SequenceNumber,
		TopicID:        topicID,
		TransactionID:  topicMessage.ConsensusTimestamp,
		Memo:           message.Memo,
	})

	indexer.state.LastProcessedSequence++
	indexer.state.LastProcessedTimestamp = topicMessage.ConsensusTimestamp
}

func (indexer *PointsIndexer) processTransfer(
	topicID string,
	topicMessage mirror.TopicMessage,
	message Message,
	isPrivate bool,
) {
	normalizedTick := NormalizeTick(message.Tick)

	indexer.mutex.Lock()
	defer indexer.mutex.Unlock()

	tickBalances, exists := indexer.state.Balances[normalizedTick]
	if !exists {
		return
	}

	if !isPrivate && topicMessage.PayerAccountID != message.From {
		return
	}

	transferAmount, ok := parseAmount(message.Amount)
	if !ok {
		return
	}

	senderBalance, senderExists := tickBalances[message.From]
	if !senderExists {
		return
	}

	senderAmount, senderOK := parseAmount(senderBalance.Balance)
	if !senderOK {
		return
	}
	if senderAmount.Cmp(transferAmount) < 0 {
		return
	}

	updatedSender := new(big.Int).Sub(senderAmount, transferAmount)
	senderBalance.Balance = updatedSender.String()
	senderBalance.LastUpdated = topicMessage.ConsensusTimestamp
	tickBalances[message.From] = senderBalance

	receiverBalance, receiverExists := tickBalances[message.To]
	if receiverExists {
		receiverAmount, receiverOK := parseAmount(receiverBalance.Balance)
		if !receiverOK {
			return
		}
		receiverBalance.Balance = new(big.Int).Add(receiverAmount, transferAmount).String()
		receiverBalance.LastUpdated = topicMessage.ConsensusTimestamp
		tickBalances[message.To] = receiverBalance
	} else {
		tickBalances[message.To] = PointsBalance{
			Tick:        normalizedTick,
			AccountID:   message.To,
			Balance:     transferAmount.String(),
			LastUpdated: topicMessage.ConsensusTimestamp,
		}
	}

	indexer.state.Transactions = append(indexer.state.Transactions, PointsTransaction{
		ID:             transactionIDOrFallback(topicID, topicMessage),
		Operation:      OperationTransfer,
		Tick:           normalizedTick,
		Amount:         message.Amount,
		From:           message.From,
		To:             message.To,
		Timestamp:      topicMessage.ConsensusTimestamp,
		SequenceNumber: topicMessage.SequenceNumber,
		TopicID:        topicID,
		TransactionID:  topicMessage.ConsensusTimestamp,
		Memo:           message.Memo,
	})

	indexer.state.LastProcessedSequence++
	indexer.state.LastProcessedTimestamp = topicMessage.ConsensusTimestamp
}

func (indexer *PointsIndexer) processBurn(
	topicID string,
	topicMessage mirror.TopicMessage,
	message Message,
	isPrivate bool,
) {
	normalizedTick := NormalizeTick(message.Tick)

	indexer.mutex.Lock()
	defer indexer.mutex.Unlock()

	pointsInfo, pointsExists := indexer.state.DeployedPoints[normalizedTick]
	tickBalances, balancesExists := indexer.state.Balances[normalizedTick]
	if !pointsExists || !balancesExists {
		return
	}

	if !isPrivate && topicMessage.PayerAccountID != message.From {
		return
	}

	burnAmount, burnOK := parseAmount(message.Amount)
	if !burnOK {
		return
	}

	accountBalance, accountExists := tickBalances[message.From]
	if !accountExists {
		return
	}

	accountAmount, accountOK := parseAmount(accountBalance.Balance)
	if !accountOK {
		return
	}
	if accountAmount.Cmp(burnAmount) < 0 {
		return
	}

	currentSupply, supplyOK := parseAmount(pointsInfo.CurrentSupply)
	if !supplyOK {
		return
	}

	updatedBalance := new(big.Int).Sub(accountAmount, burnAmount)
	accountBalance.Balance = updatedBalance.String()
	accountBalance.LastUpdated = topicMessage.ConsensusTimestamp
	tickBalances[message.From] = accountBalance

	updatedSupply := new(big.Int).Sub(currentSupply, burnAmount)
	pointsInfo.CurrentSupply = updatedSupply.String()
	indexer.state.DeployedPoints[normalizedTick] = pointsInfo

	indexer.state.Transactions = append(indexer.state.Transactions, PointsTransaction{
		ID:             transactionIDOrFallback(topicID, topicMessage),
		Operation:      OperationBurn,
		Tick:           normalizedTick,
		Amount:         message.Amount,
		From:           message.From,
		Timestamp:      topicMessage.ConsensusTimestamp,
		SequenceNumber: topicMessage.SequenceNumber,
		TopicID:        topicID,
		TransactionID:  topicMessage.ConsensusTimestamp,
		Memo:           message.Memo,
	})

	indexer.state.LastProcessedSequence++
	indexer.state.LastProcessedTimestamp = topicMessage.ConsensusTimestamp
}

func (indexer *PointsIndexer) getOrCreateTickBalancesLocked(
	tick string,
) map[string]PointsBalance {
	tickBalances, exists := indexer.state.Balances[tick]
	if !exists {
		tickBalances = map[string]PointsBalance{}
		indexer.state.Balances[tick] = tickBalances
	}
	return tickBalances
}

func parseAmount(amount string) (*big.Int, bool) {
	trimmed := strings.TrimSpace(amount)
	parsed := new(big.Int)
	value, ok := parsed.SetString(trimmed, 10)
	if !ok {
		return nil, false
	}
	if value.Sign() < 0 {
		return nil, false
	}
	return value, true
}

func transactionIDOrFallback(topicID string, topicMessage mirror.TopicMessage) string {
	if topicMessage.ConsensusTimestamp != "" {
		return topicMessage.ConsensusTimestamp
	}
	return fmt.Sprintf("%s-%d", topicID, topicMessage.SequenceNumber)
}

func newEmptyState() PointsState {
	return PointsState{
		DeployedPoints:         map[string]PointsInfo{},
		Balances:               map[string]map[string]PointsBalance{},
		Transactions:           []PointsTransaction{},
		LastProcessedSequence:  0,
		LastProcessedTimestamp: "",
	}
}
