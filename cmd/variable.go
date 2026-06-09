package cmd

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"strings"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/sharathrnair87/tfq/resources"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

//go:generate moq -out variable_moq_test.go . VariablesAPI

// VariablesAPI defines the subset of tfe.Variables methods used by this package.
type VariablesAPI interface {
	List(ctx context.Context, workspaceID string, options *tfe.VariableListOptions) (*tfe.VariableList, error)
	Read(ctx context.Context, workspaceID string, variableID string) (*tfe.Variable, error)
	Create(ctx context.Context, workspaceID string, options tfe.VariableCreateOptions) (*tfe.Variable, error)
	Update(ctx context.Context, workspaceID string, variableID string, options tfe.VariableUpdateOptions) (*tfe.Variable, error)
	Delete(ctx context.Context, workspaceID string, variableID string) error
}

type Variable struct {
	ID          string           `json:"id"`
	Key         string           `json:"key"`
	Value       string           `json:"value"`
	Description string           `json:"description"`
	Category    tfe.CategoryType `json:"category"`
	HCL         bool             `json:"hcl"`
	Sensitive   bool             `json:"sensitive"`
}

type Variables struct {
	Variables []Variable `json:"variables"`
}

type WorkspaceLite struct {
	WorkspaceID   string `json:"workspace_id"`
	WorkspaceName string `json:"workspace_name"`
}

type WorkspaceVar struct {
	WorkspaceLite
	Variable Variable `json:"variable"`
}

type WorkspaceVars struct {
	WorkspaceLite
	Variables []Variable `json:"variables"`
}

// variableCmd represents the variable command.
var variableCmd = &cobra.Command{
	Use:   "variable",
	Short: "Manage TFE workspace variables",
	Long:  `Manage TFE workspace variables.`,
}

var variableListCmd = &cobra.Command{
	Use:   "list",
	Short: "List TFE workspace variables",
	Long:  `List TFE workspace variables.`,
	Run: func(cmd *cobra.Command, args []string) {
		//list operations
		organization, client, err := resources.Setup(cmd)
		check(err)

		workspaceIds, _ := cmd.Flags().GetString("workspace-ids")
		workspaceFilter, _ := cmd.Flags().GetString("workspace-filter")

		if workspaceFilter != "" && workspaceIds != "" {
			log.Fatal("workspace-filter and workspace-ids are mutually exclusive, use one or the other!")
		}

		if workspaceFilter == "" && workspaceIds == "" {
			log.Fatal("please provide one of workspace-ids or workspace-filter to perform this operation!")
		}

		var workspaceList []WorkspaceLite
		var tmpWorkspace WorkspaceLite

		if workspaceFilter != "" {
			workspaces, err := listWorkspaces(client.Workspaces, organization, workspaceFilter)
			check(err)

			for _, workspace := range workspaces {
				tmpWorkspace.WorkspaceID = workspace.ID
				tmpWorkspace.WorkspaceName = workspace.Name

				workspaceList = append(workspaceList, tmpWorkspace)
			}
		}

		if workspaceIds != "" {
			workspaceIdList := strings.Split(workspaceIds, ",")
			for _, id := range workspaceIdList {
				workspaceName, err := getWorkspaceNameByID(client.Workspaces, organization, id)
				check(err)
				tmpWorkspace.WorkspaceID = id
				tmpWorkspace.WorkspaceName = workspaceName

				workspaceList = append(workspaceList, tmpWorkspace)
			}
		}

		var workspaceVarsListJson []byte
		var workspaceVarsList []WorkspaceVars

		for _, wrk := range workspaceList {
			w, err := listVariables(client.Variables, wrk)
			check(err)
			workspaceVarsList = append(workspaceVarsList, w)
		}

		workspaceVarsListJson, _ = json.MarshalIndent(workspaceVarsList, "", "  ")
		outputData(cmd, workspaceVarsListJson)
	},
}

