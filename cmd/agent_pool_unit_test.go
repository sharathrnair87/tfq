package cmd

import (
	"context"
	"testing"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/stretchr/testify/assert"
)

func TestListAgentPools_Unit(t *testing.T) {
	mockAgentPools := &AgentPoolsAPIMock{
		ListFunc: func(ctx context.Context, organization string, options *tfe.AgentPoolListOptions) (*tfe.AgentPoolList, error) {
			if options.PageNumber == 1 {
				return &tfe.AgentPoolList{
					Items: []*tfe.AgentPool{
						{
							ID:                 "apool-1",
							Name:               "pool-1",
							AgentCount:         1,
							OrganizationScoped: true,
							Workspaces: []*tfe.Workspace{
								{ID: "ws-1"},
							},
						},
					},
					Pagination: &tfe.Pagination{
						NextPage: 2,
					},
				}, nil
			}
			return &tfe.AgentPoolList{
				Items: []*tfe.AgentPool{
					{
						ID:   "apool-2",
						Name: "pool-2",
						AllowedWorkspaces: []*tfe.Workspace{
							{ID: "ws-2"},
						},
					},
				},
				Pagination: &tfe.Pagination{
					NextPage: 0,
				},
			}, nil
		},
	}

	agentPools, err := listAgentPools(mockAgentPools, "my-org")

	assert.NoError(t, err)
	assert.Len(t, agentPools, 2)
	assert.Equal(t, "pool-1", agentPools[0].Name)
	assert.Equal(t, []string{"ws-1"}, agentPools[0].Workspaces)
	assert.Equal(t, "pool-2", agentPools[1].Name)
	assert.Equal(t, []string{"ws-2"}, agentPools[1].AllowedWorkspaces)
}

func TestAgentPoolStruct_Unit(t *testing.T) {
	ap := AgentPool{
		ID:   "ap-1",
		Name: "Pool 1",
	}
	assert.Equal(t, "ap-1", ap.ID)
	assert.Equal(t, "Pool 1", ap.Name)
}
