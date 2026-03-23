package secrets

// OAuthClient represents an OAuth client configuration
type OAuthClient struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	TokenURL     string `json:"token_url"`
	Scopes       string `json:"scopes"`
	Audience     string `json:"audience"`
}

// SecretsProvider defines the generic interface for secrets operations
// This allows for mocking in tests and provides standard CRUD operations.
// Keys are stored with the naming convention: cli.<appName>.<namespace>.<name>
type SecretsProvider interface {
	Read(namespace, name string) (string, error)
	Write(namespace, name, value string) error
	List(namespace string) ([]string, error)
	Delete(namespace, name string) error
}

// defaultProvider is the secrets provider used by package-level functions.
// Call SetDefaultProvider before using the package-level Read/Write/List/Delete functions.
var defaultProvider SecretsProvider

// SetDefaultProvider initializes the package-level secrets provider with the given app name.
func SetDefaultProvider(appName string) {
	defaultProvider = NewRealSecrets(appName)
}

// Read retrieves a secret value by namespace and name
func Read(namespace, name string) (string, error) {
	return defaultProvider.Read(namespace, name)
}

// Write stores a secret value by namespace and name
func Write(namespace, name, value string) error {
	return defaultProvider.Write(namespace, name, value)
}

// List returns all secret names within a namespace
func List(namespace string) ([]string, error) {
	return defaultProvider.List(namespace)
}

// Delete removes a secret by namespace and name
func Delete(namespace, name string) error {
	return defaultProvider.Delete(namespace, name)
}
