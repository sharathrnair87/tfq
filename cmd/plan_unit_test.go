package cmd

import (
	"context"
	"testing"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/stretchr/testify/assert"
)

func TestShowPlan_Unit(t *testing.T) {
	mockPlans := &PlansAPIMock{
		ReadFunc: func(ctx context.Context, planID string) (*tfe.Plan, error) {
			return &tfe.Plan{
				ID:                   planID,
				Status:               tfe.PlanFinished,
				HasChanges:           true,
				ResourceAdditions:    1,
				ResourceChanges:      2,
				ResourceDestructions: 3,
				ResourceImports:      4,
			}, nil
		},
	}

	plan, err := showPlan(mockPlans, "plan-123", false)

	assert.NoError(t, err)
	assert.Equal(t, "plan-123", plan.ID)
	assert.Equal(t, "finished", plan.Status)
	assert.True(t, plan.HasChanges)
	assert.Equal(t, 1, plan.ResourceAdditions)
	assert.Equal(t, 2, plan.ResourceChanges)
	assert.Equal(t, 3, plan.ResourceDestructions)
	assert.Equal(t, 4, plan.ResourceImports)
	assert.Nil(t, plan.ChangedResourceProperties)
}

func TestShowPlan_DetailedChanges_Unit(t *testing.T) {
	// Mock JSON output from TFE
	mockJSON := `{
		"resource_changes": [
			{
				"address": "null_resource.test",
				"change": {
					"actions": ["update"],
					"before": { "foo": "bar" },
					"after": { "foo": "baz" }
				}
			}
		]
	}`

	mockPlans := &PlansAPIMock{
		ReadFunc: func(ctx context.Context, planID string) (*tfe.Plan, error) {
			return &tfe.Plan{
				ID:     planID,
				Status: tfe.PlanFinished,
			}, nil
		},
		ReadJSONOutputFunc: func(ctx context.Context, planID string) ([]byte, error) {
			return []byte(mockJSON), nil
		},
	}

	plan, err := showPlan(mockPlans, "plan-123", true)

	assert.NoError(t, err)
	assert.NotNil(t, plan.ChangedResourceProperties)

	// ChangedResourceProperties should be populated by jq result
	// Note: result depends on resources.JqRun implementation and the query string in plan.go
}

func TestPlanStruct_Unit(t *testing.T) {
	p := Plan{
		ID:         "p1",
		HasChanges: true,
		Status:     "finished",
	}
	assert.Equal(t, "p1", p.ID)
	assert.True(t, p.HasChanges)
	assert.Equal(t, "finished", p.Status)
}
