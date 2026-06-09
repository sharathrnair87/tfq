package cmd

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"time"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/sharathrnair87/tfq/resources"

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
		workspaceIDs, _ := cmd.Flags().GetString("workspace-ids")

		if teamIDs == "" {
			log.Fatal("team-id is required")
		}

		teamIDList := strings.Split(teamIDs, ",")

		var workspaces []*tfe.Workspace

		if workspaceIDs == "all" {
			workspaces, err = listWorkspaces(client.Workspaces, organization, "")
			check(err)
		} else {
			workspaceIDList := strings.Split(workspaceIDs, ",")
			for _, workspace := range workspaceIDList {
				workspaces = append(workspaces, &tfe.Workspace{
					ID: workspace,
				})
			}
		}

		teamAccessResults := make([]map[string][]TeamAccess, 0)
		for _, teamID := range teamIDList {
			teamAccessResults = append(teamAccessResults, map[string][]TeamAccess{
				teamID: {},
			})
		}
		wg := sync.WaitGroup{}
		type workerResult struct {
			workspaceID string
			accesses    map[string]*TeamAccess
			err         error
		}
		ch := make(chan workerResult, len(workspaces))

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
					accesses, err := getWorkspaceTeamAccess(client.TeamAccess, client.Workspaces, workspaceID, teamIDList, organization)
					if err != nil {
						log.Debugf("Error getting access for workspace %s: %v", workspaceID, err)
					}
					ch <- workerResult{workspaceID: workspaceID, accesses: accesses, err: err}
				}(ws.ID)
			}
			wg.Wait()
			time.Sleep(500 * time.Millisecond)
		}
		close(ch)

		for res := range ch {
			if res.err != nil || res.accesses == nil {
				continue
			}
			for teamID, access := range res.accesses {
				if access != nil {
					teamAccessResults[0][teamID] = append(teamAccessResults[0][teamID], *access)
				}
			}
		}

		teamAccessJson, _ := json.MarshalIndent(teamAccessResults, "", "  ")
		outputData(cmd, teamAccessJson)
	},
}

func init() {
	rootCmd.AddCommand(teamAccessCmd)
	teamAccessCmd.AddCommand(teamAccessListCmd)
	teamAccessListCmd.Flags().String("team-ids", "", "Comma separated list of IDs of the teams whose access is to be determined")
	teamAccessListCmd.Flags().String("workspace-ids", "all", "Comma separated list of workspace IDs")
}

func getWorkspaceTeamAccess(teamAccess TeamAccessAPI, workspaces WorkspacesAPI, workspaceID string, teamIDs []string, organization string) (map[string]*TeamAccess, error) {
	results := make(map[string]*TeamAccess)
	targetTeams := make(map[string]bool)
	for _, id := range teamIDs {
		targetTeams[id] = true
	}
	currentPage := 1
	for {
		options := &tfe.TeamAccessListOptions{
			ListOptions: tfe.ListOptions{
				PageNumber: currentPage,
				PageSize:   50,
			},
			WorkspaceID: workspaceID,
		}

		ta, err := teamAccess.List(context.Background(), options)
		if err != nil {
			return nil, err
		}

		workspaceName, err := getWorkspaceNameByID(workspaces, organization, workspaceID)
		check(err)

		for _, item := range ta.Items {
			if item.Team != nil && targetTeams[item.Team.ID] {
				results[item.Team.ID] = &TeamAccess{
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
				}
			}
		}

		if ta.NextPage == 0 {
			break
		}
		currentPage++
	}

	return results, nil
}
