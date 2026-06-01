package cmd

import (
	"context"
	"testing"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/stretchr/testify/assert"
)

func TestListPrivateModules_Unit(t *testing.T) {
	mockRegistry := &RegistryModulesAPIMock{
		ListFunc: func(ctx context.Context, organization string, options *tfe.RegistryModuleListOptions) (*tfe.RegistryModuleList, error) {
			if options.PageNumber == 1 {
				return &tfe.RegistryModuleList{
					Items: []*tfe.RegistryModule{
						{
							ID:           "mod-1",
							Name:         "module-1",
							Provider:     "aws",
							RegistryName: tfe.RegistryName("private"),
							Namespace:    "my-org",
							VCSRepo: &tfe.VCSRepo{
								DisplayIdentifier: "org/repo",
							},
							VersionStatuses: []tfe.RegistryModuleVersionStatuses{
								{Version: "1.0.0", Status: "ok"},
							},
						},
					},
					Pagination: &tfe.Pagination{
						NextPage: 2,
					},
				}, nil
			}
			return &tfe.RegistryModuleList{
				Items: []*tfe.RegistryModule{
					{
						ID:   "mod-2",
						Name: "module-2",
					},
				},
				Pagination: &tfe.Pagination{
					NextPage: 0,
				},
			}, nil
		},
	}

	modules, err := listPrivateModules(mockRegistry, "my-org")

	assert.NoError(t, err)
	assert.Len(t, modules, 2)
	assert.Equal(t, "module-1", modules[0].Name)
	assert.Equal(t, "1.0.0", modules[0].ModuleLatestVersion)
	assert.Equal(t, "org/repo", modules[0].VCSRepo)
	assert.Equal(t, "module-2", modules[1].Name)
}

func TestRegistryModuleStruct_Unit(t *testing.T) {
	rm := RegistryModule{
		ID:   "mod-1",
		Name: "test-module",
	}
	assert.Equal(t, "mod-1", rm.ID)
	assert.Equal(t, "test-module", rm.Name)
}
