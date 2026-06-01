package cmd

import (
	"context"
	"testing"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/stretchr/testify/assert"
)

func TestListPolicies_Unit(t *testing.T) {
	mockPolicies := &PoliciesAPIMock{
		ListFunc: func(ctx context.Context, organization string, options *tfe.PolicyListOptions) (*tfe.PolicyList, error) {
			if options.PageNumber == 1 {
				return &tfe.PolicyList{
					Items: []*tfe.Policy{
						{
							ID:   "pol-1",
							Name: "Policy 1",
							Kind: tfe.Sentinel,
							Enforce: []*tfe.Enforcement{
								{Mode: tfe.EnforcementMandatory},
							},
						},
					},
					Pagination: &tfe.Pagination{
						NextPage: 2,
					},
				}, nil
			}
			return &tfe.PolicyList{
				Items: []*tfe.Policy{
					{
						ID:   "pol-2",
						Name: "Policy 2",
						Kind: tfe.OPA,
						Enforce: []*tfe.Enforcement{
							{Mode: tfe.EnforcementAdvisory},
						},
					},
				},
				Pagination: &tfe.Pagination{
					NextPage: 0,
				},
			}, nil
		},
	}

	policies, err := listPolicies(mockPolicies, "org-1", "")
	assert.NoError(t, err)
	assert.Len(t, policies, 2)
	assert.Equal(t, "Policy 1", policies[0].Name)
	assert.Equal(t, "Policy 2", policies[1].Name)
}

func TestListPolicies_Filter_Unit(t *testing.T) {
	mockPolicies := &PoliciesAPIMock{
		ListFunc: func(ctx context.Context, organization string, options *tfe.PolicyListOptions) (*tfe.PolicyList, error) {
			assert.Equal(t, "my-search", options.Search)
			return &tfe.PolicyList{
				Items: []*tfe.Policy{
					{ID: "pol-1", Name: "my-search"},
				},
				Pagination: &tfe.Pagination{NextPage: 0},
			}, nil
		},
	}

	policies, err := listPolicies(mockPolicies, "org-1", "my-search")
	assert.NoError(t, err)
	assert.Len(t, policies, 1)
}

func TestPolicyStruct_Unit(t *testing.T) {
	p := Policy{
		ID:             "pol-1",
		Name:           "Test",
		Kind:           "sentinel",
		Enforce:        "hard-mandatory",
		PolicySetCount: 5,
	}
	assert.Equal(t, "pol-1", p.ID)
	assert.Equal(t, "sentinel", p.Kind)
	assert.Equal(t, 5, p.PolicySetCount)
}