var variableReadCmd = &cobra.Command{
	Use:   "read",
	Short: "Read TFE workspace variables",
	Long:  `Read TFE workspace variables.`,
	Run: func(cmd *cobra.Command, args []string) {
		//read operations
		organization, client, err := resources.Setup(cmd)
		check(err)

		workspaceID, _ := cmd.Flags().GetString("workspace-id")
		variableID, _ := cmd.Flags().GetString("variable-id")

		var tmpWorkspace WorkspaceLite

		workspaceName, err := getWorkspaceNameByID(client.Workspaces, organization, workspaceID)
		check(err)
		tmpWorkspace.WorkspaceID = workspaceID
		tmpWorkspace.WorkspaceName = workspaceName

		v, err := readVariable(client.Variables, tmpWorkspace, variableID)
		check(err)

		variableJson, _ := json.MarshalIndent(v, "", "  ")
		outputData(cmd, variableJson)
	},
}

var variableCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create TFE workspace variables",
	Long:  `Create TFE workspace variables.`,
	Run: func(cmd *cobra.Command, args []string) {
		//create operations
		_, client, err := resources.Setup(cmd)
		check(err)

		workspaceID, _ := cmd.Flags().GetString("workspace-id")

		key, _ := cmd.Flags().GetString("key")
		value, _ := cmd.Flags().GetString("value")
		description, _ := cmd.Flags().GetString("description")
		categoryTypeStr, _ := cmd.Flags().GetString("type")
		hcl, _ := cmd.Flags().GetBool("hcl")
		sensitive, _ := cmd.Flags().GetBool("sensitive")

		categoryType := tfe.CategoryType(categoryTypeStr)

		v, err := createVariable(client.Variables, workspaceID, &key, &value, &description, &categoryType, &hcl, &sensitive)
		check(err)

		variableJson, _ := json.MarshalIndent(v, "", "  ")
		outputData(cmd, variableJson)
	},
}

var variableCreateFromFileCmd = &cobra.Command{
	Use:   "from-file",
	Short: "Create variables using JSON file",
	Long:  `Create variables using JSON file`,
	Run: func(cmd *cobra.Command, args []string) {
		_, client, err := resources.Setup(cmd)
		check(err)

		file, _ := cmd.Flags().GetString("file")
		workspaceID, _ := cmd.Flags().GetString("workspace-id")

		byteVarJson := readJsonFile(file)

		var variables Variables
		var outputVariablesList []Variable
		var outputVariablesListJson []byte

		err = json.Unmarshal([]byte(byteVarJson), &variables)
		check(err)

		for _, newVar := range variables.Variables {
			v, err := createVariable(client.Variables, workspaceID, &newVar.Key, &newVar.Value, &newVar.Description, &newVar.Category, &newVar.HCL, &newVar.Sensitive)
			check(err)
			outputVariablesList = append(outputVariablesList, v)
		}
		outputVariablesListJson, _ = json.MarshalIndent(outputVariablesList, "", "  ")
		outputData(cmd, outputVariablesListJson)
	},
}

var variableUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update TFE workspace variables",
	Long:  `Update TFE workspace variables.`,
	Run: func(cmd *cobra.Command, args []string) {
		//update operations
		_, client, err := resources.Setup(cmd)
		check(err)

		workspaceID, _ := cmd.Flags().GetString("workspace-id")
		variableID, _ := cmd.Flags().GetString("variable-id")
		key, _ := cmd.Flags().GetString("key")
		value, _ := cmd.Flags().GetString("value")
		description, _ := cmd.Flags().GetString("description")
		hcl, _ := cmd.Flags().GetBool("hcl")
		sensitive, _ := cmd.Flags().GetBool("sensitive")

		v, err := updateVariable(client.Variables, workspaceID, variableID, &key, &value, &description, &hcl, &sensitive)
		check(err)

		variableJson, _ := json.MarshalIndent(v, "", "  ")
		outputData(cmd, variableJson)
	},
}

