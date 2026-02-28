package hcs27

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"testing"
)

func TestEmptyRoot(t *testing.T) {
	root := EmptyRoot()
	expected := sha256.Sum256([]byte{})
	if hex.EncodeToString(root) != hex.EncodeToString(expected[:]) {
		t.Fatal("EmptyRoot mismatch")
	}
}

func TestHashLeaf(t *testing.T) {
	data := []byte("hello")
	result := HashLeaf(data)
	if len(result) != sha256.Size {
		t.Fatalf("expected %d bytes, got %d", sha256.Size, len(result))
	}
	payload := make([]byte, 1+len(data))
	payload[0] = 0x00
	copy(payload[1:], data)
	expected := sha256.Sum256(payload)
	if hex.EncodeToString(result) != hex.EncodeToString(expected[:]) {
		t.Fatal("HashLeaf mismatch")
	}
}

func TestHashNode(t *testing.T) {
	left := HashLeaf([]byte("left"))
	right := HashLeaf([]byte("right"))
	result := HashNode(left, right)
	if len(result) != sha256.Size {
		t.Fatalf("expected %d bytes, got %d", sha256.Size, len(result))
	}
}

func TestMerkleRootFromCanonicalEntriesEmpty(t *testing.T) {
	root := MerkleRootFromCanonicalEntries(nil)
	expected := EmptyRoot()
	if hex.EncodeToString(root) != hex.EncodeToString(expected) {
		t.Fatal("empty root mismatch")
	}
}

func TestMerkleRootFromCanonicalEntriesSingle(t *testing.T) {
	entry := []byte("single")
	root := MerkleRootFromCanonicalEntries([][]byte{entry})
	expected := HashLeaf(entry)
	if hex.EncodeToString(root) != hex.EncodeToString(expected) {
		t.Fatal("single entry root mismatch")
	}
}

func TestMerkleRootFromCanonicalEntriesMultiple(t *testing.T) {
	entries := [][]byte{[]byte("a"), []byte("b"), []byte("c")}
	root := MerkleRootFromCanonicalEntries(entries)
	if len(root) != sha256.Size {
		t.Fatalf("expected %d bytes, got %d", sha256.Size, len(root))
	}
}

func TestMerkleRootFromCanonicalEntriesFour(t *testing.T) {
	entries := [][]byte{[]byte("a"), []byte("b"), []byte("c"), []byte("d")}
	root := MerkleRootFromCanonicalEntries(entries)
	left := HashNode(HashLeaf([]byte("a")), HashLeaf([]byte("b")))
	right := HashNode(HashLeaf([]byte("c")), HashLeaf([]byte("d")))
	expected := HashNode(left, right)
	if hex.EncodeToString(root) != hex.EncodeToString(expected) {
		t.Fatal("four entry root mismatch")
	}
}

