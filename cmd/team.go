package cmd

import (
	"context"
	"encoding/json"
	"strings"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/sharathrnair87/tfq/resources"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

//go:generate moq -out team_moq_test.go . TeamsAPI OrganizationMembershipsAPI

// TeamsAPI defines the subset of tfe.Teams methods used by this package.
type TeamsAPI interface {
	List(ctx context.Context, organization string, options *tfe.TeamListOptions) (*tfe.TeamList, error)
	Read(ctx context.Context, teamID string) (*tfe.Team, error)
}

// OrganizationMembershipsAPI defines the subset of tfe.OrganizationMemberships methods used by this package.
type OrganizationMembershipsAPI interface {
	Read(ctx context.Context, orgMemID string) (*tfe.OrganizationMembership, error)
}

type User struct {
	ID     string                           `json:"user_id"`
	Email  string                           `json:"email"`
	Status tfe.OrganizationMembershipStatus `json:"status"`
}

type TeamDetail struct {
	Team  Team   `json:"team"`
	Users []User `json:"user_list"`
}

type Team struct {
	Name      string `json:"name"`
	ID        string `json:"id"`
	UserCount int    `json:"user_count"`
}

var teamCmd = &cobra.Command{
	Use:   "team",
	Short: "Manage TFE teams",
	Long:  `Manage TFE teams.`,
}

var teamListCmd = &cobra.Command{
	Use:   "list",
	Short: "List TFE teams",
	Long:  `List TFE teams.`,
	Run: func(cmd *cobra.Command, args []string) {
		// setup
		organization, client, err := resources.Setup(cmd)
		check(err)

		// List teams.
		teams, err := listTeams(client.Teams, organization, []string{})
		check(err)

		var teamJson []byte
		var teamList []Team

		for _, team := range teams {
			var tmpTeam Team

			tmpTeam.ID = team.ID
			tmpTeam.Name = team.Name
			tmpTeam.UserCount = team.UserCount

			log.Debugf("Adding team %v", tmpTeam)
			teamList = append(teamList, tmpTeam)
		}

		teamJson, _ = json.MarshalIndent(teamList, "", "  ")

		outputData(cmd, teamJson)
	},
}

var teamGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get TFE team details",
	Long:  `Get TFE team details.`,
	Run: func(cmd *cobra.Command, args []string) {
		// setup
		organization, client, err := resources.Setup(cmd)
		check(err)

		ids, _ := cmd.Flags().GetString("ids")
		names, _ := cmd.Flags().GetString("names")

		if names != "" && ids != "" {
			log.Fatal("names and ids are mutually exclusive, use one or the other!")
		}

		if names == "" && ids == "" {
			log.Fatal("please provide one of ids or names to perform this operation!")
		}

		var teamJson []byte
		var teamList []TeamDetail

		if ids != "" {
			idList := strings.Split(ids, ",")

			for _, id := range idList {
				var tmpTeam TeamDetail

				team, _ := readTeam(client.Teams, id)
				tmpTeam = genTeamDetail(client.OrganizationMemberships, team)

				log.Debugf("Adding team %v", tmpTeam)
				teamList = append(teamList, tmpTeam)
			}
		}

		if names != "" {
			namesList := strings.Split(names, ",")
			teams, err := listTeams(client.Teams, organization, namesList)
			check(err)

			for _, team := range teams {
				tmpTeam := genTeamDetail(client.OrganizationMemberships, team)

				log.Debugf("Adding team %v", tmpTeam)
				teamList = append(teamList, tmpTeam)
			}
		}

		teamJson, _ = json.MarshalIndent(teamList, "", "  ")

		outputData(cmd, teamJson)
	},
}

func init() {
	rootCmd.AddCommand(teamCmd)

	// List sub-command
	teamCmd.AddCommand(teamListCmd)
	teamListCmd.Flags().Bool("detail", false, "Provide team membership details: userID, email and status")

	// Get sub-command
	teamCmd.AddCommand(teamGetCmd)
	// Begin mutually-exclusive flags
	teamGetCmd.Flags().String("ids", "", "Comma separated string of team ids")
	teamGetCmd.Flags().String("names", "", "comma separated string of Team names to filter")
	// End mutually-exclusive flags
}

func listTeams(teams TeamsAPI, organization string, filters []string) ([]*tfe.Team, error) {
	results := []*tfe.Team{}
	currentPage := 1

	// Go through the pages of results until there is no more pages.
	for {
		log.Debugf("Processing page %d.\n", currentPage)

		options := &tfe.TeamListOptions{
			ListOptions: tfe.ListOptions{
				PageNumber: currentPage,
				PageSize:   50,
			},
		}

		if len(filters) != 0 {
			options.Names = filters
		}

		log.Debugf("options: %v", options)

		t, err := teams.List(context.Background(), organization, options)
		check(err)

		log.Debugf("%v", t.TotalPages)
		log.Debugf("%v", t.NextPage)

		results = append(results, t.Items...)

		// Check if there is another page to retrieve.
		if t.NextPage == 0 {
			break
		}

		// Increment the page number.
		currentPage++
	}

	return results, nil
}

func getOrgMember(orgMem OrganizationMembershipsAPI, orgMemID string) (User, error) {
	result := User{}

	o, err := orgMem.Read(context.Background(), orgMemID)
	check(err)

	result.ID = o.User.ID
	result.Email = o.Email
	result.Status = o.Status

	return result, nil
}

func readTeam(teams TeamsAPI, teamID string) (*tfe.Team, error) {
	result, err := teams.Read(context.Background(), teamID)
	check(err)

	return result, nil
}

func genTeamDetail(orgMem OrganizationMembershipsAPI, team *tfe.Team) TeamDetail {
	result := TeamDetail{}

	result.Team.ID = team.ID
	result.Team.Name = team.Name
	result.Team.UserCount = team.UserCount

	for _, orgMemItem := range team.OrganizationMemberships {
		var tmpUser User
		log.Debugf("Org mem: %v", &orgMemItem)
		tmpUser, _ = getOrgMember(orgMem, orgMemItem.ID)

		log.Debugf("Adding User %v", tmpUser)
		result.Users = append(result.Users, tmpUser)
	}

	return result
}
