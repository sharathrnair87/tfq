package cmd

import (
	"context"
	"testing"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/stretchr/testify/assert"
)

func TestShowPolicyChecks_Unit(t *testing.T) {
	mockPC := &PolicyChecksAPIMock{
		ListFunc: func(ctx context.Context, runID string, options *tfe.PolicyCheckListOptions) (*tfe.PolicyCheckList, error) {
			return &tfe.PolicyCheckList{
				Items: []*tfe.PolicyCheck{
					{
						ID:     "pc-123",
						Status: tfe.PolicyStatus("passed"),
						Result: &tfe.PolicyResult{
							Passed: 10,
							Result: true,
						},
					},
				},
			}, nil
		},
	}

	pc, err := showPolicyChecks(mockPC, "run-1")

	assert.NoError(t, err)
	assert.Equal(t, "pc-123", pc.ID)
	assert.True(t, pc.Result.Result)
	assert.Equal(t, 10, pc.Result.Passed)
}

func TestOverridePolicyChecks_Unit(t *testing.T) {
	mockPC := &PolicyChecksAPIMock{
		OverrideFunc: func(ctx context.Context, policyCheckID string) (*tfe.PolicyCheck, error) {
			return &tfe.PolicyCheck{
				ID:     policyCheckID,
				Status: tfe.PolicyStatus("overridden"),
				Result: &tfe.PolicyResult{
					Result: true,
				},
			}, nil
		},
	}

	pc, err := overridePolicyChecks(mockPC, "pc-123")

	assert.NoError(t, err)
	assert.Equal(t, "pc-123", pc.ID)
	assert.Equal(t, tfe.PolicyStatus("overridden"), pc.Status)
}
