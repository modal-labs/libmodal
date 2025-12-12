package modal

import (
	"context"
	"fmt"

	pb "github.com/modal-labs/libmodal/modal-go/proto/modal_proto"
)

// Environment represents a Modal Environment with its server-provided settings.
type Environment struct {
	ID       string
	Name     string
	Settings EnvironmentSettings
}

// EnvironmentSettings contains environment-scoped configuration from the server.
type EnvironmentSettings struct {
	ImageBuilderVersion string
	WebhookSuffix       string
}

// fetchEnvironment fetches an environment from the server, caching the result.
// Pass an empty string for name to get the default environment.
func (c *Client) fetchEnvironment(ctx context.Context, name string) (*Environment, error) {
	if cached, ok := c.environmentsCache.Load(name); ok {
		return cached.(*Environment), nil
	}

	result, err, _ := c.environmentGroup.Do(name, func() (interface{}, error) {
		if cached, ok := c.environmentsCache.Load(name); ok {
			return cached.(*Environment), nil
		}

		c.logger.DebugContext(ctx, "Fetching environment from server", "environment_name", name)

		resp, err := c.cpClient.EnvironmentGetOrCreate(ctx, pb.EnvironmentGetOrCreateRequest_builder{
			DeploymentName: name,
		}.Build())
		if err != nil {
			return nil, err
		}

		metadata := resp.GetMetadata()
		settings := metadata.GetSettings()

		env := &Environment{
			ID:   resp.GetEnvironmentId(),
			Name: metadata.GetName(),
			Settings: EnvironmentSettings{
				ImageBuilderVersion: settings.GetImageBuilderVersion(),
				WebhookSuffix:       settings.GetWebhookSuffix(),
			},
		}

		c.environmentsCache.Store(name, env)
		c.logger.DebugContext(ctx, "Cached environment",
			"environment_name", name,
			"environment_id", env.ID,
			"image_builder_version", env.Settings.ImageBuilderVersion)

		return env, nil
	})

	if err != nil {
		return nil, err
	}
	return result.(*Environment), nil
}

// imageBuilderVersion returns the image builder version to use for image builds.
// Precedence: local config > server-provided value.
func (c *Client) imageBuilderVersion(ctx context.Context, environmentName string) (string, error) {
	if c.profile.ImageBuilderVersion != "" {
		return c.profile.ImageBuilderVersion, nil
	}

	env, err := c.fetchEnvironment(ctx, environmentName)
	if err != nil {
		return "", fmt.Errorf("failed to get environment for image builder version: %w", err)
	}

	return env.Settings.ImageBuilderVersion, nil
}
