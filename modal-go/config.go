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
	TokenId             string
	TokenSecret         string
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

// GetProfile resolves a profile by name.  Pass an empty string to follow the
// same precedence as the TypeScript original:
//
//  1. MODAL_PROFILE env var
//  2. first profile in the file with active = true
//
// Returned Profile is ready for use; error describes what is missing.
func GetProfile(name string) (Profile, error) {
	// 1. explicit argument overrides everything
	if name == "" {
		name = os.Getenv("MODAL_PROFILE")
	}

	// 2. look for `active = true` if still unspecified
	if name == "" {
		for n, p := range defaultConfig {
			if p.Active {
				name = n
				break
			}
		}
	}

	// 3. verify existence
	raw, ok := defaultConfig[name]
	if name != "" && !ok {
		return Profile{}, fmt.Errorf("profile %q not found in ~/.modal.toml", name)
	}

	// 4. env-vars override file values
	serverURL := firstNonEmpty(os.Getenv("MODAL_SERVER_URL"), raw.ServerURL, "https://api.modal.com:443")
	tokenId := firstNonEmpty(os.Getenv("MODAL_TOKEN_ID"), raw.TokenId)
	tokenSecret := firstNonEmpty(os.Getenv("MODAL_TOKEN_SECRET"), raw.TokenSecret)
	environment := firstNonEmpty(os.Getenv("MODAL_ENVIRONMENT"), raw.Environment)
	imageBuilderVersion := firstNonEmpty(os.Getenv("MODAL_IMAGE_BUILDER_VERSION"), raw.ImageBuilderVersion)

	if tokenId == "" || tokenSecret == "" {
		return Profile{}, fmt.Errorf("profile %q missing token_id or token_secret (env-vars take precedence)", name)
	}

	return Profile{
		ServerURL:           serverURL,
		TokenId:             tokenId,
		TokenSecret:         tokenSecret,
		Environment:         environment,
		ImageBuilderVersion: imageBuilderVersion,
	}, nil
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
	return firstNonEmpty(environment, defaultProfile.Environment)
}

func imageBuilderVersion(version string) string {
	return firstNonEmpty(version, defaultProfile.ImageBuilderVersion)
}
