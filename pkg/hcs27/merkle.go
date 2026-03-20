package hcs27

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

var (
	leafDomainPrefix = [1]byte{0x00}
	nodeDomainPrefix = [1]byte{0x01}
)

// EmptyRoot performs the requested operation.
func EmptyRoot() []byte {
	sum := sha256.Sum256([]byte{})
	return sum[:]
}

// HashLeaf performs the requested operation.
func HashLeaf(canonicalEntry []byte) []byte {
	hasher := sha256.New()
	_, _ = hasher.Write(leafDomainPrefix[:])
	_, _ = hasher.Write(canonicalEntry)
	return hasher.Sum(nil)
}

// HashNode performs the requested operation.
func HashNode(left, right []byte) []byte {
	hasher := sha256.New()
	_, _ = hasher.Write(nodeDomainPrefix[:])
	_, _ = hasher.Write(left)
	_, _ = hasher.Write(right)
	return hasher.Sum(nil)
}

// MerkleRootFromCanonicalEntries performs the requested operation.
func MerkleRootFromCanonicalEntries(entries [][]byte) []byte {
	switch len(entries) {
	case 0:
		return EmptyRoot()
	case 1:
		return HashLeaf(entries[0])
	default:
		split := largestPowerOfTwoLessThan(uint64(len(entries)))
		left := MerkleRootFromCanonicalEntries(entries[:split])
		right := MerkleRootFromCanonicalEntries(entries[split:])
		return HashNode(left, right)
	}
}

// MerkleRootFromEntries performs the requested operation.
func MerkleRootFromEntries(entries []any) ([]byte, error) {
	canonicalEntries := make([][]byte, 0, len(entries))
	for _, entry := range entries {
		canonical, err := CanonicalizeJSON(entry)
		if err != nil {
			return nil, err
		}
		canonicalEntries = append(canonicalEntries, canonical)
	}

	return MerkleRootFromCanonicalEntries(canonicalEntries), nil
}

// LeafHashHexFromEntry performs the requested operation.
func LeafHashHexFromEntry(entry any) (string, error) {
	canonical, err := CanonicalizeJSON(entry)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(HashLeaf(canonical)), nil
}

// VerifyInclusionProof performs the requested operation.
func VerifyInclusionProof(
	leafIndex uint64,
	treeSize uint64,
	leafHashHex string,
	path []string,
	expectedRootB64 string,
) (bool, error) {
	if treeSize == 0 {
		return false, fmt.Errorf("treeSize must be greater than zero for inclusion proofs")
	}
	if leafIndex >= treeSize {
		return false, fmt.Errorf("leafIndex must be less than treeSize")
	}

	leafHash, err := hex.DecodeString(strings.TrimSpace(leafHashHex))
	if err != nil {
		return false, fmt.Errorf("leafHash must be valid hex: %w", err)
	}

	fn := leafIndex
	sn := treeSize - 1
	current := make([]byte, len(leafHash))
	copy(current, leafHash)

	for _, node := range path {
		if sn == 0 {
			return false, nil
		}

		sibling, err := base64.StdEncoding.DecodeString(node)
		if err != nil {
			return false, fmt.Errorf("path element must be valid base64: %w", err)
		}

		if leastSignificantBit(fn) == 1 || fn == sn {
			current = HashNode(sibling, current)
			if leastSignificantBit(fn) == 0 {
				for leastSignificantBit(fn) == 0 && fn != 0 {
					fn /= 2
					sn /= 2
				}
			}
		} else {
			current = HashNode(current, sibling)
		}

		fn /= 2
		sn /= 2
	}

	return sn == 0 && base64.StdEncoding.EncodeToString(current) == expectedRootB64, nil
}

// VerifyInclusionProofObject verifies a proof object that follows the HCS-27 draft shape.
func VerifyInclusionProofObject(proof *InclusionProof) (bool, error) {
	if proof == nil {
		return false, fmt.Errorf("proof is required")
	}
	if proof.TreeVersion != 1 {
		return false, fmt.Errorf("treeVersion must be 1")
	}

	leafIndex, err := parseCanonicalUint64("leafIndex", proof.LeafIndex)
	if err != nil {
		return false, err
	}
	treeSize, err := parseCanonicalUint64("treeSize", proof.TreeSize)
	if err != nil {
		return false, err
	}
	rootHash, err := normalizeProofRootHash("rootHash", proof.RootHash)
	if err != nil {
		return false, err
	}

	return VerifyInclusionProof(
		leafIndex,
		treeSize,
		proof.LeafHash,
		proof.Path,
		rootHash,
	)
}

