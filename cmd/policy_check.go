package cmd

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/go-tfe"
	"github.com/sharathrnair87/tfq/resources"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

//go:generate moq -out policy_check_moq_test.go . PolicyChecksAPI

// PolicyChecksAPI defines the subset of tfe.PolicyChecks methods used by this package.
type PolicyChecksAPI interface {
	List(ctx context.Context, runID string, options *tfe.PolicyCheckListOptions) (*tfe.PolicyCheckList, error)
	Override(ctx context.Context, policyCheckID string) (*tfe.PolicyCheck, error)
}

type LocalPolicyCheckResult struct {
	AdvisoryFailed int  `json:"advisory_failed"`
	HardFailed     int  `json:"hard_failed"`
	Passed         int  `json:"passed"`
	Result         bool `json:"result"`
	SoftFailed     int  `json:"soft_failed"`
	TotalFailed    int  `json:"total_failed"`
	Sentinel       any  `json:"sentinel"`
}

type PolicyCheck struct {
	ID     string                 `json:"id"`
	Result LocalPolicyCheckResult `json:"result"`
	Status tfe.PolicyStatus       `json:"status"`
	Scope  tfe.PolicyScope        `json:"scope"`
}

var policyCheckCmd = &cobra.Command{
	Use:   "policy-check",
	Short: "Manage policy check workflows of a TFE run",
	Long:  `Manage policy check workflows of a TFE run.`,
}

var policyCheckShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show details of the policy check in a TFE run",
	Long:  `Show details of the policy check in a TFE run`,
	Run: func(cmd *cobra.Command, args []string) {
		// policy check show function
		_, client, err := resources.Setup(cmd)
		check(err)

		runId, _ := cmd.Flags().GetString("run-id")

		var policyCheckJson []byte

		policyCheck, _ := showPolicyChecks(client.PolicyChecks, runId)

		policyCheckJson, _ = json.MarshalIndent(policyCheck, "", "  ")
		outputData(cmd, policyCheckJson)
	},
}

var policyCheckOverrideCmd = &cobra.Command{
	Use:   "override",
	Short: "Override the policy check for a given TFE run",
	Long:  `Override the policy  check for a given TFE run.`,
	Run: func(cmd *cobra.Command, args []string) {

		_, client, err := resources.Setup(cmd)
		check(err)

		policyCheckId, _ := cmd.Flags().GetString("policy-check-id")

		var policyCheckJson []byte

		policyCheck, _ := overridePolicyChecks(client.PolicyChecks, policyCheckId)

		policyCheckJson, _ = json.MarshalIndent(policyCheck, "", "  ")
		outputData(cmd, policyCheckJson)
	},
}

func init() {
	rootCmd.AddCommand(policyCheckCmd)

	// Show sub-command
	// Returns the detailed policy check results for a given list of RunIDs
	policyCheckCmd.AddCommand(policyCheckShowCmd)
	policyCheckShowCmd.Flags().String("run-id", "", "RunId to inspect")

	// Override sub-command
	// Overrides a given policy check
	policyCheckCmd.AddCommand(policyCheckOverrideCmd)
	policyCheckOverrideCmd.Flags().String("policy-check-id", "", "ID of the policy-check to override")
}

func showPolicyChecks(policyChecks PolicyChecksAPI, runID string) (PolicyCheck, error) {
	result := PolicyCheck{}
	log.Debugf("Retrieving policy checks for run: %s\n", runID)
	options := &tfe.PolicyCheckListOptions{}

	pc, err := policyChecks.List(context.Background(), runID, options)
	check(err)

	polchk := pc.Items[0]

	result.ID = polchk.ID
	result.Scope = polchk.Scope
	result.Status = polchk.Status
	result.Result.AdvisoryFailed = polchk.Result.AdvisoryFailed
	result.Result.HardFailed = polchk.Result.HardFailed
	result.Result.TotalFailed = polchk.Result.TotalFailed
	result.Result.SoftFailed = polchk.Result.SoftFailed
	result.Result.Passed = polchk.Result.Passed
	result.Result.Sentinel = polchk.Result.Sentinel
	result.Result.Result = polchk.Result.Result

	return result, nil
}

func overridePolicyChecks(policyChecks PolicyChecksAPI, policyCheckID string) (PolicyCheck, error) {
	result := PolicyCheck{}
	log.Debugf("Overriding policy check: %s\n", policyCheckID)

	polchk, err := policyChecks.Override(context.Background(), policyCheckID)
	check(err)

	result.ID = polchk.ID
	result.Scope = polchk.Scope
	result.Status = polchk.Status
	result.Result.AdvisoryFailed = polchk.Result.AdvisoryFailed
	result.Result.HardFailed = polchk.Result.HardFailed
	result.Result.TotalFailed = polchk.Result.TotalFailed
	result.Result.SoftFailed = polchk.Result.SoftFailed
	result.Result.Passed = polchk.Result.Passed
	result.Result.Sentinel = polchk.Result.Sentinel
	result.Result.Result = polchk.Result.Result

	return result, nil
}
