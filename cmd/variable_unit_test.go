package cmd

import (
	"context"
	"testing"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/stretchr/testify/assert"
)

func TestListVariables_Unit(t *testing.T) {
	mockVars := &VariablesAPIMock{
		ListFunc: func(ctx context.Context, workspaceID string, options *tfe.VariableListOptions) (*tfe.VariableList, error) {
			return &tfe.VariableList{
				Items: []*tfe.Variable{
					{ID: "var-1", Key: "key1", Value: "val1"},
				},
				Pagination: &tfe.Pagination{NextPage: 0},
			}, nil
		},
	}

	ws := WorkspaceLite{WorkspaceID: "ws-1", WorkspaceName: "WS1"}
	vars, err := listVariables(mockVars, ws)

	assert.NoError(t, err)
	assert.Equal(t, "ws-1", vars.WorkspaceID)
	assert.Len(t, vars.Variables, 1)
	assert.Equal(t, "key1", vars.Variables[0].Key)
}

func TestReadVariable_Unit(t *testing.T) {
	mockVars := &VariablesAPIMock{
		ReadFunc: func(ctx context.Context, workspaceID string, variableID string) (*tfe.Variable, error) {
			return &tfe.Variable{ID: variableID, Key: "key1"}, nil
		},
	}

	ws := WorkspaceLite{WorkspaceID: "ws-1"}
	v, err := readVariable(mockVars, ws, "var-123")

	assert.NoError(t, err)
	assert.Equal(t, "var-123", v.Variable.ID)
}

func TestCreateVariable_Unit(t *testing.T) {
	mockVars := &VariablesAPIMock{
		CreateFunc: func(ctx context.Context, workspaceID string, options tfe.VariableCreateOptions) (*tfe.Variable, error) {
			return &tfe.Variable{ID: "var-new", Key: *options.Key}, nil
		},
	}

	key := "K"
	val := "V"
	cat := tfe.CategoryTerraform
	hcl := false
	sens := false
	v, err := createVariable(mockVars, "ws-1", &key, &val, nil, &cat, &hcl, &sens)

	assert.NoError(t, err)
	assert.Equal(t, "K", v.Key)
}

func TestUpdateVariable_Unit(t *testing.T) {
	mockVars := &VariablesAPIMock{
		UpdateFunc: func(ctx context.Context, workspaceID string, variableID string, options tfe.VariableUpdateOptions) (*tfe.Variable, error) {
			return &tfe.Variable{ID: variableID, Key: *options.Key}, nil
		},
	}

	key := "K2"
	val := "V2"
	hcl := true
	sens := true
	v, err := updateVariable(mockVars, "ws-1", "var-1", &key, &val, nil, &hcl, &sens)

	assert.NoError(t, err)
	assert.Equal(t, "K2", v.Key)
	assert.Equal(t, "var-1", v.ID)
}

func TestDeleteVariable_Unit(t *testing.T) {
	deleted := false
	mockVars := &VariablesAPIMock{
		DeleteFunc: func(ctx context.Context, workspaceID string, variableID string) error {
			deleted = true
			return nil
		},
	}

	err := deleteVariable(mockVars, "ws-1", "var-1")
	assert.NoError(t, err)
	assert.True(t, deleted)
}
