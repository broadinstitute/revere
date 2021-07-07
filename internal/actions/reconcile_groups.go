package actions

import (
	"fmt"
	"github.com/broadinstitute/revere/internal/configuration"
	"github.com/broadinstitute/revere/internal/shared"
	"github.com/broadinstitute/revere/internal/statuspage"
	"github.com/broadinstitute/revere/internal/statuspage/statuspagetypes"
	"github.com/go-resty/resty/v2"
	"github.com/mitchellh/mapstructure"
	"reflect"
	"sort"
)

// makeComponentMapping creates an name-ID mapping based on the server's list of components.
// If multiple IDs exist for the same name, the value for the name will be the "greatest" ID.
// This behavior exists just to stabilize handling of the "garbage in, garbage out" multiple-
// components-with-same-name case.
func makeComponentMapping(client *resty.Client, pageID string) (map[string]string, error) {
	statuspageComponents, err := statuspage.GetComponents(client, pageID)
	if err != nil {
		return nil, err
	}
	componentNameToID := make(map[string]string)
	for _, remoteComponent := range *statuspageComponents {
		if existingID, found := componentNameToID[remoteComponent.Name]; !found || existingID < remoteComponent.ID {
			componentNameToID[remoteComponent.Name] = remoteComponent.ID
		}
	}
	return componentNameToID, nil
}

// makeStatuspageGroupMapping creates a name-group mapping based on the server's list of groups.
// If multiple groups exist for the same name, the value for the name will be the "greatest" ID.
// This behavior exists just to stabilize handling of the "garbage in, garbage out" multiple-
// groups-with-same-name case.
func makeStatuspageGroupMapping(client *resty.Client, pageID string) (map[string]statuspagetypes.Group, error) {
	statuspageGroups, err := statuspage.GetGroups(client, pageID)
	if err != nil {
		return nil, err
	}
	statuspageGroupMap := make(map[string]statuspagetypes.Group)
	for _, statuspageGroup := range *statuspageGroups {
		if existing, found := statuspageGroupMap[statuspageGroup.Name]; !found || existing.ID < statuspageGroup.ID {
			statuspageGroupMap[statuspageGroup.Name] = statuspageGroup
		}
	}
	return statuspageGroupMap, nil
}

func listGroupsToDelete(
	configGroupMap map[string]configuration.ComponentGroup,
	remoteGroupMap map[string]statuspagetypes.Group,
) []statuspagetypes.Group {
	var groupsToDelete []statuspagetypes.Group
	for name, remoteGroup := range remoteGroupMap {
		if _, found := configGroupMap[name]; !found {
			groupsToDelete = append(groupsToDelete, remoteGroup)
		}
	}
	return groupsToDelete
}

func listGroupsToCreate(
	configGroupMap map[string]configuration.ComponentGroup,
	remoteGroupMap map[string]statuspagetypes.Group,
	componentNameToID map[string]string,
) ([]statuspagetypes.Group, error) {
	var groupsToCreate []statuspagetypes.Group
	for name, configGroup := range configGroupMap {
		if _, found := remoteGroupMap[name]; !found {
			var newGroup statuspagetypes.Group
			err := statuspagetypes.MergeConfigGroupToApi(configGroup, &newGroup, componentNameToID)
			if err != nil {
				return nil, err
			}
			groupsToCreate = append(groupsToCreate, newGroup)
		}
	}
	return groupsToCreate, nil
}

func listGroupsToModify(
	configGroupMap map[string]configuration.ComponentGroup,
	remoteGroupMap map[string]statuspagetypes.Group,
	componentNameToID map[string]string,
) ([]statuspagetypes.Group, error) {
	var groupsToModify []statuspagetypes.Group
	for name, configGroup := range configGroupMap {
		if remoteGroup, found := remoteGroupMap[name]; found {
			sort.Strings(remoteGroup.Components)
			var modifiedGroup statuspagetypes.Group
			err := mapstructure.Decode(remoteGroup, &modifiedGroup)
			if err != nil {
				return nil, fmt.Errorf("error decoding statuspage.Group to statuspage.Group: %w", err)
			}
			err = statuspagetypes.MergeConfigGroupToApi(configGroup, &modifiedGroup, componentNameToID)
			if err != nil {
				return nil, err
			}
			// if remote group is different from remote+configuration group, it must be modified
			if !reflect.DeepEqual(remoteGroup, modifiedGroup) {
				groupsToModify = append(groupsToModify, modifiedGroup)
			}
		}
	}
	return groupsToModify, nil
}

func ReconcileGroups(config *configuration.Config, client *resty.Client) error {
	componentNameToID, err := makeComponentMapping(client, config.Statuspage.PageID)
	if err != nil {
		return err
	}

	statuspageGroupNameToGroup, err := makeStatuspageGroupMapping(client, config.Statuspage.PageID)
	if err != nil {
		return err
	}

	configGroupNameToGroup := make(map[string]configuration.ComponentGroup)
	for _, configGroup := range config.Statuspage.Groups {
		configGroupNameToGroup[configGroup.Name] = configGroup
	}

	toDelete := listGroupsToDelete(configGroupNameToGroup, statuspageGroupNameToGroup)
	toCreate, err := listGroupsToCreate(configGroupNameToGroup, statuspageGroupNameToGroup, componentNameToID)
	if err != nil {
		return err
	}
	toModify, err := listGroupsToModify(configGroupNameToGroup, statuspageGroupNameToGroup, componentNameToID)
	if err != nil {
		return err
	}

	for _, group := range toDelete {
		shared.LogLn(config, fmt.Sprintf("deleting %s group from statuspage", group.Name),
			fmt.Sprintf(" - deleting: %+v", group))
		err := statuspage.DeleteGroup(client, config.Statuspage.PageID, group.ID)
		if err != nil {
			return err
		}
	}
	for _, group := range toCreate {
		shared.LogLn(config, fmt.Sprintf("creating %s group on statuspage", group.Name),
			fmt.Sprintf(" - new: %+v", group))
		_, err := statuspage.PostGroup(client, config.Statuspage.PageID, group)
		if err != nil {
			return err
		}
	}
	for _, group := range toModify {
		shared.LogLn(config, fmt.Sprintf("modifying %s group on statuspage", group.Name),
			fmt.Sprintf(" - config: %+v", configGroupNameToGroup[group.Name]),
			fmt.Sprintf(" - remote: %+v", statuspageGroupNameToGroup[group.Name]),
			fmt.Sprintf(" - modified: %+v", group))
		_, err := statuspage.PatchGroup(client, config.Statuspage.PageID, group.ID, group)
		if err != nil {
			return err
		}
	}

	return nil
}
