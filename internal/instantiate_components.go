package internal

import (
	"fmt"
	"github.com/broadinstitute/terra-status-manager/internal/shared"
	"github.com/broadinstitute/terra-status-manager/pkg"
	"github.com/go-resty/resty/v2"
	"github.com/mitchellh/mapstructure"
)

func componentsToDelete(
	configComponentMap map[string]pkg.StatuspageComponent,
	remoteComponentMap map[string]shared.StatuspageResponseComponent,
) []string {

	var deleteComponentIDs []string
	for name, remoteComponent := range remoteComponentMap {
		if _, found := configComponentMap[name]; !found {
			deleteComponentIDs = append(deleteComponentIDs, remoteComponent.ID)
		}
	}
	return deleteComponentIDs
}

func componentsToCreate(
	configComponentMap map[string]pkg.StatuspageComponent,
	remoteComponentMap map[string]shared.StatuspageResponseComponent,
) ([]shared.StatuspageRequestComponent, error) {

	var createComponentReqs []shared.StatuspageRequestComponent
	for name, configComponent := range configComponentMap {
		if _, found := remoteComponentMap[name]; !found {
			var req shared.StatuspageRequestComponent
			err := mapstructure.Decode(configComponent, &req)
			if err != nil {
				return nil, fmt.Errorf("error making component creation request: %w", err)
			}
			createComponentReqs = append(createComponentReqs, req)
		}
	}
	return createComponentReqs, nil
}

func componentsToModify(
	configComponentMap map[string]pkg.StatuspageComponent,
	remoteComponentMap map[string]shared.StatuspageResponseComponent,
) (map[string]shared.StatuspageRequestComponent, error) {

	modifyComponentReqs := make(map[string]shared.StatuspageRequestComponent)
	for name, configComponent := range configComponentMap {
		if remoteComponent, found := remoteComponentMap[name]; found {
			var req shared.StatuspageRequestComponent
			err := mapstructure.Decode(remoteComponent, &req)
			if err != nil {
				return nil, fmt.Errorf("error making component modification request from remote: %w", err)
			}
			err = mapstructure.Decode(configComponent, &req)
			if err != nil {
				return nil, fmt.Errorf("error making component modification request from config: %w", err)
			}
			modifyComponentReqs[remoteComponent.ID] = req
		}
	}
	return modifyComponentReqs, nil
}

func InstantiateComponents(config *pkg.Config, client *resty.Client) error {
	response, err := client.R().
		SetResult([]shared.StatuspageResponseComponent{}).
		Get(fmt.Sprintf("/pages/%s/components", config.Statuspage.PageID))
	if err = shared.CheckResponse(response, err); err != nil {
		return err
	}

	remoteComponentList := response.Result().(*[]shared.StatuspageResponseComponent)
	remoteComponentMap := make(map[string]shared.StatuspageResponseComponent)
	for _, remoteComponent := range *remoteComponentList {
		remoteComponentMap[remoteComponent.Name] = remoteComponent
	}

	configComponentMap := make(map[string]pkg.StatuspageComponent)
	for _, configComponent := range config.Statuspage.Components {
		configComponentMap[configComponent.Name] = configComponent
	}

	toDelete := componentsToDelete(configComponentMap, remoteComponentMap)
	toCreate, err := componentsToCreate(configComponentMap, remoteComponentMap)
	if err != nil {
		return err
	}
	toModify, err := componentsToModify(configComponentMap, remoteComponentMap)
	if err != nil {
		return err
	}
	println("To delete:", len(toDelete))
	println("To modify:", len(toModify))
	println("To create:", len(toCreate))

	return nil
}
