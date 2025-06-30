package modal

// config.go houses the logic for loading and resolving Modal profiles
// from ~/.modal.toml or environment variables.

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

// Profile holds a fully-resolved configuration ready for use by the client.
type Profile struct {
	ServerURL           string // e.g. https://api.modal.com:443
	TokenId             string // optional (if InitializeClient is called)
	TokenSecret         string // optional (if InitializeClient is called)
	Environment         string // optional
	ImageBuilderVersion string // optional
}

// rawProfile mirrors the TOML structure on disk.
type rawProfile struct {
	ServerURL           string `toml:"server_url"`
	TokenId             string `toml:"token_id"`
	TokenSecret         string `toml:"token_secret"`
	Environment         string `toml:"environment"`
	ImageBuilderVersion string `toml:"image_builder_version"`
	Active              bool   `toml:"active"`
}

type config map[string]rawProfile

// readConfigFile loads ~/.modal.toml, returning an empty config if the file
// does not exist.
func readConfigFile() (config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("cannot locate homedir: %w", err)
	}
	path := filepath.Join(home, ".modal.toml")
	content, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return config{}, nil // silent absence is fine
	} else if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}

	var cfg config
	if err := toml.Unmarshal(content, &cfg); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}
	return cfg, nil
}

// getProfile resolves a profile by name. Pass an empty string to instead return
// the first profile in the configuration file with `active = true`.
//
// Returned Profile is ready for use; error describes what is missing.
func getProfile(name string) Profile {
	if name == "" {
		for n, p := range defaultConfig {
			if p.Active {
				name = n
				break
			}
		}
	}

	var raw rawProfile
	if name != "" {
		raw = defaultConfig[name]
	}

	// Env-vars override file values.
	serverURL := firstNonEmpty(os.Getenv("MODAL_SERVER_URL"), raw.ServerURL, "https://api.modal.com:443")
	tokenId := firstNonEmpty(os.Getenv("MODAL_TOKEN_ID"), raw.TokenId)
	tokenSecret := firstNonEmpty(os.Getenv("MODAL_TOKEN_SECRET"), raw.TokenSecret)
	environment := firstNonEmpty(os.Getenv("MODAL_ENVIRONMENT"), raw.Environment)
	imageBuilderVersion := firstNonEmpty(os.Getenv("MODAL_IMAGE_BUILDER_VERSION"), raw.ImageBuilderVersion)

	return Profile{
		ServerURL:           serverURL,
		TokenId:             tokenId,
		TokenSecret:         tokenSecret,
		Environment:         environment,
		ImageBuilderVersion: imageBuilderVersion,
	}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

func environmentName(environment string) string {
	return firstNonEmpty(environment, clientProfile.Environment)
}

func imageBuilderVersion(version string) string {
	return firstNonEmpty(version, clientProfile.ImageBuilderVersion, "2024.10")
}