// VerifyConsistencyProof performs the requested operation.
func VerifyConsistencyProof(
	oldTreeSize uint64,
	newTreeSize uint64,
	oldRootB64 string,
	newRootB64 string,
	consistencyPath []string,
) (bool, error) {
	if oldTreeSize == 0 {
		return true, nil
	}
	if oldTreeSize == newTreeSize {
		return oldRootB64 == newRootB64 && len(consistencyPath) == 0, nil
	}
	if oldTreeSize > newTreeSize {
		return false, nil
	}
	if len(consistencyPath) == 0 {
		return false, nil
	}

	path := make([]string, 0, len(consistencyPath)+1)
	if isExactPowerOfTwo(oldTreeSize) {
		path = append(path, oldRootB64)
	}
	path = append(path, consistencyPath...)

	fn := oldTreeSize - 1
	sn := newTreeSize - 1

	for leastSignificantBit(fn) == 1 {
		fn /= 2
		sn /= 2
	}

	firstHash, err := base64.StdEncoding.DecodeString(path[0])
	if err != nil {
		return false, fmt.Errorf("consistency path element must be base64: %w", err)
	}

	fr := make([]byte, len(firstHash))
	sr := make([]byte, len(firstHash))
	copy(fr, firstHash)
	copy(sr, firstHash)

	for index := 1; index < len(path); index++ {
		nodeHash, err := base64.StdEncoding.DecodeString(path[index])
		if err != nil {
			return false, fmt.Errorf("consistency path element must be base64: %w", err)
		}

		if sn == 0 {
			return false, nil
		}

		if leastSignificantBit(fn) == 1 || fn == sn {
			fr = HashNode(nodeHash, fr)
			sr = HashNode(nodeHash, sr)

			if leastSignificantBit(fn) == 0 {
				for leastSignificantBit(fn) == 0 && fn != 0 {
					fn /= 2
					sn /= 2
				}
			}
		} else {
			sr = HashNode(sr, nodeHash)
		}

		fn /= 2
		sn /= 2
	}

	return sn == 0 &&
		base64.StdEncoding.EncodeToString(fr) == oldRootB64 &&
		base64.StdEncoding.EncodeToString(sr) == newRootB64, nil
}

// VerifyConsistencyProofObject verifies a consistency proof object that follows the HCS-27 draft shape.
func VerifyConsistencyProofObject(proof *ConsistencyProof) (bool, error) {
	if proof == nil {
		return false, fmt.Errorf("proof is required")
	}
	if proof.TreeVersion != 1 {
		return false, fmt.Errorf("treeVersion must be 1")
	}

	oldTreeSize, err := parseCanonicalUint64("oldTreeSize", proof.OldTreeSize)
	if err != nil {
		return false, err
	}
	newTreeSize, err := parseCanonicalUint64("newTreeSize", proof.NewTreeSize)
	if err != nil {
		return false, err
	}
	if oldTreeSize == 0 {
		return VerifyConsistencyProof(
			oldTreeSize,
			newTreeSize,
			proof.OldRootHash,
			proof.NewRootHash,
			proof.ConsistencyPath,
		)
	}
	oldRootHash, err := normalizeProofRootHash("oldRootHash", proof.OldRootHash)
	if err != nil {
		return false, err
	}
	newRootHash, err := normalizeProofRootHash("newRootHash", proof.NewRootHash)
	if err != nil {
		return false, err
	}

	return VerifyConsistencyProof(
		oldTreeSize,
		newTreeSize,
		oldRootHash,
		newRootHash,
		proof.ConsistencyPath,
	)
}

func normalizeProofRootHash(fieldName, value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", fmt.Errorf("%s is required", fieldName)
	}
	if decoded, err := base64.StdEncoding.DecodeString(trimmed); err == nil {
		return base64.StdEncoding.EncodeToString(decoded), nil
	}
	decoded, err := base64.RawURLEncoding.DecodeString(trimmed)
	if err != nil {
		return "", fmt.Errorf("%s must be valid base64 or base64url: %w", fieldName, err)
	}
	return base64.StdEncoding.EncodeToString(decoded), nil
}

