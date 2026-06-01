package cmd

import (
	"context"
	"testing"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/stretchr/testify/assert"
)

func TestListAdminRuns_Unit(t *testing.T) {
	mockAdminRuns := &AdminRunsAPIMock{
		ListFunc: func(ctx context.Context, options *tfe.AdminRunsListOptions) (*tfe.AdminRunsList, error) {
			if options.PageNumber == 1 {
				return &tfe.AdminRunsList{
					Items: []*tfe.AdminRun{
						{ID: "run-1", Status: tfe.RunApplied},
					},
					Pagination: &tfe.Pagination{
						NextPage: 2,
					},
				}, nil
			}
			return &tfe.AdminRunsList{
				Items: []*tfe.AdminRun{
					{ID: "run-2", Status: tfe.RunPlanned},
				},
				Pagination: &tfe.Pagination{
					NextPage: 0,
				},
			}, nil
		},
	}

	adminRuns, err := listAdminRuns(mockAdminRuns, "")
	assert.NoError(t, err)
	assert.Len(t, adminRuns, 2)
	assert.Equal(t, "run-1", adminRuns[0].ID)
	assert.Equal(t, "run-2", adminRuns[1].ID)
}

func TestListAdminRuns_Filter_Unit(t *testing.T) {
	mockAdminRuns := &AdminRunsAPIMock{
		ListFunc: func(ctx context.Context, options *tfe.AdminRunsListOptions) (*tfe.AdminRunsList, error) {
			assert.Equal(t, "pending", options.RunStatus)
			return &tfe.AdminRunsList{
				Items: []*tfe.AdminRun{
					{ID: "run-1", Status: tfe.RunPending},
				},
				Pagination: &tfe.Pagination{NextPage: 0},
			}, nil
		},
	}

	adminRuns, err := listAdminRuns(mockAdminRuns, "pending")
	assert.NoError(t, err)
	assert.Len(t, adminRuns, 1)
	assert.Equal(t, tfe.RunPending, adminRuns[0].Status)
}

func TestForceCancelAdminRuns_Unit(t *testing.T) {
	cancelled := false
	mockAdminRuns := &AdminRunsAPIMock{
		ForceCancelFunc: func(ctx context.Context, runID string, options tfe.AdminRunForceCancelOptions) error {
			assert.Equal(t, "run-1", runID)
			cancelled = true
			return nil
		},
	}

	adminForceCancelRun(mockAdminRuns, "run-1")
	assert.True(t, cancelled)
}
