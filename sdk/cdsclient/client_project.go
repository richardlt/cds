package cdsclient

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"

	"github.com/ovh/cds/engine/api/permission"
	"github.com/ovh/cds/sdk"
)

func (c *client) ProjectCreate(p *sdk.Project, groupName string) error {
	if groupName != "" {
		// if the group does not exist, POST /project will create it
		p.ProjectGroups = []sdk.GroupPermission{
			sdk.GroupPermission{
				Group:      sdk.Group{Name: groupName},
				Permission: permission.PermissionReadWriteExecute,
			},
		}
	}

	if _, err := c.PostJSON("/project", p, nil); err != nil {
		return err
	}
	return nil
}

func (c *client) ProjectDelete(key string) error {
	_, err := c.DeleteJSON("/project/"+key, nil, nil)
	return err
}

func (c *client) ProjectGet(key string, mods ...RequestModifier) (*sdk.Project, error) {
	p := &sdk.Project{}
	if _, err := c.GetJSON("/project/"+key, p, mods...); err != nil {
		return nil, err
	}
	return p, nil
}

func (c *client) ProjectList(withApplications, withWorkflows bool, filters ...Filter) ([]sdk.Project, error) {
	p := []sdk.Project{}
	path := fmt.Sprintf("/project?application=%v&workflow=%v", withApplications, withWorkflows)

	for _, f := range filters {
		path += fmt.Sprintf("&%s=%s", url.QueryEscape(f.Name), url.QueryEscape(f.Value))
	}

	if _, err := c.GetJSON(path, &p); err != nil {
		return nil, err
	}
	return p, nil
}

func (c *client) ProjectGroupsImport(projectKey string, content io.Reader, format string, force bool) (sdk.Project, error) {
	var proj sdk.Project
	url := fmt.Sprintf("/project/%s/group/import?format=%s", projectKey, format)

	if force {
		url += "&forceUpdate=true"
	}

	btes, _, _, errReq := c.Request("POST", url, content)
	if errReq != nil {
		return proj, errReq
	}

	if err := json.Unmarshal(btes, &proj); err != nil {
		return proj, err
	}

	return proj, nil
}