var variableUpdateFromFileCmd = &cobra.Command{
	Use:   "from-file",
	Short: "Update variables using JSON file",
	Long:  `Update variables using JSON file`,
	Run: func(cmd *cobra.Command, args []string) {
		_, client, err := resources.Setup(cmd)
		check(err)

		file, _ := cmd.Flags().GetString("file")
		workspaceID, _ := cmd.Flags().GetString("workspace-id")

		byteVarJson := readJsonFile(file)

		var variables Variables
		var outputVariablesList []Variable
		var outputVariablesListJson []byte

		err = json.Unmarshal([]byte(byteVarJson), &variables)
		check(err)

		for _, newVar := range variables.Variables {
			v, err := updateVariable(client.Variables, workspaceID, newVar.ID, &newVar.Key, &newVar.Value, &newVar.Description, &newVar.HCL, &newVar.Sensitive)
			check(err)
			outputVariablesList = append(outputVariablesList, v)
		}
		outputVariablesListJson, _ = json.MarshalIndent(outputVariablesList, "", "  ")
		outputData(cmd, outputVariablesListJson)
	},
}

var variableDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete TFE workspace variables",
	Long:  `Delete TFE workspace variables.`,
	Run: func(cmd *cobra.Command, args []string) {
		//delete operations
		organization, client, err := resources.Setup(cmd)
		check(err)

		workspaceID, _ := cmd.Flags().GetString("workspace-id")
		variableID, _ := cmd.Flags().GetString("variable-id")

		err = deleteVariable(client.Variables, workspaceID, variableID)
		check(err)

		var workspaceVarsListJson []byte
		var workspaceVarsList []WorkspaceVars
		var tmpWorkspace WorkspaceLite

		workspaceName, err := getWorkspaceNameByID(client.Workspaces, organization, workspaceID)
		check(err)
		tmpWorkspace.WorkspaceID = workspaceID
		tmpWorkspace.WorkspaceName = workspaceName

		w, err := listVariables(client.Variables, tmpWorkspace)
		check(err)

		workspaceVarsList = append(workspaceVarsList, w)
		workspaceVarsListJson, _ = json.MarshalIndent(workspaceVarsList, "", "  ")
		outputData(cmd, workspaceVarsListJson)
	},
}

func init() {
	rootCmd.AddCommand(variableCmd)

	// List sub-command
	variableCmd.AddCommand(variableListCmd)
	variableListCmd.Flags().String("workspace-ids", "", "Comma separated list of workspaceIDs")
	variableListCmd.Flags().String("workspace-filter", "", "Search filter for workspace")

	// Read sub-command
	variableCmd.AddCommand(variableReadCmd)
	variableReadCmd.Flags().String("workspace-id", "", "workspaceID")
	variableReadCmd.Flags().String("variable-id", "", "variableID of the variable")

	// Create sub-command
	variableCmd.AddCommand(variableCreateCmd)
	variableCreateCmd.Flags().String("workspace-id", "", "workspaceID")
	variableCreateCmd.Flags().String("key", "", "Variable Name")
	variableCreateCmd.Flags().String("value", "", "Variable Value")
	variableCreateCmd.Flags().Bool("sensitive", false, "Set sensitive flag for variable")
	variableCreateCmd.Flags().Bool("hcl", false, "Set if variable has HCL syntax")
	variableCreateCmd.Flags().String("type", "env", "Variable type")
	variableCreateCmd.Flags().String("description", "Variable Created by tfq", "Description for the variable")
	// Create from file sub-command
	variableCreateCmd.AddCommand(variableCreateFromFileCmd)
	variableCreateFromFileCmd.Flags().String("file", "", "File containing workspace variables")
	variableCreateFromFileCmd.Flags().String("workspace-id", "", "workspaceID")

	// Update sub-command
	variableCmd.AddCommand(variableUpdateCmd)
	variableUpdateCmd.Flags().String("workspace-id", "", "workspaceID")
	variableUpdateCmd.Flags().String("variable-id", "", "variableID")
	variableUpdateCmd.Flags().String("key", "", "Variable Name")
	variableUpdateCmd.Flags().String("value", "", "Variable Value")
	variableUpdateCmd.Flags().Bool("sensitive", false, "Set sensitive flag for variable")
	variableUpdateCmd.Flags().Bool("hcl", false, "Set if variable has HCL syntax")
	variableUpdateCmd.Flags().String("description", "Variable Updated by tfq", "Description for the variable")
	// Update from file sub-command
	variableUpdateCmd.AddCommand(variableUpdateFromFileCmd)
	variableUpdateFromFileCmd.Flags().String("file", "", "File containing workspace variables")
	variableUpdateFromFileCmd.Flags().String("workspace-id", "", "workspaceID")

	// Delete sub-command
	variableCmd.AddCommand(variableDeleteCmd)
	variableDeleteCmd.Flags().String("workspace-id", "", "workspaceID")
	variableDeleteCmd.Flags().String("variable-id", "", "variableID of the variable")

}

