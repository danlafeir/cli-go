/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package secrets

import (
	"encoding/json"
	"testing"
)

func TestOAuthClient_JSON(t *testing.T) {
	client := &OAuthClient{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		TokenURL:     "https://example.com/oauth/token",
		Scopes:       "read write",
		Audience:     "https://api.example.com",
	}

	// Test marshaling
	data, err := json.Marshal(client)
	if err != nil {
		t.Fatalf("Failed to marshal OAuth client: %v", err)
	}

	// Test unmarshaling
	var unmarshaledClient OAuthClient
	err = json.Unmarshal(data, &unmarshaledClient)
	if err != nil {
		t.Fatalf("Failed to unmarshal OAuth client: %v", err)
	}

	// Verify fields match
	if unmarshaledClient.ClientID != client.ClientID {
		t.Errorf("ClientID mismatch: got %s, want %s", unmarshaledClient.ClientID, client.ClientID)
	}
	if unmarshaledClient.ClientSecret != client.ClientSecret {
		t.Errorf("ClientSecret mismatch: got %s, want %s", unmarshaledClient.ClientSecret, client.ClientSecret)
	}
	if unmarshaledClient.TokenURL != client.TokenURL {
		t.Errorf("TokenURL mismatch: got %s, want %s", unmarshaledClient.TokenURL, client.TokenURL)
	}
	if unmarshaledClient.Scopes != client.Scopes {
		t.Errorf("Scopes mismatch: got %s, want %s", unmarshaledClient.Scopes, client.Scopes)
	}
	if unmarshaledClient.Audience != client.Audience {
		t.Errorf("Audience mismatch: got %s, want %s", unmarshaledClient.Audience, client.Audience)
	}
}
