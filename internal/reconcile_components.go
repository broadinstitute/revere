package internal

import (
	"fmt"
	"github.com/broadinstitute/terra-status-manager/internal/shared"
	"github.com/broadinstitute/terra-status-manager/internal/statuspage"
	"github.com/broadinstitute/terra-status-manager/pkg"
	"github.com/go-resty/resty/v2"
	"github.com/mitchellh/mapstructure"
	"reflect"
)

// componentsToDelete provides a slice of remote components that don't correlate to an entry in the config
func componentsToDelete(
	configComponentMap map[string]pkg.Component,
	remoteComponentMap map[string]statuspage.Component,
) []statuspage.Component {
	var ret []statuspage.Component
	for name, remoteComponent := range remoteComponentMap {
		if _, found := configComponentMap[name]; !found {
			ret = append(ret, remoteComponent)
		}
	}
	return ret
}

// componentsToCreate provides a slice of components that aren't present on the remote
func componentsToCreate(
	configComponentMap map[string]pkg.Component,
	remoteComponentMap map[string]statuspage.Component,
) ([]statuspage.Component, error) {
	var ret []statuspage.Component
	for name, configComponent := range configComponentMap {
		if _, found := remoteComponentMap[name]; !found {
			var c statuspage.Component
			err := statuspage.ComponentConfigToApi(configComponent, &c)
			if err != nil {
				return nil, fmt.Errorf("error decoding pkg.Component to statuspage.Component: %w", err)
			}
			// We specifically don't want status to be influenced by config file; components start out operational
			c.Status = "operational"
			ret = append(ret, c)
		}
	}
	return ret, nil
}

// componentsToModify provides a slice of components that should be modified on the remote
func componentsToModify(
	configComponentMap map[string]pkg.Component,
	remoteComponentMap map[string]statuspage.Component,
) ([]statuspage.Component, error) {
	var ret []statuspage.Component
	for name, configComponent := range configComponentMap {
		if remoteComponent, found := remoteComponentMap[name]; found {
			var c statuspage.Component
			err := mapstructure.Decode(remoteComponent, &c)
			if err != nil {
				return nil, fmt.Errorf("error decoding statuspage.Component to statuspage.Component: %w", err)
			}
			err = statuspage.ComponentConfigToApi(configComponent, &c)
			if err != nil {
				return nil, fmt.Errorf("error decoding pkg.Component to statuspage.Component: %w", err)
			}
			// if remote component is different from remote+config component, it must be modified
			if !reflect.DeepEqual(remoteComponent, c) {
				ret = append(ret, c)
			}
		}
	}
	return ret, nil
}

func ReconcileComponents(config *pkg.Config, client *resty.Client) error {
	statuspageComponents, err := statuspage.GetComponents(client, config.Statuspage.PageID)
	if err != nil {
		return err
	}
	statuspageComponentMap := make(map[string]statuspage.Component)
	for _, statuspageComponent := range *statuspageComponents {
		statuspageComponentMap[statuspageComponent.Name] = statuspageComponent
	}
	configComponentMap := make(map[string]pkg.Component)
	for _, configComponent := range config.Statuspage.Components {
		configComponentMap[configComponent.Name] = configComponent
	}

	toDelete := componentsToDelete(configComponentMap, statuspageComponentMap)
	toCreate, err := componentsToCreate(configComponentMap, statuspageComponentMap)
	if err != nil {
		return err
	}
	toModify, err := componentsToModify(configComponentMap, statuspageComponentMap)
	if err != nil {
		return err
	}

	for _, c := range toDelete {
		shared.LogLn(config, fmt.Sprintf("deleting %s component from statuspage", c.Name),
			fmt.Sprintf("Deleting: %+v", c))
		err := statuspage.DeleteComponent(client, config.Statuspage.PageID, c.ID)
		if err != nil {
			return err
		}
	}
	for _, c := range toCreate {
		shared.LogLn(config, fmt.Sprintf("creating %s component on statuspage", c.Name),
			fmt.Sprintf("New: %+v", c))
		_, err := statuspage.PostComponent(client, config.Statuspage.PageID, c)
		if err != nil {
			return err
		}
	}
	for _, c := range toModify {
		shared.LogLn(config, fmt.Sprintf("modifying %s component on statuspage", c.Name),
			fmt.Sprintf("Config: %+v", configComponentMap[c.Name]),
			fmt.Sprintf("Original: %+v", statuspageComponentMap[c.Name]),
			fmt.Sprintf("New: %+v", c))
		_, err := statuspage.PatchComponent(client, config.Statuspage.PageID, c.ID, c)
		if err != nil {
			return err
		}
	}

	return nil
}
