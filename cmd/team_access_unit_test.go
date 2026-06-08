package cmd

import (
	"context"
	"testing"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/stretchr/testify/assert"
)

func TestGetWorkspaceTeamAccess_Found(t *testing.T) {
	mockTeamAccess := &TeamAccessAPIMock{
		ListFunc: func(ctx context.Context, options *tfe.TeamAccessListOptions) (*tfe.TeamAccessList, error) {
			assert.Equal(t, "ws-123", options.WorkspaceID)
			return &tfe.TeamAccessList{
				Items: []*tfe.TeamAccess{
					{
						ID: "ta-1",
						Team: &tfe.Team{
							ID: "team-abc",
						},
						Access:           "custom",
						Runs:             "apply",
						Variables:        "write",
						StateVersions:    "read",
						SentinelMocks:    "none",
						WorkspaceLocking: true,
						RunTasks:         false,
					},
				},
				Pagination: &tfe.Pagination{
					NextPage: 0,
				},
			}, nil
		},
	}

	mockWorkspaces := &WorkspacesAPIMock{
		ReadByIDFunc: func(ctx context.Context, workspaceID string) (*tfe.Workspace, error) {
			assert.Equal(t, "ws-123", workspaceID)
			return &tfe.Workspace{
				ID:   "ws-123",
				Name: "my-workspace",
			}, nil
		},
	}

	result, err := getWorkspaceTeamAccess(mockTeamAccess, mockWorkspaces, "ws-123", []string{"team-abc"}, "my-org")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	teamAccess, ok := result["team-abc"]
	assert.True(t, ok)
	assert.Equal(t, "ws-123", teamAccess.WorkspaceID)
	assert.Equal(t, "my-workspace", teamAccess.WorkspaceName)
	assert.Equal(t, "custom", teamAccess.Attributes.Access)
	assert.Equal(t, "apply", teamAccess.Attributes.Runs)
	assert.Equal(t, "write", teamAccess.Attributes.Variables)
	assert.Equal(t, "read", teamAccess.Attributes.StateVersions)
	assert.Equal(t, "none", teamAccess.Attributes.SentinelMocks)
	assert.True(t, teamAccess.Attributes.WorkspaceLocking)
	assert.False(t, teamAccess.Attributes.RunTasks)
}

func TestGetWorkspaceTeamAccess_NotFound(t *testing.T) {
	mockTeamAccess := &TeamAccessAPIMock{
		ListFunc: func(ctx context.Context, options *tfe.TeamAccessListOptions) (*tfe.TeamAccessList, error) {
			return &tfe.TeamAccessList{
				Items: []*tfe.TeamAccess{
					{
						ID: "ta-1",
						Team: &tfe.Team{
							ID: "team-other",
						},
					},
				},
				Pagination: &tfe.Pagination{
					NextPage: 0,
				},
			}, nil
		},
	}

	mockWorkspaces := &WorkspacesAPIMock{
		ReadByIDFunc: func(ctx context.Context, workspaceID string) (*tfe.Workspace, error) {
			return &tfe.Workspace{
				ID:   "ws-123",
				Name: "my-workspace",
			}, nil
		},
	}

	result, err := getWorkspaceTeamAccess(mockTeamAccess, mockWorkspaces, "ws-123", []string{"team-abc"}, "my-org")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Nil(t, result["team-abc"])
}

func TestGetWorkspaceTeamAccess_Pagination(t *testing.T) {
	mockTeamAccess := &TeamAccessAPIMock{
		ListFunc: func(ctx context.Context, options *tfe.TeamAccessListOptions) (*tfe.TeamAccessList, error) {
			if options.PageNumber == 1 {
				return &tfe.TeamAccessList{
					Items: []*tfe.TeamAccess{
						{
							ID: "ta-1",
							Team: &tfe.Team{
								ID: "team-other",
							},
						},
					},
					Pagination: &tfe.Pagination{
						NextPage: 2,
					},
				}, nil
			}
			return &tfe.TeamAccessList{
				Items: []*tfe.TeamAccess{
					{
						ID: "ta-2",
						Team: &tfe.Team{
							ID: "team-abc",
						},
						Access: "read",
					},
				},
				Pagination: &tfe.Pagination{
					NextPage: 0,
				},
			}, nil
		},
	}

	mockWorkspaces := &WorkspacesAPIMock{
		ReadByIDFunc: func(ctx context.Context, workspaceID string) (*tfe.Workspace, error) {
			return &tfe.Workspace{
				ID:   "ws-123",
				Name: "my-workspace",
			}, nil
		},
	}

	result, err := getWorkspaceTeamAccess(mockTeamAccess, mockWorkspaces, "ws-123", []string{"team-abc"}, "my-org")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	teamAccess, ok := result["team-abc"]
	assert.True(t, ok)
	assert.Equal(t, "ws-123", teamAccess.WorkspaceID)
	assert.Equal(t, "read", teamAccess.Attributes.Access)
}

func TestGetWorkspaceTeamAccess_APIError(t *testing.T) {
	mockTeamAccess := &TeamAccessAPIMock{
		ListFunc: func(ctx context.Context, options *tfe.TeamAccessListOptions) (*tfe.TeamAccessList, error) {
			return nil, assert.AnError
		},
	}

	mockWorkspaces := &WorkspacesAPIMock{}

	result, err := getWorkspaceTeamAccess(mockTeamAccess, mockWorkspaces, "ws-123", []string{"team-abc"}, "my-org")
	assert.Error(t, err)
	assert.Nil(t, result)
}
