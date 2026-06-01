package cmd

import (
	"context"
	"testing"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/stretchr/testify/assert"
)

func TestListTeams_Unit(t *testing.T) {
	mockTeams := &TeamsAPIMock{
		ListFunc: func(ctx context.Context, organization string, options *tfe.TeamListOptions) (*tfe.TeamList, error) {
			if options.PageNumber == 1 {
				return &tfe.TeamList{
					Items: []*tfe.Team{
						{ID: "team-1", Name: "Team 1", UserCount: 2},
					},
					Pagination: &tfe.Pagination{
						CurrentPage: 1,
						NextPage:    2,
						TotalPages:  2,
					},
				}, nil
			}
			return &tfe.TeamList{
				Items: []*tfe.Team{
					{ID: "team-2", Name: "Team 2", UserCount: 5},
				},
				Pagination: &tfe.Pagination{
					CurrentPage: 2,
					NextPage:    0,
					TotalPages:  2,
				},
			}, nil
		},
	}

	teams, err := listTeams(mockTeams, "my-org", []string{})

	assert.NoError(t, err)
	assert.Len(t, teams, 2)
	assert.Equal(t, "Team 1", teams[0].Name)
	assert.Equal(t, "Team 2", teams[1].Name)
	assert.Equal(t, 2, len(mockTeams.ListCalls()))
}

func TestListTeams_Filters_Unit(t *testing.T) {
	mockTeams := &TeamsAPIMock{
		ListFunc: func(ctx context.Context, organization string, options *tfe.TeamListOptions) (*tfe.TeamList, error) {
			assert.Equal(t, []string{"FilterMe"}, options.Names)
			return &tfe.TeamList{
				Items: []*tfe.Team{
					{ID: "team-1", Name: "FilterMe"},
				},
				Pagination: &tfe.Pagination{
					NextPage: 0,
				},
			}, nil
		},
	}

	teams, err := listTeams(mockTeams, "my-org", []string{"FilterMe"})

	assert.NoError(t, err)
	assert.Len(t, teams, 1)
	assert.Equal(t, "FilterMe", teams[0].Name)
}

func TestReadTeam_Unit(t *testing.T) {
	mockTeams := &TeamsAPIMock{
		ReadFunc: func(ctx context.Context, teamID string) (*tfe.Team, error) {
			return &tfe.Team{ID: teamID, Name: "Found Team"}, nil
		},
	}

	team, err := readTeam(mockTeams, "team-123")

	assert.NoError(t, err)
	assert.Equal(t, "team-123", team.ID)
	assert.Equal(t, "Found Team", team.Name)
}

func TestGetOrgMember_Unit(t *testing.T) {
	mockOrgMem := &OrganizationMembershipsAPIMock{
		ReadFunc: func(ctx context.Context, orgMemID string) (*tfe.OrganizationMembership, error) {
			return &tfe.OrganizationMembership{
				ID: orgMemID,
				User: &tfe.User{
					ID: "user-1",
				},
				Email:  "user@example.com",
				Status: tfe.OrganizationMembershipActive,
			}, nil
		},
	}

	user, err := getOrgMember(mockOrgMem, "mem-1")

	assert.NoError(t, err)
	assert.Equal(t, "user-1", user.ID)
	assert.Equal(t, "user@example.com", user.Email)
	assert.Equal(t, tfe.OrganizationMembershipActive, user.Status)
}

func TestGenTeamDetail_Unit(t *testing.T) {
	mockOrgMem := &OrganizationMembershipsAPIMock{
		ReadFunc: func(ctx context.Context, orgMemID string) (*tfe.OrganizationMembership, error) {
			return &tfe.OrganizationMembership{
				ID:    orgMemID,
				User:  &tfe.User{ID: "user-" + orgMemID},
				Email: orgMemID + "@example.com",
			}, nil
		},
	}

	team := &tfe.Team{
		ID:        "team-1",
		Name:      "Team 1",
		UserCount: 2,
		OrganizationMemberships: []*tfe.OrganizationMembership{
			{ID: "mem-1"},
			{ID: "mem-2"},
		},
	}

	detail := genTeamDetail(mockOrgMem, team)

	assert.Equal(t, "team-1", detail.Team.ID)
	assert.Equal(t, "Team 1", detail.Team.Name)
	assert.Len(t, detail.Users, 2)
	assert.Equal(t, "user-mem-1", detail.Users[0].ID)
	assert.Equal(t, "mem-1@example.com", detail.Users[0].Email)
}

func TestTeamStruct_Unit(t *testing.T) {
	team := Team{
		ID:        "id1",
		Name:      "name1",
		UserCount: 10,
	}
	assert.Equal(t, "id1", team.ID)
	assert.Equal(t, "name1", team.Name)
	assert.Equal(t, 10, team.UserCount)
}

func TestUserStruct_Unit(t *testing.T) {
	user := User{
		ID:     "uid",
		Email:  "email@example.com",
		Status: tfe.OrganizationMembershipActive,
	}
	assert.Equal(t, "uid", user.ID)
	assert.Equal(t, "email@example.com", user.Email)
	assert.Equal(t, tfe.OrganizationMembershipActive, user.Status)
}

func TestTeamDetailStruct_Unit(t *testing.T) {
	detail := TeamDetail{
		Team: Team{ID: "tid"},
		Users: []User{
			{ID: "uid"},
		},
	}
	assert.Equal(t, "tid", detail.Team.ID)
	assert.Len(t, detail.Users, 1)
}
