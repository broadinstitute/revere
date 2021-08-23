package statuspage

import (
	"fmt"
	"github.com/broadinstitute/revere/internal/configuration"
	"github.com/broadinstitute/revere/internal/shared"
	"github.com/broadinstitute/revere/internal/statuspage/statuspageapi"
	"github.com/broadinstitute/revere/internal/statuspage/statuspagetypes"
	"github.com/go-resty/resty/v2"
	"github.com/mitchellh/mapstructure"
	"reflect"
)

// listComponentsToDelete provides a slice of remote components that don't correlate to an entry in the configuration
func listComponentsToDelete(
	configComponentMap map[string]configuration.Component,
	remoteComponentMap map[string]statuspagetypes.Component,
) []statuspagetypes.Component {
	var componentsToDelete []statuspagetypes.Component
	for name, remoteComponent := range remoteComponentMap {
		if _, found := configComponentMap[name]; !found {
			componentsToDelete = append(componentsToDelete, remoteComponent)
		}
	}
	return componentsToDelete
}

// listComponentsToCreate provides a slice of components that aren't present on the remote
func listComponentsToCreate(
	configComponentMap map[string]configuration.Component,
	remoteComponentMap map[string]statuspagetypes.Component,
) []statuspagetypes.Component {
	var componentsToCreate []statuspagetypes.Component
	for name, configComponent := range configComponentMap {
		if _, found := remoteComponentMap[name]; !found {
			var newComponent statuspagetypes.Component
			statuspagetypes.MergeConfigComponentToApi(configComponent, &newComponent)
			// We specifically don't want status to be influenced by configuration file; components start out operational
			newComponent.Status = "operational"
			componentsToCreate = append(componentsToCreate, newComponent)
		}
	}
	return componentsToCreate
}

// listComponentsToModify provides a slice of components that should be modified on the remote
func listComponentsToModify(
	configComponentMap map[string]configuration.Component,
	remoteComponentMap map[string]statuspagetypes.Component,
) ([]statuspagetypes.Component, error) {
	var componentsToModify []statuspagetypes.Component
	for name, configComponent := range configComponentMap {
		if remoteComponent, found := remoteComponentMap[name]; found {
			var modifiedComponent statuspagetypes.Component
			err := mapstructure.Decode(remoteComponent, &modifiedComponent)
			if err != nil {
				return nil, fmt.Errorf("error decoding statuspage.Component to statuspage.Component: %w", err)
			}
			statuspagetypes.MergeConfigComponentToApi(configComponent, &modifiedComponent)
			// if remote component is different from remote+configuration component, it must be modified
			if !reflect.DeepEqual(remoteComponent, modifiedComponent) {
				componentsToModify = append(componentsToModify, modifiedComponent)
			}
		}
	}
	return componentsToModify, nil
}

// ReconcileComponents modifies the set of components on Statuspage.io to match what is given in the config file.
// It creates statuspage.Component slices for deletion, creation, and modification, and then hands that data
// to the correct functions in statuspage/component_api.go
func ReconcileComponents(config *configuration.Config, client *resty.Client) error {
	statuspageComponents, err := statuspageapi.GetComponents(client, config.Statuspage.PageID)
	if err != nil {
		return err
	}
	statuspageComponentMap := make(map[string]statuspagetypes.Component)
	for _, statuspageComponent := range *statuspageComponents {
		statuspageComponentMap[statuspageComponent.Name] = statuspageComponent
	}
	configComponentMap := make(map[string]configuration.Component)
	for _, configComponent := range config.Statuspage.Components {
		configComponentMap[configComponent.Name] = configComponent
	}

	toDelete := listComponentsToDelete(configComponentMap, statuspageComponentMap)
	toCreate := listComponentsToCreate(configComponentMap, statuspageComponentMap)
	toModify, err := listComponentsToModify(configComponentMap, statuspageComponentMap)
	if err != nil {
		return err
	}

	for _, component := range toDelete {
		shared.LogLn(config, fmt.Sprintf("deleting %s component from statuspage", component.Name),
			fmt.Sprintf(" - deleting: %+v", component))
		err := statuspageapi.DeleteComponent(client, config.Statuspage.PageID, component.ID)
		if err != nil {
			return err
		}
	}
	for _, component := range toCreate {
		shared.LogLn(config, fmt.Sprintf("creating %s component on statuspage", component.Name),
			fmt.Sprintf(" - new: %+v", component))
		_, err := statuspageapi.PostComponent(client, config.Statuspage.PageID, component)
		if err != nil {
			return err
		}
	}
	for _, component := range toModify {
		shared.LogLn(config, fmt.Sprintf("modifying %s component on statuspage", component.Name),
			fmt.Sprintf(" - config: %+v", configComponentMap[component.Name]),
			fmt.Sprintf(" - remote: %+v", statuspageComponentMap[component.Name]),
			fmt.Sprintf(" - modified: %+v", component))
		_, err := statuspageapi.PatchComponent(client, config.Statuspage.PageID, component.ID, component)
		if err != nil {
			return err
		}
	}

	return nil
}
