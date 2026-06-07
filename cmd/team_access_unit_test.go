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

	result, err := getWorkspaceTeamAccess(mockTeamAccess, mockWorkspaces, "ws-123", "team-abc", "my-org")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "ws-123", result.WorkspaceID)
	assert.Equal(t, "my-workspace", result.WorkspaceName)
	assert.Equal(t, "custom", result.Attributes.Access)
	assert.Equal(t, "apply", result.Attributes.Runs)
	assert.Equal(t, "write", result.Attributes.Variables)
	assert.Equal(t, "read", result.Attributes.StateVersions)
	assert.Equal(t, "none", result.Attributes.SentinelMocks)
	assert.True(t, result.Attributes.WorkspaceLocking)
	assert.False(t, result.Attributes.RunTasks)
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

	result, err := getWorkspaceTeamAccess(mockTeamAccess, mockWorkspaces, "ws-123", "team-abc", "my-org")
	assert.NoError(t, err)
	assert.Nil(t, result)
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

	result, err := getWorkspaceTeamAccess(mockTeamAccess, mockWorkspaces, "ws-123", "team-abc", "my-org")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "ws-123", result.WorkspaceID)
	assert.Equal(t, "read", result.Attributes.Access)
}

func TestGetWorkspaceTeamAccess_APIError(t *testing.T) {
	mockTeamAccess := &TeamAccessAPIMock{
		ListFunc: func(ctx context.Context, options *tfe.TeamAccessListOptions) (*tfe.TeamAccessList, error) {
			return nil, assert.AnError
		},
	}

	mockWorkspaces := &WorkspacesAPIMock{}

	result, err := getWorkspaceTeamAccess(mockTeamAccess, mockWorkspaces, "ws-123", "team-abc", "my-org")
	assert.Error(t, err)
	assert.Nil(t, result)
}
