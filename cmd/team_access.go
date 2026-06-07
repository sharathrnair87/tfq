package cmd

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"time"

	"github.com/sharathrnair87/tfq/resources"
	tfe "github.com/hashicorp/go-tfe"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type TeamAccess struct {
	WorkspaceID   string               `json:"workspace_id"`
	WorkspaceName string               `json:"workspace_name"`
	Attributes    TeamAccessAttributes `json:"attributes"`
}

type TeamAccessAttributes struct {
	Access           string `json:"access"`
	Runs             string `json:"runs"`
	Variables        string `json:"variables"`
	StateVersions    string `json:"state-versions"`
	SentinelMocks    string `json:"sentinel-mocks"`
	WorkspaceLocking bool   `json:"workspace-locking"`
	RunTasks         bool   `json:"run-tasks"`
}

type TeamAccessAPI interface {
  List(ctx context.Context, options *tfe.TeamAccessListOptions) (*tfe.TeamAccessList, error)
}

var teamAccessCmd = &cobra.Command{
	Use:   "team-access",
	Short: "Query TFE workspace team access",
	Long:  `Query TFE workspace team access.`,
}

var teamAccessListCmd = &cobra.Command{
	Use:   "list",
	Short: "List TFE workspace team access",
	Long:  `List TFE workspace team access`,
	Run: func(cmd *cobra.Command, args []string) {
		organization, client, err := resources.Setup(cmd)
		check(err)

		teamIDs, _ := cmd.Flags().GetString("team-ids")

		if teamIDs == "" {
			log.Fatal("team-ids is required")
		}
		idList := strings.Split(teamIDs, ",")

		// List all workspaces to find where these teams have access
		workspaces, err := listWorkspaces(client.Workspaces, organization, "")
		check(err)

		var teamAccessResults []TeamAccess
		wg := sync.WaitGroup{}
		ch := make(chan *TeamAccess, len(workspaces)*len(idList))

		// Ratelimit
		chunkSize := 3
		if len(workspaces) < 3 {
			chunkSize = len(workspaces)
		}

		for i := 0; i < len(workspaces); i += chunkSize {
			if i+chunkSize > len(workspaces) {
				chunkSize = len(workspaces) - i
			}
			workspacesChunk := workspaces[i : i+chunkSize]
			for _, ws := range workspacesChunk {
				wg.Add(1)
				go func(workspaceID string) {
					defer wg.Done()
					for _, teamID := range idList {
						access, err := getWorkspaceTeamAccess(client.TeamAccess, client.Workspaces, workspaceID, teamID, organization)
						if err != nil {
							log.Debugf("Error getting access for workspace %s, team %s: %v", workspaceID, teamID, err)
							continue
						}
						if access != nil {
							ch <- access
						}
					}
				}(ws.ID)
			}
			wg.Wait()
			time.Sleep(500 * time.Millisecond)
		}
		close(ch)

		for access := range ch {
			teamAccessResults = append(teamAccessResults, *access)
		}

		teamAccessJson, _ := json.MarshalIndent(teamAccessResults, "", "  ")
		outputData(cmd, teamAccessJson)
	},
}

func init() {
	rootCmd.AddCommand(teamAccessCmd)
	teamAccessCmd.AddCommand(teamAccessListCmd)
	teamAccessListCmd.Flags().String("team-ids", "", "Comma separated list of team IDs to list workspace access for")
}

func getWorkspaceTeamAccess(teamAccess TeamAccessAPI, workspaces WorkspacesAPI, workspaceID string, teamID string, organization string) (*TeamAccess, error) {
	currentPage := 1
	for {
		options := &tfe.TeamAccessListOptions{
			ListOptions: tfe.ListOptions{
				PageNumber: currentPage,
				PageSize:   50,
			},
			WorkspaceID: workspaceID,
		}

		//ta, err := client.TeamAccess.List(context.Background(), options)
    ta, err := teamAccess.List(context.Background(), options)
		if err != nil {
			return nil, err
		}

		workspaceName, err := getWorkspaceNameByID(workspaces, organization, workspaceID)
		check(err)

		for _, item := range ta.Items {
			if item.Team != nil && item.Team.ID == teamID {
				return &TeamAccess{
					WorkspaceID:   workspaceID,
					WorkspaceName: workspaceName,
					Attributes: TeamAccessAttributes{
						Access:           string(item.Access),
						Runs:             string(item.Runs),
						Variables:        string(item.Variables),
						StateVersions:    string(item.StateVersions),
						SentinelMocks:    string(item.SentinelMocks),
						WorkspaceLocking: item.WorkspaceLocking,
						RunTasks:         item.RunTasks,
					},
				}, nil
			}
		}

		if ta.NextPage == 0 {
			break
		}
		currentPage++
	}

	return nil, nil
}
