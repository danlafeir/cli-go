//go:build linux

/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package secrets

import (
	"fmt"

	"github.com/keybase/dbus"
	"github.com/keybase/go-keychain/secretservice"
)

// RealSecrets implements SecretsProvider using the Linux Secret Service API
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

// openSession connects to the Secret Service and opens an encrypted session.
func (r *RealSecrets) openSession() (*secretservice.SecretService, *secretservice.Session, error) {
	svc, err := secretservice.NewService()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to secret service: %w", err)
	}

	session, err := svc.OpenSession(secretservice.AuthenticationDHAES)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open secret service session: %w", err)
	}

	err = svc.Unlock([]dbus.ObjectPath{secretservice.DefaultCollection})
	if err != nil {
		svc.CloseSession(session)
		return nil, nil, fmt.Errorf("failed to unlock default collection: %w", err)
	}

	return svc, session, nil
}

func (r *RealSecrets) Read(namespace, name string) (string, error) {
	serviceName := r.buildServiceName(namespace)

	svc, session, err := r.openSession()
	if err != nil {
		return "", err
	}
	defer svc.CloseSession(session)

	attrs := secretservice.Attributes{
		"service": serviceName,
		"account": name,
	}

	items, err := svc.SearchCollection(secretservice.DefaultCollection, attrs)
	if err != nil {
		return "", fmt.Errorf("failed to search secret service: %w", err)
	}
	if len(items) == 0 {
		return "", fmt.Errorf("secret not found: %s/%s", namespace, name)
	}

	data, err := svc.GetSecret(items[0], *session)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve secret: %w", err)
	}

	return string(data), nil
}

func (r *RealSecrets) Write(namespace, name, value string) error {
	serviceName := r.buildServiceName(namespace)

	svc, session, err := r.openSession()
	if err != nil {
		return err
	}
	defer svc.CloseSession(session)

	attrs := secretservice.Attributes{
		"service": serviceName,
		"account": name,
	}

	secret, err := session.NewSecret([]byte(value))
	if err != nil {
		return fmt.Errorf("failed to create secret: %w", err)
	}

	label := fmt.Sprintf("%s - %s", serviceName, name)
	properties := secretservice.NewSecretProperties(label, attrs)

	_, err = svc.CreateItem(secretservice.DefaultCollection, properties, secret, secretservice.ReplaceBehaviorReplace)
	if err != nil {
		return fmt.Errorf("failed to store secret: %w", err)
	}

	return nil
}

func (r *RealSecrets) List(namespace string) ([]string, error) {
	serviceName := r.buildServiceName(namespace)

	svc, session, err := r.openSession()
	if err != nil {
		return nil, err
	}
	defer svc.CloseSession(session)

	attrs := secretservice.Attributes{
		"service": serviceName,
	}

	items, err := svc.SearchCollection(secretservice.DefaultCollection, attrs)
	if err != nil {
		return []string{}, nil
	}

	var accounts []string
	for _, item := range items {
		itemAttrs, err := svc.GetAttributes(item)
		if err != nil {
			continue
		}
		if account, ok := itemAttrs["account"]; ok {
			accounts = append(accounts, account)
		}
	}

	return accounts, nil
}

func (r *RealSecrets) Delete(namespace, name string) error {
	serviceName := r.buildServiceName(namespace)

	svc, session, err := r.openSession()
	if err != nil {
		return err
	}
	defer svc.CloseSession(session)

	attrs := secretservice.Attributes{
		"service": serviceName,
		"account": name,
	}

	items, err := svc.SearchCollection(secretservice.DefaultCollection, attrs)
	if err != nil {
		return fmt.Errorf("failed to search secret service: %w", err)
	}
	if len(items) == 0 {
		return fmt.Errorf("secret not found: %s/%s", namespace, name)
	}

	err = svc.DeleteItem(items[0])
	if err != nil {
		return fmt.Errorf("failed to delete secret: %w", err)
	}

	return nil
}
