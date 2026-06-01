package cmd

import (
	"context"
	"testing"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/stretchr/testify/assert"
)

func TestListPrivateProviders_Unit(t *testing.T) {
	mockProviders := &RegistryProvidersAPIMock{
		ListFunc: func(ctx context.Context, organization string, options *tfe.RegistryProviderListOptions) (*tfe.RegistryProviderList, error) {
			return &tfe.RegistryProviderList{
				Items: []*tfe.RegistryProvider{
					{ID: "prov-1", Name: "provider-1", Namespace: "ns-1"},
				},
				Pagination: &tfe.Pagination{NextPage: 0},
			}, nil
		},
	}

	providers, err := listPrivateProviders(mockProviders, "my-org", "")

	assert.NoError(t, err)
	assert.Len(t, providers, 1)
	assert.Equal(t, "provider-1", providers[0].Name)
}

func TestGetPrivateProviderDetails_Unit(t *testing.T) {
	mockProviders := &RegistryProvidersAPIMock{
		ListFunc: func(ctx context.Context, organization string, options *tfe.RegistryProviderListOptions) (*tfe.RegistryProviderList, error) {
			return &tfe.RegistryProviderList{
				Items: []*tfe.RegistryProvider{
					{ID: "prov-1", Name: "provider-1", Namespace: "ns-1"},
				},
				Pagination: &tfe.Pagination{NextPage: 0},
			}, nil
		},
		ReadFunc: func(ctx context.Context, providerID tfe.RegistryProviderID, options *tfe.RegistryProviderReadOptions) (*tfe.RegistryProvider, error) {
			return &tfe.RegistryProvider{ID: "prov-1", Name: "provider-1", Namespace: "ns-1"}, nil
		},
	}

	mockVersions := &RegistryProviderVersionsAPIMock{
		ListFunc: func(ctx context.Context, providerID tfe.RegistryProviderID, options *tfe.RegistryProviderVersionListOptions) (*tfe.RegistryProviderVersionList, error) {
			return &tfe.RegistryProviderVersionList{
				Items: []*tfe.RegistryProviderVersion{
					{Version: "1.0.0"},
				},
				Pagination: &tfe.Pagination{TotalPages: 1},
			}, nil
		},
	}

	mockPlatforms := &RegistryProviderPlatformsAPIMock{
		ListFunc: func(ctx context.Context, versionID tfe.RegistryProviderVersionID, options *tfe.RegistryProviderPlatformListOptions) (*tfe.RegistryProviderPlatformList, error) {
			return &tfe.RegistryProviderPlatformList{
				Items: []*tfe.RegistryProviderPlatform{
					{ID: "plat-1", OS: "linux", Arch: "amd64", Filename: "file.zip"},
				},
			}, nil
		},
	}

	detail, err := getPrivateProviderDetails(mockProviders, mockVersions, mockPlatforms, "my-org", "provider-1")

	assert.NoError(t, err)
	assert.Equal(t, "1.0.0", detail.ProviderLatestVersion)
	assert.Len(t, detail.ProviderPlatforms, 1)
	assert.Equal(t, "linux", detail.ProviderPlatforms[0].OS)
}
