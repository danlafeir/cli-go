//go:build darwin

/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package secrets

import (
	"fmt"

	"github.com/keybase/go-keychain"
)

// RealSecrets implements SecretsProvider using the system keychain
type RealSecrets struct {
	appName string
}

// NewRealSecrets creates a RealSecrets provider with the given app name.
func NewRealSecrets(appName string) *RealSecrets {
	return &RealSecrets{appName: appName}
}

// buildServiceName creates the service name using the naming convention: cli.<appName>.<namespace>
func (r *RealSecrets) buildServiceName(namespace string) string {
	return fmt.Sprintf("cli.%s.%s", r.appName, namespace)
}

func (r *RealSecrets) Read(namespace, name string) (string, error) {
	serviceName := r.buildServiceName(namespace)

	data, err := keychain.GetGenericPassword(serviceName, name, "", "")
	if err != nil {
		return "", fmt.Errorf("failed to retrieve secret from keychain: %w", err)
	}

	return string(data), nil
}

func (r *RealSecrets) Write(namespace, name, value string) error {
	serviceName := r.buildServiceName(namespace)

	item := keychain.NewGenericPassword(serviceName, name, "", []byte(value), "")
	item.SetSynchronizable(keychain.SynchronizableNo)
	item.SetAccessible(keychain.AccessibleWhenUnlocked)

	// Try to add the item
	err := keychain.AddItem(item)
	if err != nil {
		// If add fails (item might exist), try to update
		queryItem := keychain.NewGenericPassword(serviceName, name, "", nil, "")
		updateItem := keychain.NewGenericPassword(serviceName, name, "", []byte(value), "")
		err = keychain.UpdateItem(queryItem, updateItem)
		if err != nil {
			return fmt.Errorf("failed to store secret in keychain: %w", err)
		}
	}

	return nil
}

func (r *RealSecrets) List(namespace string) ([]string, error) {
	serviceName := r.buildServiceName(namespace)

	// Use the convenience function to get accounts for a service
	accounts, err := keychain.GetAccountsForService(serviceName)
	if err != nil {
		// If no items found, return empty list
		return []string{}, nil
	}

	return accounts, nil
}

func (r *RealSecrets) Delete(namespace, name string) error {
	serviceName := r.buildServiceName(namespace)

	err := keychain.DeleteGenericPasswordItem(serviceName, name)
	if err != nil {
		return fmt.Errorf("failed to delete secret from keychain: %w", err)
	}

	return nil
}