// CanonicalizeJSON performs the requested operation.
func CanonicalizeJSON(value any) ([]byte, error) {
	normalized, err := normalizeJSONValue(value)
	if err != nil {
		return nil, err
	}

	var buffer bytes.Buffer
	if err := writeCanonicalJSON(&buffer, normalized); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func normalizeJSONValue(value any) (any, error) {
	switch typed := value.(type) {
	case nil, bool, string, json.Number,
		float32, float64,
		int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64:
		return typed, nil
	case []any:
		result := make([]any, 0, len(typed))
		for _, item := range typed {
			normalizedItem, err := normalizeJSONValue(item)
			if err != nil {
				return nil, err
			}
			result = append(result, normalizedItem)
		}
		return result, nil
	case map[string]any:
		result := make(map[string]any, len(typed))
		for key, item := range typed {
			normalizedItem, err := normalizeJSONValue(item)
			if err != nil {
				return nil, err
			}
			result[key] = normalizedItem
		}
		return result, nil
	default:
		payload, err := json.Marshal(typed)
		if err != nil {
			return nil, fmt.Errorf("failed to normalize JSON input: %w", err)
		}

		var parsed any
		decoder := json.NewDecoder(bytes.NewReader(payload))
		decoder.UseNumber()
		if err := decoder.Decode(&parsed); err != nil {
			return nil, fmt.Errorf("failed to decode JSON input: %w", err)
		}

		return normalizeJSONValue(parsed)
	}
}

func writeCanonicalJSON(buffer *bytes.Buffer, value any) error {
	switch typed := value.(type) {
	case nil:
		buffer.WriteString("null")
	case bool:
		if typed {
			buffer.WriteString("true")
		} else {
			buffer.WriteString("false")
		}
	case string:
		encoded, err := json.Marshal(typed)
		if err != nil {
			return err
		}
		buffer.Write(encoded)
	case json.Number:
		buffer.WriteString(typed.String())
	case float32:
		buffer.WriteString(strconv.FormatFloat(float64(typed), 'g', -1, 32))
	case float64:
		buffer.WriteString(strconv.FormatFloat(typed, 'g', -1, 64))
	case int:
		buffer.WriteString(strconv.FormatInt(int64(typed), 10))
	case int8:
		buffer.WriteString(strconv.FormatInt(int64(typed), 10))
	case int16:
		buffer.WriteString(strconv.FormatInt(int64(typed), 10))
	case int32:
		buffer.WriteString(strconv.FormatInt(int64(typed), 10))
	case int64:
		buffer.WriteString(strconv.FormatInt(typed, 10))
	case uint:
		buffer.WriteString(strconv.FormatUint(uint64(typed), 10))
	case uint8:
		buffer.WriteString(strconv.FormatUint(uint64(typed), 10))
	case uint16:
		buffer.WriteString(strconv.FormatUint(uint64(typed), 10))
	case uint32:
		buffer.WriteString(strconv.FormatUint(uint64(typed), 10))
	case uint64:
		buffer.WriteString(strconv.FormatUint(typed, 10))
	case []any:
		buffer.WriteByte('[')
		for index, item := range typed {
			if index > 0 {
				buffer.WriteByte(',')
			}
			if err := writeCanonicalJSON(buffer, item); err != nil {
				return err
			}
		}
		buffer.WriteByte(']')
	case map[string]any:
		keys := make([]string, 0, len(typed))
		for key := range typed {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		buffer.WriteByte('{')
		for index, key := range keys {
			if index > 0 {
				buffer.WriteByte(',')
			}

			encodedKey, err := json.Marshal(key)
			if err != nil {
				return err
			}
			buffer.Write(encodedKey)
			buffer.WriteByte(':')

			if err := writeCanonicalJSON(buffer, typed[key]); err != nil {
				return err
			}
		}
		buffer.WriteByte('}')
	default:
		return fmt.Errorf("unsupported JSON value type %T", typed)
	}

	return nil
}

func leastSignificantBit(value uint64) uint64 {
	return value & 1
}

func isExactPowerOfTwo(value uint64) bool {
	return value != 0 && (value&(value-1)) == 0
}

func largestPowerOfTwoLessThan(value uint64) int {
	if value <= 1 {
		return 0
	}

	result := uint64(1)
	for result<<1 < value {
		result <<= 1
	}

	return int(result)
}
