package cmd

import (
	"context"
	"encoding/json"
	"fmt"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/sharathrnair87/tfq/resources"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

//go:generate moq -out registry_provider_moq_test.go . RegistryProvidersAPI RegistryProviderVersionsAPI RegistryProviderPlatformsAPI

// RegistryProvidersAPI defines the subset of tfe.RegistryProviders methods used by this package.
type RegistryProvidersAPI interface {
	List(ctx context.Context, organization string, options *tfe.RegistryProviderListOptions) (*tfe.RegistryProviderList, error)
	Read(ctx context.Context, providerID tfe.RegistryProviderID, options *tfe.RegistryProviderReadOptions) (*tfe.RegistryProvider, error)
}

// RegistryProviderVersionsAPI defines the subset of tfe.RegistryProviderVersions methods used by this package.
type RegistryProviderVersionsAPI interface {
	List(ctx context.Context, providerID tfe.RegistryProviderID, options *tfe.RegistryProviderVersionListOptions) (*tfe.RegistryProviderVersionList, error)
}

// RegistryProviderPlatformsAPI defines the subset of tfe.RegistryProviderPlatforms methods used by this package.
type RegistryProviderPlatformsAPI interface {
	List(ctx context.Context, versionID tfe.RegistryProviderVersionID, options *tfe.RegistryProviderPlatformListOptions) (*tfe.RegistryProviderPlatformList, error)
}

type RegistryProvider struct {
	ID           string           `json:"id"`
	Name         string           `json:"name"`
	Namespace    string           `json:"namespace"`
	RegistryName tfe.RegistryName `json:"registry_name"`
}

type ProviderPlatform struct {
	ID       string `json:"id"`
	OS       string `json:"os"`
	Arch     string `json:"arch"`
	Filename string `json:"filename"`
}

type PrivateProviderDetail struct {
	RegistryProvider
	ProviderLatestVersion string             `json:"provider_latest_version"`
	ProviderPlatforms     []ProviderPlatform `json:"provider_platforms"`
}

var registryProviderCmd = &cobra.Command{
	Use:   "registry-provider",
	Short: "Manage TFE private provider Registry",
	Long:  `Manage TFE private provider Registry.`,
}

var registryProviderListCmd = &cobra.Command{
	Use:   "list",
	Short: "List private providers in a TFE Organization",
	Long:  `List private providers in a TFE Organization.`,
	Run: func(cmd *cobra.Command, args []string) {

		organization, client, err := resources.Setup(cmd)
		check(err)

		filter, _ := cmd.Flags().GetString("filter")

		providerList, err := listPrivateProviders(client.RegistryProviders, organization, filter)
		check(err)

		providerListJson, _ := json.MarshalIndent(providerList, "", "  ")

		outputData(cmd, providerListJson)
	},
}

var registryProviderGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Show details of the TFE/C private provider registry",
	Long:  `Show details of the TFE/C private provider registry`,
	Run: func(cmd *cobra.Command, args []string) {

		organization, client, err := resources.Setup(cmd)
		check(err)

		name, _ := cmd.Flags().GetString("name")

		privateProviderDetail, err := getPrivateProviderDetails(client.RegistryProviders, client.RegistryProviderVersions, client.RegistryProviderPlatforms, organization, name)
		check(err)

		privateProviderDetailJson, _ := json.MarshalIndent(privateProviderDetail, "", "  ")
		outputData(cmd, privateProviderDetailJson)
	},
}

func init() {
	rootCmd.AddCommand(registryProviderCmd)
	registryProviderCmd.AddCommand(registryProviderListCmd)
	registryProviderCmd.AddCommand(registryProviderGetCmd)

	// List sub-command
	registryProviderListCmd.Flags().String("filter", "", "Search for private provider registries by name")

	// Show sub-command
	registryProviderGetCmd.Flags().String("name", "", "Name of the private provider in the registry")
}

func listPrivateProviders(providers RegistryProvidersAPI, organization string, filter string) ([]RegistryProvider, error) {
	results := []RegistryProvider{}
	currentPage := 1

	for {
		log.Debugf("Processing page %d\n", currentPage)
		options := &tfe.RegistryProviderListOptions{
			ListOptions: tfe.ListOptions{
				PageNumber: currentPage,
				PageSize:   50,
			},
			Search:       filter,
			RegistryName: "private",
		}

		rps, err := providers.List(context.Background(), organization, options)
		if err != nil {
			return nil, err
		}

		for _, rpItem := range rps.Items {
			result := RegistryProvider{}
			result.RegistryName = rpItem.RegistryName
			result.ID = rpItem.ID
			result.Name = rpItem.Name
			result.Namespace = rpItem.Namespace

			results = append(results, result)
		}

		if rps.NextPage == 0 {
			break
		}

		currentPage++
	}

	return results, nil
}

func getPrivateProviderDetails(providers RegistryProvidersAPI, versions RegistryProviderVersionsAPI, platforms RegistryProviderPlatformsAPI, organization string, name string) (PrivateProviderDetail, error) {
	var result PrivateProviderDetail

	registryProviderList, err := listPrivateProviders(providers, organization, name)
	check(err)

	if len(registryProviderList) > 1 {
		return result, fmt.Errorf("query returns more than one Provider for name: %s", name)
	}

	registryProvider := registryProviderList[0]

	registryProviderID := tfe.RegistryProviderID{
		OrganizationName: organization,
		RegistryName:     "private",
		Namespace:        registryProvider.Namespace,
		Name:             registryProvider.Name,
	}

	pr, err := providers.Read(context.Background(), registryProviderID, &tfe.RegistryProviderReadOptions{})
	check(err)

	//Get latest provider version
	currentPage := 1
	prv, err := versions.List(context.Background(), registryProviderID, &tfe.RegistryProviderVersionListOptions{})
	check(err)

	if len(prv.Items) == 0 {
		return result, fmt.Errorf("unable to query Provider with given id: %s", registryProvider.ID)
	}

	log.Debugf("CurrentPage: %d, LastPage: %d", prv.CurrentPage, prv.TotalPages)
	lastPage := prv.TotalPages

	if currentPage != lastPage {
		prv, err = versions.List(context.Background(), registryProviderID, &tfe.RegistryProviderVersionListOptions{ListOptions: tfe.ListOptions{PageNumber: lastPage}})
		check(err)
	}

	items := prv.Items

	latestProviderVersion := items[len(items)-1]
	latestVersion := latestProviderVersion.Version

	rpv := tfe.RegistryProviderVersionID{
		RegistryProviderID: registryProviderID,
		Version:            latestVersion,
	}

	//Get provider platform details
	prpv, err := platforms.List(context.Background(), rpv, &tfe.RegistryProviderPlatformListOptions{ListOptions: tfe.ListOptions{PageSize: 100}})
	check(err)

	if len(prpv.Items) == 0 {
		return result, fmt.Errorf("unable to query Provider Platforms for Provider with given name: %s", name)
	}

	var providerPlatforms []ProviderPlatform

	result.ID = registryProvider.ID
	result.Name = pr.Name
	result.Namespace = pr.Namespace
	result.RegistryName = pr.RegistryName
	result.ProviderLatestVersion = latestVersion

	for _, platform := range prpv.Items {
		var tmpProviderPlatform ProviderPlatform

		tmpProviderPlatform.ID = platform.ID
		tmpProviderPlatform.OS = platform.OS
		tmpProviderPlatform.Arch = platform.Arch
		tmpProviderPlatform.Filename = platform.Filename

		providerPlatforms = append(providerPlatforms, tmpProviderPlatform)
	}

	result.ProviderPlatforms = providerPlatforms

	return result, nil
}
