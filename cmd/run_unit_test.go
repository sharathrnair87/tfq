package cmd

import (
	"context"
	"testing"
	"time"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/stretchr/testify/assert"
)

func TestListRuns_Unit(t *testing.T) {
	mockRuns := &RunsAPIMock{
		ListFunc: func(ctx context.Context, workspaceID string, options *tfe.RunListOptions) (*tfe.RunList, error) {
			if options.PageNumber == 1 {
				return &tfe.RunList{
					Items: []*tfe.Run{
						{ID: "run-1", Status: tfe.RunApplied},
					},
					Pagination: &tfe.Pagination{
						NextPage: 2,
					},
				}, nil
			}
			return &tfe.RunList{
				Items: []*tfe.Run{
					{ID: "run-2", Status: tfe.RunPlanned},
				},
				Pagination: &tfe.Pagination{
					NextPage: 0,
				},
			}, nil
		},
	}

	// Test with listAll = true
	runs, err := listRuns(mockRuns, "ws-1", "", "", true)
	assert.NoError(t, err)
	assert.Len(t, runs, 2)
	assert.Equal(t, "run-1", runs[0].ID)
	assert.Equal(t, "run-2", runs[1].ID)

	// Test with listAll = false
	runs, err = listRuns(mockRuns, "ws-1", "", "", false)
	assert.NoError(t, err)
	assert.Len(t, runs, 1)
	assert.Equal(t, "run-1", runs[0].ID)
}

func TestQueueRun_Unit(t *testing.T) {
	mockRuns := &RunsAPIMock{
		CreateFunc: func(ctx context.Context, options tfe.RunCreateOptions) (*tfe.Run, error) {
			assert.Contains(t, *options.Message, "Queue plan on TestWS")
			return &tfe.Run{
				ID:        "run-123",
				Status:    tfe.RunPending,
				CreatedAt: time.Now(),
				Plan:      &tfe.Plan{ID: "plan-1"},
			}, nil
		},
	}

	ws := &tfe.Workspace{
		Name: "TestWS",
		ID:   "ws-1",
	}

	run, err := queueRun(mockRuns, "my-org", ws)
	assert.NoError(t, err)
	assert.Equal(t, "run-123", run.ID)
	assert.Equal(t, tfe.RunPending, run.Status)
}

func TestApplyRun_Unit(t *testing.T) {
	applied := false
	mockRuns := &RunsAPIMock{
		ApplyFunc: func(ctx context.Context, runID string, options tfe.RunApplyOptions) error {
			assert.Equal(t, "run-1", runID)
			applied = true
			return nil
		},
	}

	applyRun(mockRuns, "run-1")
	assert.True(t, applied)
}

func TestGetRun_Unit(t *testing.T) {
	mockRuns := &RunsAPIMock{
		ReadFunc: func(ctx context.Context, runID string) (*tfe.Run, error) {
			return &tfe.Run{
				ID:     runID,
				Status: tfe.RunApplied,
			}, nil
		},
	}

	run, err := getRun(mockRuns, "run-1")
	assert.NoError(t, err)
	assert.Equal(t, "run-1", run.ID)
}

func TestCancelRun_Unit(t *testing.T) {
	cancelled := false
	mockRuns := &RunsAPIMock{
		CancelFunc: func(ctx context.Context, runID string, options tfe.RunCancelOptions) error {
			assert.Equal(t, "run-1", runID)
			cancelled = true
			return nil
		},
	}

	cancelRun(mockRuns, "run-1")
	assert.True(t, cancelled)
}

func TestForceCancelRun_Unit(t *testing.T) {
	forceCancelled := false
	mockRuns := &RunsAPIMock{
		ForceCancelFunc: func(ctx context.Context, runID string, options tfe.RunForceCancelOptions) error {
			assert.Equal(t, "run-1", runID)
			forceCancelled = true
			return nil
		},
	}

	forceCancelRun(mockRuns, "run-1")
	assert.True(t, forceCancelled)
}

func TestDiscardRun_Unit(t *testing.T) {
	discarded := false
	mockRuns := &RunsAPIMock{
		DiscardFunc: func(ctx context.Context, runID string, options tfe.RunDiscardOptions) error {
			assert.Equal(t, "run-1", runID)
			discarded = true
			return nil
		},
	}

	discardRun(mockRuns, "run-1")
	assert.True(t, discarded)
}
