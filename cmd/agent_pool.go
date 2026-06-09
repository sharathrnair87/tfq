package cmd

import (
	"context"
	"encoding/json"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/sharathrnair87/tfq/resources"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

//go:generate moq -out agent_pool_moq_test.go . AgentPoolsAPI

// AgentPoolsAPI defines the subset of tfe.AgentPools methods used by this package.
type AgentPoolsAPI interface {
	List(ctx context.Context, organization string, options *tfe.AgentPoolListOptions) (*tfe.AgentPoolList, error)
}

type AgentPool struct {
	ID                 string   `json:"id"`
	Name               string   `json:"name"`
	AgentCount         int      `json:"agent_count"`
	OrganizationScoped bool     `json:"organization_scoped"`
	Organization       string   `json:"organization"`
	Workspaces         []string `json:"workspaces"`
	AllowedWorkspaces  []string `json:"allowed_workspaces"`
}

var agentPoolCmd = &cobra.Command{
	Use:   "agent-pool",
	Short: "Query TFE/TFC Agent Pools",
	Long:  `Query TFE/TFC Agent Pools.`,
}

var agentPoolListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configured Agent Pools in the TFE/TFC Organization",
	Long:  `List all configured Agent Pools in the TFE/TFC Organization.`,
	Run: func(cmd *cobra.Command, args []string) {
		organization, client, err := resources.Setup(cmd)
		check(err)

		query, _ := cmd.Flags().GetString("query")

		agentPools, err := listAgentPools(client.AgentPools, organization)
		check(err)

		agentPoolsJson, err := json.MarshalIndent(agentPools, "", "  ")
		check(err)

		if query != "" {
			outputJsonStr, err := resources.JqRun(agentPoolsJson, query)
			check(err)
			cmd.Println(string(outputJsonStr))
		} else {
			cmd.Println(string(agentPoolsJson))
		}
	},
}

func init() {
	rootCmd.AddCommand(agentPoolCmd)
	agentPoolCmd.AddCommand(agentPoolListCmd)
}

func listAgentPools(agents AgentPoolsAPI, organization string) ([]AgentPool, error) {
	results := []AgentPool{}
	currentPage := 1

	for {
		log.Debugf("Processing page %d\n", currentPage)
		options := &tfe.AgentPoolListOptions{
			ListOptions: tfe.ListOptions{
				PageNumber: currentPage,
				PageSize:   50,
			},
		}

		aps, err := agents.List(context.Background(), organization, options)
		if err != nil {
			return nil, err
		}

		for _, apsItem := range aps.Items {
			result := AgentPool{}
			result.ID = apsItem.ID
			result.Name = apsItem.Name
			result.AgentCount = apsItem.AgentCount
			result.OrganizationScoped = apsItem.OrganizationScoped
			if apsItem.Organization != nil {
				result.Organization = apsItem.Organization.Name
			}

			for _, wk := range apsItem.Workspaces {
				result.Workspaces = append(result.Workspaces, wk.ID)
			}

			for _, wk := range apsItem.AllowedWorkspaces {
				result.AllowedWorkspaces = append(result.AllowedWorkspaces, wk.ID)
			}

			results = append(results, result)
		}

		if aps.NextPage == 0 {
			break
		}

		currentPage++
	}

	return results, nil
}
