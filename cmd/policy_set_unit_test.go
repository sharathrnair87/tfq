package cmd

import (
	"context"
	"testing"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/stretchr/testify/assert"
)

func TestListPolicySets_Unit(t *testing.T) {
	mockPolicySets := &PolicySetsAPIMock{
		ListFunc: func(ctx context.Context, organization string, options *tfe.PolicySetListOptions) (*tfe.PolicySetList, error) {
			if options.PageNumber == 1 {
				return &tfe.PolicySetList{
					Items: []*tfe.PolicySet{
						{ID: "ps-1", Name: "Policy Set 1"},
					},
					Pagination: &tfe.Pagination{
						NextPage: 2,
					},
				}, nil
			}
			return &tfe.PolicySetList{
				Items: []*tfe.PolicySet{
					{ID: "ps-2", Name: "Policy Set 2"},
				},
				Pagination: &tfe.Pagination{
					NextPage: 0,
				},
			}, nil
		},
	}

	ps, err := listPolicySets(mockPolicySets, "my-org", "")

	assert.NoError(t, err)
	assert.Len(t, ps, 2)
	assert.Equal(t, "Policy Set 1", ps[0].Name)
	assert.Equal(t, "Policy Set 2", ps[1].Name)
}

func TestPolicySetStruct_Unit(t *testing.T) {
	ps := PolicySet{
		ID:   "ps-1",
		Name: "Test",
	}
	assert.Equal(t, "ps-1", ps.ID)
	assert.Equal(t, "Test", ps.Name)
}
