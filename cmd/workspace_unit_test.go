package cmd

import (
	"context"
	"testing"
	"time"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/stretchr/testify/assert"
)

func TestListWorkspaces_Unit(t *testing.T) {
	mockWorkspaces := &WorkspacesAPIMock{
		ListFunc: func(ctx context.Context, organization string, options *tfe.WorkspaceListOptions) (*tfe.WorkspaceList, error) {
			return &tfe.WorkspaceList{
				Items: []*tfe.Workspace{
					{ID: "ws-1", Name: "Workspace 1"},
					{ID: "ws-2", Name: "Workspace 2"},
				},
				Pagination: &tfe.Pagination{
					NextPage: 0,
				},
			}, nil
		},
	}

	workspaces, err := listWorkspaces(mockWorkspaces, "my-org", "")

	assert.NoError(t, err)
	assert.Len(t, workspaces, 2)
	assert.Equal(t, "Workspace 1", workspaces[0].Name)
}

func TestGetWorkspaceNameByID_Unit(t *testing.T) {
	mockWorkspaces := &WorkspacesAPIMock{
		ReadByIDFunc: func(ctx context.Context, workspaceID string) (*tfe.Workspace, error) {
			return &tfe.Workspace{ID: workspaceID, Name: "Workspace Name"}, nil
		},
	}

	name, err := getWorkspaceNameByID(mockWorkspaces, "my-org", "ws-123")

	assert.NoError(t, err)
	assert.Equal(t, "Workspace Name", name)
}

func TestGetWorkspace_Unit(t *testing.T) {
	now := time.Now()
	mockWorkspaces := &WorkspacesAPIMock{
		ReadByIDFunc: func(ctx context.Context, workspaceID string) (*tfe.Workspace, error) {
			return &tfe.Workspace{
				ID:               workspaceID,
				Name:             "WS1",
				CreatedAt:        now.Add(-24 * time.Hour),
				UpdatedAt:        now.Add(-12 * time.Hour),
				ExecutionMode:    "remote",
				TerraformVersion: "1.0.0",
				AgentPool:        &tfe.AgentPool{ID: "ap-1"},
				TagNames:         []string{"tag1"},
			}, nil
		},
	}

	mockRuns := &RunsAPIMock{
		ListFunc: func(ctx context.Context, workspaceID string, options *tfe.RunListOptions) (*tfe.RunList, error) {
			return &tfe.RunList{
				Items: []*tfe.Run{
					{
						ID:        "run-1",
						Status:    tfe.RunApplied,
						CreatedAt: now.Add(-2 * time.Hour),
						StatusTimestamps: &tfe.RunStatusTimestamps{
							AppliedAt: now.Add(-1 * time.Hour),
						},
					},
				},
			}, nil
		},
	}

	mockStateVersions := &StateVersionsAPIMock{
		ReadCurrentFunc: func(ctx context.Context, workspaceID string) (*tfe.StateVersion, error) {
			return &tfe.StateVersion{
				ID:        "sv-1",
				CreatedAt: now.Add(-3 * time.Hour),
			}, nil
		},
	}

	detail, err := getWorkspace(mockWorkspaces, mockRuns, mockStateVersions, "my-org", "ws-1")

	assert.NoError(t, err)
	assert.Equal(t, "ws-1", detail.ID)
	assert.Equal(t, "WS1", detail.Name)
	assert.Equal(t, "ap-1", detail.AgentPoolID)
	assert.Contains(t, detail.Tags, "tag1")
	// Run details
	assert.NotEqual(t, "NA", detail.LastRemoteRunDaysAgo)
	assert.NotEqual(t, "NA", detail.AverageRunDuration)
	assert.NotEqual(t, "NA", detail.LastStateUpdateDaysAgo)
}

func TestLockWorkspace_Unit(t *testing.T) {
	mockWorkspaces := &WorkspacesAPIMock{
		LockFunc: func(ctx context.Context, workspaceID string, options tfe.WorkspaceLockOptions) (*tfe.Workspace, error) {
			return &tfe.Workspace{ID: workspaceID, Locked: true}, nil
		},
	}

	reason := "Testing"
	ws, err := lockWorkspace(mockWorkspaces, "my-org", "ws-1", &reason)

	assert.NoError(t, err)
	assert.True(t, ws.Locked)
	assert.Equal(t, "ws-1", ws.ID)
}

func TestUnlockWorkspace_Unit(t *testing.T) {
	mockWorkspaces := &WorkspacesAPIMock{
		UnlockFunc: func(ctx context.Context, workspaceID string) (*tfe.Workspace, error) {
			return &tfe.Workspace{ID: workspaceID, Locked: false}, nil
		},
	}

	ws, err := unlockWorkspace(mockWorkspaces, "my-org", "ws-1")

	assert.NoError(t, err)
	assert.False(t, ws.Locked)
	assert.Equal(t, "ws-1", ws.ID)
}