func listVariables(variables VariablesAPI, workspace WorkspaceLite) (WorkspaceVars, error) {
	result := WorkspaceVars{}
	currentPage := 1

	for {
		log.Debugf("Processing page %d.\n", currentPage)
		options := &tfe.VariableListOptions{
			ListOptions: tfe.ListOptions{
				PageNumber: currentPage,
				PageSize:   50,
			},
		}

		varList, err := variables.List(context.Background(), workspace.WorkspaceID, options)
		check(err)

		var tmpVarList []Variable
		for _, v := range varList.Items {
			var tmpVar = Variable{
				ID:          v.ID,
				Key:         v.Key,
				Value:       v.Value,
				Description: v.Description,
				Category:    v.Category,
				HCL:         v.HCL,
				Sensitive:   v.Sensitive,
			}

			tmpVarList = append(tmpVarList, tmpVar)
		}

		result = WorkspaceVars{
			WorkspaceLite: workspace,
			Variables:     tmpVarList,
		}

		if varList.NextPage == 0 {
			break
		}

		currentPage++
	}

	return result, nil
}

func readVariable(variables VariablesAPI, workspace WorkspaceLite, variableID string) (WorkspaceVar, error) {
	result := WorkspaceVar{}

	v, err := variables.Read(context.Background(), workspace.WorkspaceID, variableID)
	check(err)

	result.WorkspaceID = workspace.WorkspaceID
	result.WorkspaceName = workspace.WorkspaceName
	result.Variable.ID = v.ID
	result.Variable.Key = v.Key
	result.Variable.Value = v.Value
	result.Variable.Description = v.Description
	result.Variable.Category = v.Category
	result.Variable.HCL = v.HCL
	result.Variable.Sensitive = v.Sensitive

	return result, nil
}

func createVariable(variables VariablesAPI, workspaceID string, key *string, value *string, description *string, category *tfe.CategoryType, hcl *bool, sensitive *bool) (Variable, error) {
	var result Variable

	options := tfe.VariableCreateOptions{
		Key:         key,
		Value:       value,
		Description: description,
		Category:    category,
		HCL:         hcl,
		Sensitive:   sensitive,
	}

	v, err := variables.Create(context.Background(), workspaceID, options)
	check(err)

	result = Variable{
		ID:          v.ID,
		Key:         v.Key,
		Value:       v.Value,
		Description: v.Description,
		Category:    v.Category,
		HCL:         v.HCL,
		Sensitive:   v.Sensitive,
	}

	return result, nil
}

func updateVariable(variables VariablesAPI, workspaceID string, variableID string, key *string, value *string, description *string, hcl *bool, sensitive *bool) (Variable, error) {
	var result Variable

	options := tfe.VariableUpdateOptions{
		Key:         key,
		Value:       value,
		Description: description,
		HCL:         hcl,
		Sensitive:   sensitive,
	}

	v, err := variables.Update(context.Background(), workspaceID, variableID, options)
	check(err)

	result = Variable{
		ID:          v.ID,
		Key:         v.Key,
		Value:       v.Value,
		Description: v.Description,
		Category:    v.Category,
		HCL:         v.HCL,
		Sensitive:   v.Sensitive,
	}

	return result, nil
}

func deleteVariable(variables VariablesAPI, workspaceID string, variableID string) error {

	err := variables.Delete(context.Background(), workspaceID, variableID)

	return err
}

func readJsonFile(file string) []byte {
	jsonFile, err := os.Open(file)
	check(err)

	defer jsonFile.Close()

	byteJson, err := io.ReadAll(jsonFile)
	check(err)

	return []byte(byteJson)
}
