package actions

import (
	"fmt"
	"github.com/broadinstitute/revere/internal/configuration"
	"reflect"

	"github.com/broadinstitute/revere/internal/shared"
	"github.com/broadinstitute/revere/internal/statuspage"
	"github.com/go-resty/resty/v2"
	"github.com/mitchellh/mapstructure"
)

// listComponentsToDelete provides a slice of remote components that don't correlate to an entry in the configuration
func listComponentsToDelete(
	configComponentMap map[string]configuration.Component,
	remoteComponentMap map[string]statuspage.Component,
) []statuspage.Component {
	var componentsToDelete []statuspage.Component
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
	remoteComponentMap map[string]statuspage.Component,
) ([]statuspage.Component, error) {
	var componentsToCreate []statuspage.Component
	for name, configComponent := range configComponentMap {
		if _, found := remoteComponentMap[name]; !found {
			var newComponent statuspage.Component
			statuspage.ComponentConfigToApi(configComponent, &newComponent)
			// We specifically don't want status to be influenced by configuration file; components start out operational
			newComponent.Status = "operational"
			componentsToCreate = append(componentsToCreate, newComponent)
		}
	}
	return componentsToCreate, nil
}

// listComponentsToModify provides a slice of components that should be modified on the remote
func listComponentsToModify(
	configComponentMap map[string]configuration.Component,
	remoteComponentMap map[string]statuspage.Component,
) ([]statuspage.Component, error) {
	var componentsToModify []statuspage.Component
	for name, configComponent := range configComponentMap {
		if remoteComponent, found := remoteComponentMap[name]; found {
			var modifiedComponent statuspage.Component
			err := mapstructure.Decode(remoteComponent, &modifiedComponent)
			if err != nil {
				return nil, fmt.Errorf("error decoding statuspage.Component to statuspage.Component: %w", err)
			}
			statuspage.ComponentConfigToApi(configComponent, &modifiedComponent)
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
// to the correct functions in statuspage/api.go
func ReconcileComponents(config *configuration.Config, client *resty.Client) error {
	statuspageComponents, err := statuspage.GetComponents(client, config.Statuspage.PageID)
	if err != nil {
		return err
	}
	statuspageComponentMap := make(map[string]statuspage.Component)
	for _, statuspageComponent := range *statuspageComponents {
		statuspageComponentMap[statuspageComponent.Name] = statuspageComponent
	}
	configComponentMap := make(map[string]configuration.Component)
	for _, configComponent := range config.Statuspage.Components {
		configComponentMap[configComponent.Name] = configComponent
	}

	toDelete := listComponentsToDelete(configComponentMap, statuspageComponentMap)
	toCreate, err := listComponentsToCreate(configComponentMap, statuspageComponentMap)
	if err != nil {
		return err
	}
	toModify, err := listComponentsToModify(configComponentMap, statuspageComponentMap)
	if err != nil {
		return err
	}

	for _, component := range toDelete {
		shared.LogLn(config, fmt.Sprintf("deleting %s component from statuspage", component.Name),
			fmt.Sprintf("Deleting: %+v", component))
		err := statuspage.DeleteComponent(client, config.Statuspage.PageID, component.ID)
		if err != nil {
			return err
		}
	}
	for _, component := range toCreate {
		shared.LogLn(config, fmt.Sprintf("creating %s component on statuspage", component.Name),
			fmt.Sprintf("New: %+v", component))
		_, err := statuspage.PostComponent(client, config.Statuspage.PageID, component)
		if err != nil {
			return err
		}
	}
	for _, component := range toModify {
		shared.LogLn(config, fmt.Sprintf("modifying %s component on statuspage", component.Name),
			fmt.Sprintf("Config: %+v", configComponentMap[component.Name]),
			fmt.Sprintf("Original: %+v", statuspageComponentMap[component.Name]),
			fmt.Sprintf("New: %+v", component))
		_, err := statuspage.PatchComponent(client, config.Statuspage.PageID, component.ID, component)
		if err != nil {
			return err
		}
	}

	return nil
}
