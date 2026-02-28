package hcs7

import "testing"

func TestNewClientMissingOperatorID(t *testing.T) {
	_, err := NewClient(ClientConfig{
		OperatorPrivateKey: "302e020100300506032b6570042204200f1fd57ad073188fc1a3f0c174f2f7ce5eb58f3d82405463e99e36fcff7bcac6",
		Network:            "testnet",
	})
	if err == nil {
		t.Fatalf("expected error for missing operator account ID")
	}
}

