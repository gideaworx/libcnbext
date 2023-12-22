package registry

import (
	"context"

	"github.com/docker/docker/client"
)

func ImageExists(ctx context.Context, imageURI string) (bool, error) {
	client, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return false, err
	}

	auth := GetEncodedAuth(ctx, imageURI)
	_, err = client.DistributionInspect(ctx, imageURI, auth)
	return err == nil, err
}