func TestMerkleRootFromEntries(t *testing.T) {
	entries := []any{
		map[string]any{"key": "value1"},
		map[string]any{"key": "value2"},
	}
	root, err := MerkleRootFromEntries(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(root) != sha256.Size {
		t.Fatalf("expected %d bytes, got %d", sha256.Size, len(root))
	}
}

func TestMerkleRootFromEntriesEmpty(t *testing.T) {
	root, err := MerkleRootFromEntries(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := EmptyRoot()
	if hex.EncodeToString(root) != hex.EncodeToString(expected) {
		t.Fatal("empty entries root mismatch")
	}
}

func TestLeafHashHexFromEntryCoverage(t *testing.T) {
	entry := map[string]any{"key": "value"}
	hexStr, err := LeafHashHexFromEntry(entry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hexStr) != sha256.Size*2 {
		t.Fatalf("expected %d hex chars, got %d", sha256.Size*2, len(hexStr))
	}
}

func TestCanonicalizeJSONSortedKeys(t *testing.T) {
	input := map[string]any{"b": 2, "a": 1}
	result, err := CanonicalizeJSON(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(result) != `{"a":2,"b":1}` && string(result) != `{"a":1,"b":2}` {
		if string(result) == `{"b":2,"a":1}` {
			t.Fatalf("keys not sorted: %s", string(result))
		}
	}
}

func TestCanonicalizeJSONPrimitives(t *testing.T) {
	cases := []struct {
		input    any
		expected string
	}{
		{nil, "null"},
		{true, "true"},
		{false, "false"},
		{"hello", `"hello"`},
		{json.Number("42"), "42"},
		{float32(3.14), "3.14"},
		{float64(2.718), "2.718"},
		{int(10), "10"},
		{int8(8), "8"},
		{int16(16), "16"},
		{int32(32), "32"},
		{int64(64), "64"},
		{uint(10), "10"},
		{uint8(8), "8"},
		{uint16(16), "16"},
		{uint32(32), "32"},
		{uint64(64), "64"},
	}
	for _, tc := range cases {
		result, err := CanonicalizeJSON(tc.input)
		if err != nil {
			t.Fatalf("unexpected error for %v: %v", tc.input, err)
		}
		if string(result) != tc.expected {
			t.Fatalf("expected %q for %v, got %q", tc.expected, tc.input, string(result))
		}
	}
}

func TestCanonicalizeJSONArray(t *testing.T) {
	input := []any{"b", "a"}
	result, err := CanonicalizeJSON(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(result) != `["b","a"]` {
		t.Fatalf("unexpected: %s", string(result))
	}
}

func TestCanonicalizeJSONNestedObject(t *testing.T) {
	input := map[string]any{
		"z": map[string]any{"b": 2, "a": 1},
		"a": []any{3, 2, 1},
	}
	result, err := CanonicalizeJSON(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := `{"a":[3,2,1],"z":{"a":1,"b":2}}`
	if string(result) != expected {
		t.Fatalf("unexpected: %s", string(result))
	}
}

func TestCanonicalizeJSONCustomStruct(t *testing.T) {
	type custom struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	result, err := CanonicalizeJSON(custom{Name: "Alice", Age: 30})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(result) != `{"age":30,"name":"Alice"}` {
		t.Fatalf("unexpected: %s", string(result))
	}
}

func TestVerifyInclusionProofTreeSizeZero(t *testing.T) {
	_, err := VerifyInclusionProof(0, 0, "abc", nil, "")
	if err == nil {
		t.Fatal("expected error for treeSize=0")
	}
}

func TestVerifyInclusionProofLeafIndexOOB(t *testing.T) {
	_, err := VerifyInclusionProof(5, 3, "abc", nil, "")
	if err == nil {
		t.Fatal("expected error for leafIndex >= treeSize")
	}
}

func TestVerifyInclusionProofBadHex(t *testing.T) {
	_, err := VerifyInclusionProof(0, 1, "zzzz", nil, "")
	if err == nil {
		t.Fatal("expected error for bad hex")
	}
}

func TestVerifyInclusionProofSingleLeaf(t *testing.T) {
	entry := []byte(`"hello"`)
	leafHash := HashLeaf(entry)
	leafHex := hex.EncodeToString(leafHash)
	rootB64 := base64.StdEncoding.EncodeToString(leafHash)

	ok, err := VerifyInclusionProof(0, 1, leafHex, nil, rootB64)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected inclusion proof to verify")
	}
}

func TestVerifyInclusionProofTwoLeaves(t *testing.T) {
	leaf0 := HashLeaf([]byte("a"))
	leaf1 := HashLeaf([]byte("b"))
	root := HashNode(leaf0, leaf1)

	rootB64 := base64.StdEncoding.EncodeToString(root)
	siblingB64 := base64.StdEncoding.EncodeToString(leaf1)

	ok, err := VerifyInclusionProof(0, 2, hex.EncodeToString(leaf0), []string{siblingB64}, rootB64)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected inclusion proof to verify")
	}
}

func TestVerifyInclusionProofBadPath(t *testing.T) {
	leaf0 := HashLeaf([]byte("a"))
	_, err := VerifyInclusionProof(0, 2, hex.EncodeToString(leaf0), []string{"!!!notbase64!!!"}, "")
	if err == nil {
		t.Fatal("expected error for bad base64 path element")
	}
}

func TestVerifyConsistencyProofSameSize(t *testing.T) {
	ok, err := VerifyConsistencyProof(5, 5, "abc", "abc", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected true for same size same root")
	}
}

func TestVerifyConsistencyProofSameSizeDiffRoot(t *testing.T) {
	ok, err := VerifyConsistencyProof(5, 5, "abc", "xyz", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatal("expected false for same size diff root")
	}
}

func TestVerifyConsistencyProofOldGtNew(t *testing.T) {
	ok, err := VerifyConsistencyProof(10, 5, "abc", "xyz", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatal("expected false for old > new")
	}
}

func TestVerifyConsistencyProofEmptyPath(t *testing.T) {
	ok, err := VerifyConsistencyProof(3, 5, "abc", "xyz", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatal("expected false for empty path")
	}
}

func TestVerifyConsistencyProofBadBase64(t *testing.T) {
	_, err := VerifyConsistencyProof(1, 2, "abc", "xyz", []string{"!!!bad!!!"})
	if err == nil {
		t.Fatal("expected error for bad base64")
	}
}

func TestLargestPowerOfTwoLessThan(t *testing.T) {
	cases := []struct {
		input    uint64
		expected int
	}{
		{0, 0}, {1, 0}, {2, 1}, {3, 2}, {4, 2}, {5, 4}, {8, 4}, {9, 8}, {16, 8},
	}
	for _, tc := range cases {
		result := largestPowerOfTwoLessThan(tc.input)
		if result != tc.expected {
			t.Fatalf("largestPowerOfTwoLessThan(%d) = %d, want %d", tc.input, result, tc.expected)
		}
	}
}

func TestIsExactPowerOfTwo(t *testing.T) {
	cases := []struct {
		input    uint64
		expected bool
	}{
		{0, false}, {1, true}, {2, true}, {3, false}, {4, true}, {8, true}, {9, false},
	}
	for _, tc := range cases {
		result := isExactPowerOfTwo(tc.input)
		if result != tc.expected {
			t.Fatalf("isExactPowerOfTwo(%d) = %v, want %v", tc.input, result, tc.expected)
		}
	}
}

func TestLeastSignificantBit(t *testing.T) {
	if leastSignificantBit(0) != 0 {
		t.Fatal("expected 0 for 0")
	}
	if leastSignificantBit(1) != 1 {
		t.Fatal("expected 1 for 1")
	}
	if leastSignificantBit(2) != 0 {
		t.Fatal("expected 0 for 2")
	}
	if leastSignificantBit(3) != 1 {
		t.Fatal("expected 1 for 3")
	}
}
