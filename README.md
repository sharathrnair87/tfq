# tfq

[![GitHub license](https://img.shields.io/github/license/sharathrnair87/tfq.svg)](https://github.com/sharathrnair87/tfq/blob/main/LICENSE)
[![GoDoc](https://godoc.org/github.com/sharathrnair87/tfq?status.svg)](https://godoc.org/github.com/sharathrnair87/tfq)
[![Go Report Card](https://goreportcard.com/badge/github.com/sharathrnair87/tfq)](https://goreportcard.com/report/github.com/sharathrnair87/tfq)
[![GitHub issues](https://img.shields.io/github/issues/sharathrnair87/tfq.svg)](https://github.com/sharathrnair87/tfq/issues)

tfq is a CLI utility to query and manage Terraform Enterprise (TFE) and Terraform Cloud (TFC), inspired by [tfe-cli](https://github.com/rgreinho/tfe-cli).

## Setup

Copy the binary (either Windows or Linux) to a path on your machine. Add the `.exe` extension if using it on Windows.

```powershell
PS> .\tfq.exe
Query TFE from the command line.

Usage:
  tfq [command]

Available Commands:
  admin             Manage TFE admin operations
  agent-pool        Query TFE/TFC Agent Pools
  completion        Generate the autocompletion script for the specified shell
  help              Help about any command
  plan              Query TFE Plans
  policy            Query TFE policies
  policy-check      Manage policy check workflows of a TFE run
  policy-set        Query TFE policy sets
  registry-module   Query/Manage TFE private module registry
  registry-provider Manage TFE private provider Registry
  run               Manage TFE runs
  tag               Query TFE tags
  team              Manage TFE teams
  team-access       Query TFE workspace team access
  variable          Manage TFE workspace variables
  workspace         Manage TFE workspaces

Flags:
  -h, --help                  help for tfq
  -l, --log string            log level (debug, info, warn, error, fatal, panic)
  -o, --organization string   terraform organization or set TFE_ORG
      --output string         Specify output format. Supported values are json or tsv (default "json")
  -q, --query string          JQ compatible query to parse JSON output
  -t, --token string          terraform token or set TFE_TOKEN
  -v, --version               version for tfq

Use "tfq [command] --help" for more information about a command.
```

## Initialization

The following environment variables can be used for configuration:

*   `TFE_ADDRESS`: TFE URL (defaults to `https://app.terraform.io/`)
*   `TFE_ORG`: TFE Organization
*   `TFE_TOKEN`: Token with read access to the organization specified in `TFE_ORG`

Additionally, `TFE_ORG` and `TFE_TOKEN` can be passed via CLI flags.

## Usage

To see available options:

```bash
tfq --help
```

### Workspace

<details>
<summary>Workspace Operations</summary>

#### List

Run with no arguments to return the following for all workspaces in the Org:

| Field | Description | Type |
| :--- | :--- | :--- |
| name | Name of the workspace | string |
| id | ID of the workspace | string |
| locked | Status of the workspace | bool |
| execution_mode | Whether the workspace runs remotely, locally or on an agent | string |
| terraform_version | Version of Terraform CLI running in the workspace | string |
| tags | List of tags against workspace | list |

Run with `--filter`, which takes a workspace name or a substring of a name to get a filtered list of workspaces:

```bash
$ tfq workspace list --filter workspace-1
[
  {
    "name": "workspace-1",
    "id": "ws-RZP914jsX1Hmc9Yo",
    "locked": false,
    "execution_mode": "remote",
    "terraform_version": "1.3.0",
    "tags": [
        "tag:1",
        "tag:2"
    ]
  }
]
```

The `--filter` flag supports filtering by workspace tags using a prefix of `tags|`:

```bash
$ tfq workspace list --filter "tags|tag:1,tag:2"
[
  {
    "name": "workspace-1",
    "id": "ws-RZP914jsX1Hmc9Yo",
    "locked": false,
    "execution_mode": "remote",
    "terraform_version": "1.3.0",
    "tags": [
        "tag:1",
        "tag:2"
    ]
  },
  {
    "name": "workspace-2",
    "id": "ws-eLcff9y8r8bRBYfj",
    "locked": false,
    "execution_mode": "remote",
    "terraform_version": "1.3.7",
    "tags": [
        "tag:1",
        "tag:2"
    ]
  }
]
```

Run with the `--detail` flag to return additional details. Note: This task can take a long time due to rate-limiting; it is recommended to use the `--filter` argument.

| Field | Description | Type |
| :--- | :--- | :--- |
| name | Name of the workspace | string |
| id | ID of the workspace | string |
| locked | Status of the workspace | bool |
| execution_mode | Whether the workspace runs remotely, locally or on an agent | string |
| terraform_version | Version of Terraform CLI running in the workspace | string |
| tags | List of tags against workspace | list |
| created_days_ago | How many days ago this workspace was created | string |
| updated_days_ago | How many days ago this workspace was updated | string |
| last_remote_run_days_ago | How many days ago was a remote run performed in this workspace | string |
| last_state_update_days_ago | How many days ago was the terraform state updated in this workspace | string |
| average_run_duration | Average duration, in seconds, of a planned-and-applied run | string |

```bash
$ tfq workspace list --filter workspace-1 --detail
[
  {
    "name": "workspace-1",
    "id": "ws-RZP914jsX1Hmc9Yo",
    "locked": false,
    "terraform_version": "1.3.0",
    "tags": [
        "tag:1",
        "tag:2"
    ],
    "created_days_ago": "819.167082",
    "updated_days_ago": "2.279692",
    "last_remote_run_days_ago": "2.281231",
    "last_state_update_days_ago": "30.174812",
    "average_run_duration": "16.334562"
  }
]
```

#### Lock/Unlock

Run with a comma-separated string of workspace IDs or a workspace name filter (mutually exclusive):

```bash
$ tfq workspace lock --ids ws-SxWNNcYPkLD48ZC7
[
  {
    "id": "ws-SxWNNcYPkLD48ZC7",
    "locked": true,
    "name": "test-workspace-1"
  }
]
```

Operations can be run against a workspace that is already locked. Optionally, the `lock` operation takes a `--reason` argument.

```bash
$ tfq workspace lock --filter dev-workspace
[
  {
    "id": "ws-5xUNCXVKrryoPcEp",
    "locked": true,
    "name": "dev-workspace"
  }
]
```

#### Lock All / Unlock All

Locks or unlocks all workspaces in the specified organization:

```bash
$ tfq workspace lockall
[
  {
    "id": "ws-SxWNNcYPkLD48ZC7",
    "locked": true,
    "name": "test-workspace-1"
  },
  {
    "id": "ws-LXkPCWnJKJ1FSgjs",
    "locked": true,
    "name": "uat-workspace"
  },
  {
    "id": "ws-E9o8VitHDAvCp3wj",
    "locked": true,
    "name": "uat-2-workspace"
  },
  {
    "id": "ws-5xUNCXVKrryoPcEp",
    "locked": true,
    "name": "dev-workspace"
  }
]
```

</details>

### Runs

<details>
<summary>Run Operations</summary>

The `run` sub-command lets you manage runs against one or more workspaces.

#### List Runs

List runs in a workspace specified by workspace ID.

*   `--status` refers to valid [Run.Status](https://developer.hashicorp.com/terraform/enterprise/api-docs/run#run-states) attributes.
*   `--operation` refers to valid [Run.Operation](https://developer.hashicorp.com/terraform/enterprise/api-docs/run#run-operations) attributes.

```bash
$ tfq run list --workspace-id ws-NMH66XMnUeF8duTx --status "policy_checked"
[
    {
        "id": "run-zQFc5h2uPhEWW9Sr",
        "status": "policy_checked",
        "workspace_id": "ws-NMH66XMnUeF8duTx",
        "workspace_name": "tfc-infra-workspace",
        "created_at": "2024-09-24T06:12:56Z",
        "run_duration": "54.476822"
    }
]
```

#### Bulk Queue

Bulk queue plans against one or many workspaces:

```bash
$ tfq run queue --filter workspace-sandbox
[
  {
    "id": "run-pX9Lrq5KCrsgCYFH",
    "workspace_id": "ws-DpeRu7KpazXEWKoJ",
    "workspace_name": "workspace-sandbox",
    "status": "pending",
    "created_at": "2024-09-24T06:12:56Z",
    "run_duration": "NA"
  }
]
```

#### Apply Runs

Apply pending plans (takes a comma-separated string of run IDs):

```bash
$ tfq run apply --ids run-UowKQd1cF7bgNfCp
[
  {
    "id": "run-UowKQd1cF7bgNfCp",
    "workspace_id": "ws-N2qoyJxF1TkfeRYy",
    "workspace_name": "test-workspace-2",
    "status": "applying",
    "created_at": "2024-09-24T06:12:56Z",
    "run_duration": "NA"
  }
]
```

#### Query Runs

Query or get run details from run IDs:

```bash
$ tfq run get --ids run-UowKQd1cF7bgNfCp
[
  {
    "id": "run-UowKQd1cF7bgNfCp",
    "workspace_id": "ws-N2qoyJxF1TkfeRYy",
    "workspace_name": "test-workspace-2",
    "status": "applied",
    "created_at": "2024-09-24T06:12:56Z",
    "run_duration": "180.452271"
  }
]
```

</details>

### Variables

<details>
<summary>Variable Operations</summary>

Perform CRUD operations on workspace variables.

#### Query/List Variables

```bash
$ tfq variable list --workspace-filter workspace-sandbox
[
  {
    "workspace_id": "ws-DpeRu7KpazXEWKoJ",
    "workspace_name": "workspace-sandbox",
    "variables": [
      {
        "id": "var-RH7Q9pyD8gtgabtz",
        "key": "WORKSPACE_VAR_1",
        "value": "",
        "description": "",
        "category": "env",
        "hcl": false,
        "sensitive": false
      },
      # ... (additional variables)
    ]
  }
]
```

#### Create New Variable

```bash
$ tfq variable create --workspace-id ws-DpeRu7KpazXEWKoJ --description "test" --key "testCLI" --value "testCLI value" --sensitive true --type terraform --hcl
{
  "id": "var-uCgZrzkPhis6qXTS",
  "key": "testCLI",
  "value": "",
  "description": "test",
  "category": "terraform",
  "hcl": true,
  "sensitive": true
}
```

#### Update Existing Variable

```bash
$ tfq variable update --variable-id var-uCgZrzkPhis6qXTS --workspace-id ws-DpeRu7KpazXEWKoJ --value "test CLI Value 2" --key "testCLI" --hcl --sensitive true
{
  "id": "var-uCgZrzkPhis6qXTS",
  "key": "testCLI",
  "value": "",
  "description": "Variable Updated by tfq",
  "category": "terraform",
  "hcl": true,
  "sensitive": true
}
```

#### Delete Existing Variable

```bash
$ tfq variable delete --variable-id var-uCgZrzkPhis6qXTS --workspace-id ws-DpeRu7KpazXEWKoJ
# Returns current variables
```

#### Create Variables from File

```bash
$ tfq variable create from-file --file variables.json --workspace-id ws-DpeRu7KpazXEWKoJ
[
  {
    "id": "var-oDNV14eJf9ijjcc2",
    "key": "test1",
    "value": "value1",
    "description": "Test Variable 1",
    "category": "env",
    "hcl": false,
    "sensitive": false
  }
]
```

</details>

### Admin

<details>
<summary>Admin Operations (TFE Only)</summary>

Perform Admin operations supported by the TFE Admin API. Admin settings are only available in Terraform Enterprise.

#### Runs

##### List Runs
Lists runs filtered on run status by querying the `admin/runs` endpoint.

```bash
$ tfq admin run list --filter "plan_queued" --query '.[] | .id'
```

##### Force-Cancel
Force cancels run IDs.

```bash
$ tfq admin run force-cancel --ids run-UFaNv3rz5XnzPhCh
```

</details>

### Plan

<details>
<summary>Query Plan Summary</summary>

Query plan summaries in TFE/TFC.

#### Show Plan

```bash
$ tfq plan show --ids plan-CRetbA3L01BsxeLq
```

#### Show Detailed Changes

Use the `--detailed-changes` flag for more information:

```bash
$ tfq plan show --ids plan-v6Li1Qvx3hbaKmGi --detailed-changes
```

</details>

### Policy

<details>
<summary>Policy Operations</summary>

Query policies in TFE/TFC.

#### List Policies

```bash
$ tfq policy list --filter "production-tagging"
```

</details>

### Tag

<details>
<summary>Tag Operations</summary>

Query Organization tag information in TFE/TFC.

#### List Tags

The `--filter` flag takes a comma-separated list of workspace IDs and returns a list of all organization tags excluding those associated with these workspaces.

```bash
$ tfq tag list --filter ws-ojAyfT3ar4oXt3eA
```

The `--search` flag returns details of the specified organization tag.

```bash
$ tfq tag list --search "tag:infrastructure"
```

</details>

### Policy Set

<details>
<summary>Policy Set Operations</summary>

Query policy sets in TFE/TFC.

#### List Policy Sets

```bash
$ tfq policy-set list
```

</details>

### Policy Check

<details>
<summary>Policy Check Operations</summary>

Examine the details of a policy check performed against a given Run ID.

#### Show Policy Check

Generates the details of a policy check performed against a Run ID.

```bash
$ tfq policy-check show --run-id run-A8PuL0GnIeldng1
```

To query only those checks which have failed:

```bash
$ tfq policy-check show --run-id run-Wxk42edRCCLB5fMi --query '.result.sentinel.data | to_entries | .[].value.policies | .[] | select(.result|not) | .policy'
```

#### Override Policy Check

Where applicable, overrides policy checks with a given Policy Check ID.

```bash
$ tfq policy-check override --policy-check-id polchk-s6moSXCk5e7dm1oR
```

</details>

### Registry Modules

<details>
<summary>Private Registry Module Operations</summary>

Query Private Modules in the Organization registry.

#### List Modules

```bash
$ tfq registry-module list --query '.[] | select(.provider == "azurerm")'
```

</details>

### Registry Providers

<details>
<summary>Private Provider Registry Operations</summary>

Query Private Providers in the Organization Registry.

#### List Providers

```bash
$ tfq registry-provider list
```

#### Get Provider Details

```bash
$ tfq registry-provider get --name aws
```

</details>

### Agent Pools

<details>
<summary>Agent Pool Management</summary>

Query Agent Pools in TFE/TFC.

#### List Agent Pools

```bash
$ tfq agent-pool list
```

</details>

### Team Access

<details>
<summary>Team Access Operations</summary>

Query TFE workspace team access.

#### List Team Accesses

Determines the workspace accesses for one or more teams. The `--team-ids` flag accepts a comma-separated string of multiple team IDs. The output shows the access *per* team (keyed by team ID) for the specified workspaces (either all or comma-separated workspace IDs).

```bash
$ tfq team-access list --team-ids "team-abc,team-xyz" --workspace-ids "ws-123,ws-456"
```

Output:
```json
[
    {
        "team-abc": [
            {
                "workspace_id": "ws-123",
                "workspace_name": "my-workspace-1",
                "attributes": {
                    "access": "read",
                    "runs": "read",
                    "variables": "read",
                    "state-versions": "read",
                    "sentinel-mocks": "none",
                    "workspace-locking": false,
                    "run-tasks": false
                }
            }
        ],
        "team-xyz": [
            {
                "workspace_id": "ws-123",
                "workspace_name": "my-workspace-1",
                "attributes": {
                    "access": "custom",
                    "runs": "apply",
                    "variables": "write",
                    "state-versions": "read",
                    "sentinel-mocks": "none",
                    "workspace-locking": true,
                    "run-tasks": false
                }
            },
            {
                "workspace_id": "ws-456",
                "workspace_name": "my-workspace-2",
                "attributes": {
                    "access": "admin",
                    "runs": "apply",
                    "variables": "write",
                    "state-versions": "write",
                    "sentinel-mocks": "read",
                    "workspace-locking": true,
                    "run-tasks": true
                }
            }
        ]
    }
]
```

</details>

## Build

GoReleaser is used to produce binaries for multiple platforms (Windows, Mac, Linux).

To build all binaries locally:

1.  Install GoReleaser: [https://goreleaser.com/install/](https://goreleaser.com/install/)
2.  Run the build command:
    ```bash
    make build
    ```

Binaries will be output to the `/dist` folder.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines on how to contribute to this project.
