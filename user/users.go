package user

import (
	"fmt"
	"net/url"
	"strings"

	cfclient "github.com/cloudfoundry-community/go-cfclient"
	"github.com/pivotalservices/cf-mgmt/config"
	"github.com/pivotalservices/cf-mgmt/ldap"
	"github.com/pivotalservices/cf-mgmt/organization"
	"github.com/pivotalservices/cf-mgmt/space"
	"github.com/pivotalservices/cf-mgmt/uaa"
	"github.com/pkg/errors"
	"github.com/xchapter7x/lo"

	uaaclient "github.com/cloudfoundry-community/go-uaa"
)

// NewManager -
func NewManager(
	client CFClient,
	cfg config.Reader,
	spaceMgr space.Manager,
	orgMgr organization.Manager,
	uaaMgr uaa.Manager,
	peek bool) Manager {
	return &DefaultManager{
		Client:   client,
		Peek:     peek,
		SpaceMgr: spaceMgr,
		OrgMgr:   orgMgr,
		UAAMgr:   uaaMgr,
		Cfg:      cfg,
	}

}

type DefaultManager struct {
	Client     CFClient
	Cfg        config.Reader
	SpaceMgr   space.Manager
	OrgMgr     organization.Manager
	UAAMgr     uaa.Manager
	Peek       bool
	LdapMgr    ldap.Manager
	LdapConfig *config.LdapConfig
}

func (m *DefaultManager) RemoveSpaceAuditor(input UpdateUsersInput, userName string) error {
	if m.Peek {
		lo.G.Infof("[dry-run]: removing user %s from org/space %s/%s with role %s", userName, input.OrgName, input.SpaceName, "Auditor")
		return nil
	}
	lo.G.Infof("removing user %s from org/space %s/%s with role %s", userName, input.OrgName, input.SpaceName, "Auditor")
	return m.Client.RemoveSpaceAuditorByUsername(input.SpaceGUID, userName)
}
func (m *DefaultManager) RemoveSpaceDeveloper(input UpdateUsersInput, userName string) error {
	if m.Peek {
		lo.G.Infof("[dry-run]: removing user %s from org/space %s/%s with role %s", userName, input.OrgName, input.SpaceName, "Developer")
		return nil
	}
	lo.G.Infof("removing user %s from org/space %s/%s with role %s", userName, input.OrgName, input.SpaceName, "Developer")
	return m.Client.RemoveSpaceDeveloperByUsername(input.SpaceGUID, userName)
}
func (m *DefaultManager) RemoveSpaceManager(input UpdateUsersInput, userName string) error {
	if m.Peek {
		lo.G.Infof("[dry-run]: removing user %s from org/space %s/%s with role %s", userName, input.OrgName, input.SpaceName, "Manager")
		return nil
	}
	lo.G.Infof("removing user %s from org/space %s/%s with role %s", userName, input.OrgName, input.SpaceName, "Manager")
	return m.Client.RemoveSpaceManagerByUsername(input.SpaceGUID, userName)
}
func (m *DefaultManager) ListSpaceAuditors(spaceGUID string) (map[string]string, error) {
	if m.Peek && strings.Contains(spaceGUID, "dry-run-space-guid") {
		return nil, nil
	}
	users, err := m.Client.ListSpaceAuditors(spaceGUID)
	if err != nil {
		return nil, err
	}
	return m.userListToMap(users), nil
}
func (m *DefaultManager) ListSpaceDevelopers(spaceGUID string) (map[string]string, error) {
	if m.Peek && strings.Contains(spaceGUID, "dry-run-space-guid") {
		return nil, nil
	}
	users, err := m.Client.ListSpaceDevelopers(spaceGUID)
	if err != nil {
		return nil, err
	}
	return m.userListToMap(users), nil
}
func (m *DefaultManager) ListSpaceManagers(spaceGUID string) (map[string]string, error) {
	if m.Peek && strings.Contains(spaceGUID, "dry-run-space-guid") {
		return nil, nil
	}
	users, err := m.Client.ListSpaceManagers(spaceGUID)
	if err != nil {
		return nil, err
	}
	return m.userListToMap(users), nil
}

func (m *DefaultManager) listSpaceAuditors(input UpdateUsersInput) (map[string]string, error) {
	roleUsers, err := m.ListSpaceAuditors(input.SpaceGUID)
	if err == nil {
		lo.G.Debugf("RoleUsers for Org %s, Space %s and role %s: %+v", input.OrgName, input.SpaceName, "space-auditor", roleUsers)
	}
	return roleUsers, err
}
func (m *DefaultManager) listSpaceDevelopers(input UpdateUsersInput) (map[string]string, error) {
	roleUsers, err := m.ListSpaceDevelopers(input.SpaceGUID)
	if err == nil {
		lo.G.Debugf("RoleUsers for Org %s, Space %s and role %s: %+v", input.OrgName, input.SpaceName, "space-developer", roleUsers)
	}
	return roleUsers, err
}
func (m *DefaultManager) listSpaceManagers(input UpdateUsersInput) (map[string]string, error) {
	roleUsers, err := m.ListSpaceManagers(input.SpaceGUID)
	if err == nil {
		lo.G.Debugf("RoleUsers for Org %s, Space %s and role %s: %+v", input.OrgName, input.SpaceName, "space-manager", roleUsers)
	}
	return roleUsers, err
}

func (m *DefaultManager) userListToMap(users []cfclient.User) map[string]string {
	userMap := make(map[string]string)
	for _, user := range users {
		userMap[strings.ToLower(user.Username)] = user.Guid
	}
	return userMap
}

func (m *DefaultManager) AssociateSpaceAuditor(input UpdateUsersInput, userName string) error {
	err := m.AddUserToOrg(userName, input)
	if err != nil {
		return err
	}
	if m.Peek {
		lo.G.Infof("[dry-run]: adding %s to role %s for org/space %s/%s", userName, "auditor", input.OrgName, input.SpaceName)
		return nil
	}

	lo.G.Infof("adding %s to role %s for org/space %s/%s", userName, "auditor", input.OrgName, input.SpaceName)
	_, err = m.Client.AssociateSpaceAuditorByUsername(input.SpaceGUID, userName)
	return err
}
func (m *DefaultManager) AssociateSpaceDeveloper(input UpdateUsersInput, userName string) error {
	err := m.AddUserToOrg(userName, input)
	if err != nil {
		return err
	}
	if m.Peek {
		lo.G.Infof("[dry-run]: adding %s to role %s for org/space %s/%s", userName, "developer", input.OrgName, input.SpaceName)
		return nil
	}
	lo.G.Infof("adding %s to role %s for org/space %s/%s", userName, "developer", input.OrgName, input.SpaceName)
	_, err = m.Client.AssociateSpaceDeveloperByUsername(input.SpaceGUID, userName)
	return err
}
func (m *DefaultManager) AssociateSpaceManager(input UpdateUsersInput, userName string) error {
	err := m.AddUserToOrg(userName, input)
	if err != nil {
		return err
	}
	if m.Peek {
		lo.G.Infof("[dry-run]: adding %s to role %s for org/space %s/%s", userName, "manager", input.OrgName, input.SpaceName)
		return nil
	}

	lo.G.Infof("adding %s to role %s for org/space %s/%s", userName, "manager", input.OrgName, input.SpaceName)
	_, err = m.Client.AssociateSpaceManagerByUsername(input.SpaceGUID, userName)
	return err
}

func (m *DefaultManager) AddUserToOrg(userName string, input UpdateUsersInput) error {
	if m.Peek {
		return nil
	}
	_, err := m.Client.AssociateOrgUserByUsername(input.OrgGUID, userName)
	return err
}

func (m *DefaultManager) RemoveOrgAuditor(input UpdateUsersInput, userName string) error {
	if m.Peek {
		lo.G.Infof("[dry-run]: removing user %s from org %s with role %s", userName, input.OrgName, "auditor")
		return nil
	}
	lo.G.Infof("removing user %s from org %s with role %s", userName, input.OrgName, "auditor")
	return m.Client.RemoveOrgAuditorByUsername(input.OrgGUID, userName)
}
func (m *DefaultManager) RemoveOrgBillingManager(input UpdateUsersInput, userName string) error {
	if m.Peek {
		lo.G.Infof("[dry-run]: removing user %s from org %s with role %s", userName, input.OrgName, "billing manager")
		return nil
	}
	lo.G.Infof("removing user %s from org %s with role %s", userName, input.OrgName, "billing manager")
	return m.Client.RemoveOrgBillingManagerByUsername(input.OrgGUID, userName)
}

func (m *DefaultManager) RemoveOrgManager(input UpdateUsersInput, userName string) error {
	if m.Peek {
		lo.G.Infof("[dry-run]: removing user %s from org %s with role %s", userName, input.OrgName, "manager")
		return nil
	}
	lo.G.Infof("removing user %s from org %s with role %s", userName, input.OrgName, "manager")
	return m.Client.RemoveOrgManagerByUsername(input.OrgGUID, userName)
}

func (m *DefaultManager) ListOrgAuditors(orgGUID string) (map[string]string, error) {
	if m.Peek && strings.Contains(orgGUID, "dry-run-org-guid") {
		return nil, nil
	}
	users, err := m.Client.ListOrgAuditors(orgGUID)
	if err != nil {
		return nil, err
	}
	return m.userListToMap(users), nil
}
func (m *DefaultManager) ListOrgBillingManagers(orgGUID string) (map[string]string, error) {
	if m.Peek && strings.Contains(orgGUID, "dry-run-org-guid") {
		return nil, nil
	}
	users, err := m.Client.ListOrgBillingManagers(orgGUID)
	if err != nil {
		return nil, err
	}
	return m.userListToMap(users), nil
}
func (m *DefaultManager) ListOrgManagers(orgGUID string) (map[string]string, error) {
	if m.Peek && strings.Contains(orgGUID, "dry-run-org-guid") {
		return nil, nil
	}
	users, err := m.Client.ListOrgManagers(orgGUID)
	if err != nil {
		return nil, err
	}
	return m.userListToMap(users), nil
}
func (m *DefaultManager) listOrgAuditors(input UpdateUsersInput) (map[string]string, error) {
	roleUsers, err := m.ListOrgAuditors(input.OrgGUID)
	if err == nil {
		lo.G.Debugf("RoleUsers for Org %s and role %s: %+v", input.OrgName, "org-auditor", roleUsers)
	}
	return roleUsers, err
}
func (m *DefaultManager) listOrgBillingManagers(input UpdateUsersInput) (map[string]string, error) {
	roleUsers, err := m.ListOrgBillingManagers(input.OrgGUID)
	if err == nil {
		lo.G.Debugf("RoleUsers for Org %s and role %s: %+v", input.OrgName, "org-billing-manager", roleUsers)
	}
	return roleUsers, err
}
func (m *DefaultManager) listOrgManagers(input UpdateUsersInput) (map[string]string, error) {
	roleUsers, err := m.ListOrgManagers(input.OrgGUID)
	if err == nil {
		lo.G.Debugf("RoleUsers for Org %s and role %s: %+v", input.OrgName, "org-manager", roleUsers)
	}
	return roleUsers, err
}

func (m *DefaultManager) AssociateOrgAuditor(input UpdateUsersInput, userName string) error {
	err := m.AddUserToOrg(userName, input)
	if err != nil {
		return err
	}
	if m.Peek {
		lo.G.Infof("[dry-run]: Add User %s to role %s for org %s", userName, "auditor", input.OrgName)
		return nil
	}

	lo.G.Infof("Add User %s to role %s for org %s", userName, "auditor", input.OrgName)
	_, err = m.Client.AssociateOrgAuditorByUsername(input.OrgGUID, userName)
	return err
}
func (m *DefaultManager) AssociateOrgBillingManager(input UpdateUsersInput, userName string) error {
	err := m.AddUserToOrg(userName, input)
	if err != nil {
		return err
	}
	if m.Peek {
		lo.G.Infof("[dry-run]: Add User %s to role %s for org %s", userName, "billing manager", input.OrgName)
		return nil
	}

	lo.G.Infof("Add User %s to role %s for org %s", userName, "billing manager", input.OrgName)
	_, err = m.Client.AssociateOrgBillingManagerByUsername(input.OrgGUID, userName)
	return err
}

func (m *DefaultManager) AssociateOrgManager(input UpdateUsersInput, userName string) error {
	err := m.AddUserToOrg(userName, input)
	if err != nil {
		return err
	}
	if m.Peek {
		lo.G.Infof("[dry-run]: Add User %s to role %s for org %s", userName, "manager", input.OrgName)
		return nil
	}

	lo.G.Infof("Add User %s to role %s for org %s", userName, "manager", input.OrgName)
	_, err = m.Client.AssociateOrgManagerByUsername(input.OrgGUID, userName)
	return err
}

//UpdateSpaceUsers -
func (m *DefaultManager) UpdateSpaceUsers() error {
	uaaUsers, err := m.UAAMgr.ListUsers()
	if err != nil {
		return err
	}

	spaceConfigs, err := m.Cfg.GetSpaceConfigs()
	if err != nil {
		return err
	}

	for _, input := range spaceConfigs {
		if err := m.updateSpaceUsers(&input, uaaUsers); err != nil {
			return err
		}
	}

	return nil
}

func (m *DefaultManager) updateSpaceUsers(input *config.SpaceConfig, uaaUsers map[string]*uaaclient.User) error {
	space, err := m.SpaceMgr.FindSpace(input.Org, input.Space)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Error finding space for org %s, space %s", input.Org, input.Space))
	}

	if err = m.SyncUsers(uaaUsers, UpdateUsersInput{
		SpaceName:      space.Name,
		SpaceGUID:      space.Guid,
		OrgName:        input.Org,
		OrgGUID:        space.OrganizationGuid,
		LdapGroupNames: input.GetDeveloperGroups(),
		LdapUsers:      input.Developer.LDAPUsers,
		Users:          input.Developer.Users,
		SamlUsers:      input.Developer.SamlUsers,
		RemoveUsers:    input.RemoveUsers,
		ListUsers:      m.listSpaceDevelopers,
		RemoveUser:     m.RemoveSpaceDeveloper,
		AddUser:        m.AssociateSpaceDeveloper,
	}); err != nil {
		return errors.Wrap(err, fmt.Sprintf("Error syncing users for org %s, space %s, role %s", input.Org, input.Space, "developer"))
	}

	if err = m.SyncUsers(uaaUsers,
		UpdateUsersInput{
			SpaceName:      space.Name,
			SpaceGUID:      space.Guid,
			OrgGUID:        space.OrganizationGuid,
			OrgName:        input.Org,
			LdapGroupNames: input.GetManagerGroups(),
			LdapUsers:      input.Manager.LDAPUsers,
			Users:          input.Manager.Users,
			SamlUsers:      input.Manager.SamlUsers,
			RemoveUsers:    input.RemoveUsers,
			ListUsers:      m.listSpaceManagers,
			RemoveUser:     m.RemoveSpaceManager,
			AddUser:        m.AssociateSpaceManager,
		}); err != nil {
		return errors.Wrap(err, fmt.Sprintf("Error syncing users for org %s, space %s, role %s", input.Org, input.Space, "manager"))
	}
	if err = m.SyncUsers(uaaUsers,
		UpdateUsersInput{
			SpaceName:      space.Name,
			SpaceGUID:      space.Guid,
			OrgGUID:        space.OrganizationGuid,
			OrgName:        input.Org,
			LdapGroupNames: input.GetAuditorGroups(),
			LdapUsers:      input.Auditor.LDAPUsers,
			Users:          input.Auditor.Users,
			SamlUsers:      input.Auditor.SamlUsers,
			RemoveUsers:    input.RemoveUsers,
			ListUsers:      m.listSpaceAuditors,
			RemoveUser:     m.RemoveSpaceAuditor,
			AddUser:        m.AssociateSpaceAuditor,
		}); err != nil {
		return errors.Wrap(err, fmt.Sprintf("Error syncing users for org %s, space %s, role %s", input.Org, input.Space, "auditor"))
	}
	return nil
}

//UpdateOrgUsers -
func (m *DefaultManager) UpdateOrgUsers() error {
	uaacUsers, err := m.UAAMgr.ListUsers()
	if err != nil {
		return err
	}

	orgConfigs, err := m.Cfg.GetOrgConfigs()
	if err != nil {
		return err
	}

	for _, input := range orgConfigs {
		if err := m.updateOrgUsers(&input, uaacUsers); err != nil {
			return err
		}

	}

	return nil
}

//CleanupOrgUsers -
func (m *DefaultManager) CleanupOrgUsers() error {
	orgConfigs, err := m.Cfg.GetOrgConfigs()
	if err != nil {
		return err
	}

	for _, input := range orgConfigs {
		if err := m.cleanupOrgUsers(&input); err != nil {
			return err
		}
	}
	return nil
}

func (m *DefaultManager) cleanupOrgUsers(input *config.OrgConfig) error {
	org, err := m.OrgMgr.FindOrg(input.Org)
	if err != nil {
		return err
	}
	orgUsers, err := m.Client.ListOrgUsers(org.Guid)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Error listing org users for org %s", input.Org))
	}

	usersInRoles, err := m.usersInOrgRoles(org.Name, org.Guid)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Error usersInOrgRoles for org %s", input.Org))
	}

	lo.G.Debugf("Users In Roles %+v", usersInRoles)

	for _, orgUser := range orgUsers {
		if _, ok := usersInRoles[strings.ToLower(orgUser.Username)]; !ok {
			if m.Peek {
				lo.G.Infof("[dry-run]: Removing User %s from org %s", orgUser.Username, input.Org)
				continue
			}

			lo.G.Infof("Removing User %s from org %s", orgUser.Username, input.Org)
			err := m.Client.RemoveOrgUserByUsername(org.Guid, orgUser.Username)
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("Error removing user %s from org %s", orgUser.Username, input.Org))
			}
		}

	}

	return nil
}

func (m *DefaultManager) usersInOrgRoles(orgName, orgGUID string) (map[string]string, error) {
	userMap := make(map[string]string)

	orgAuditors, err := m.ListOrgAuditors(orgGUID)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("Error listing org auditors for org %s", orgName))
	}
	m.appendToMap(userMap, orgAuditors)

	orgManagers, err := m.ListOrgManagers(orgGUID)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("Error listing org managers for org %s", orgName))
	}
	m.appendToMap(userMap, orgManagers)

	orgBillingManagers, err := m.ListOrgBillingManagers(orgGUID)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("Error listing org billing managers for org %s", orgName))
	}
	m.appendToMap(userMap, orgBillingManagers)

	spaces, err := m.listSpaces(orgGUID)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("Error listing spaces for org %s", orgName))
	}
	for _, space := range spaces {
		spaceAuditors, err := m.ListSpaceAuditors(space.Guid)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("Error listing space auditors for org/space %s/%s", orgName, space.Name))
		}
		m.appendToMap(userMap, spaceAuditors)

		spaceDevelopers, err := m.ListSpaceAuditors(space.Guid)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("Error listing space developers for org/space %s/%s", orgName, space.Name))
		}
		m.appendToMap(userMap, spaceDevelopers)

		spaceManagers, err := m.ListSpaceManagers(space.Guid)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("Error listing space managers for org/space %s/%s", orgName, space.Name))
		}
		m.appendToMap(userMap, spaceManagers)
	}

	return userMap, nil
}

func (m *DefaultManager) appendToMap(source, append map[string]string) {
	for userName, GUID := range append {
		source[userName] = GUID
	}
}

func (m *DefaultManager) listSpaces(orgGUID string) ([]cfclient.Space, error) {
	spaces, err := m.Client.ListSpacesByQuery(url.Values{
		"q": []string{fmt.Sprintf("%s:%s", "organization_guid", orgGUID)},
	})
	if err != nil {
		return nil, err
	}
	return spaces, err

}

func (m *DefaultManager) updateOrgUsers(input *config.OrgConfig, uaacUsers map[string]*uaaclient.User) error {
	org, err := m.OrgMgr.FindOrg(input.Org)
	if err != nil {
		return err
	}

	err = m.SyncUsers(
		uaacUsers, UpdateUsersInput{
			OrgName:        org.Name,
			OrgGUID:        org.Guid,
			LdapGroupNames: input.GetBillingManagerGroups(),
			LdapUsers:      input.BillingManager.LDAPUsers,
			Users:          input.BillingManager.Users,
			SamlUsers:      input.BillingManager.SamlUsers,
			RemoveUsers:    input.RemoveUsers,
			ListUsers:      m.listOrgBillingManagers,
			RemoveUser:     m.RemoveOrgBillingManager,
			AddUser:        m.AssociateOrgBillingManager,
		})
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Error syncing users for org %s role %s", input.Org, "billing_managers"))
	}

	err = m.SyncUsers(
		uaacUsers, UpdateUsersInput{
			OrgName:        org.Name,
			OrgGUID:        org.Guid,
			LdapGroupNames: input.GetAuditorGroups(),
			LdapUsers:      input.Auditor.LDAPUsers,
			Users:          input.Auditor.Users,
			SamlUsers:      input.Auditor.SamlUsers,
			RemoveUsers:    input.RemoveUsers,
			ListUsers:      m.listOrgAuditors,
			RemoveUser:     m.RemoveOrgAuditor,
			AddUser:        m.AssociateOrgAuditor,
		})
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Error syncing users for org %s role %s", input.Org, "org-auditors"))
	}

	err = m.SyncUsers(
		uaacUsers, UpdateUsersInput{
			OrgName:        org.Name,
			OrgGUID:        org.Guid,
			LdapGroupNames: input.GetManagerGroups(),
			LdapUsers:      input.Manager.LDAPUsers,
			Users:          input.Manager.Users,
			SamlUsers:      input.Manager.SamlUsers,
			RemoveUsers:    input.RemoveUsers,
			ListUsers:      m.listOrgManagers,
			RemoveUser:     m.RemoveOrgManager,
			AddUser:        m.AssociateOrgManager,
		})

	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Error syncing users for org %s role %s", input.Org, "org-manager"))
	}

	return nil
}

//SyncUsers
func (m *DefaultManager) SyncUsers(uaaUsers map[string]*uaaclient.User, updateUsersInput UpdateUsersInput) error {
	roleUsers, err := updateUsersInput.ListUsers(updateUsersInput)
	if err != nil {
		return err
	}

	if err := m.SyncLdapUsers(roleUsers, uaaUsers, updateUsersInput); err != nil {
		return err
	}
	if err := m.SyncInternalUsers(roleUsers, uaaUsers, updateUsersInput); err != nil {
		return err
	}
	if err := m.SyncSamlUsers(roleUsers, uaaUsers, updateUsersInput); err != nil {
		return err
	}
	if err := m.RemoveUsers(roleUsers, updateUsersInput); err != nil {
		return err
	}
	return nil
}

func (m *DefaultManager) SyncInternalUsers(roleUsers map[string]string, uaaUsers map[string]*uaaclient.User, updateUsersInput UpdateUsersInput) error {
	for _, userID := range updateUsersInput.Users {
		lowerUserID := strings.ToLower(userID)
		if _, userExists := uaaUsers[lowerUserID]; !userExists {
			return fmt.Errorf("user %s doesn't exist in cloud foundry, so must add internal user first", lowerUserID)
		}
		if _, ok := roleUsers[lowerUserID]; !ok {
			if err := updateUsersInput.AddUser(updateUsersInput, userID); err != nil {
				return err
			}
		} else {
			delete(roleUsers, lowerUserID)
		}
	}
	return nil
}

func (m *DefaultManager) SyncSamlUsers(roleUsers map[string]string, uaaUsers map[string]*uaaclient.User, updateUsersInput UpdateUsersInput) error {
	for _, userEmail := range updateUsersInput.SamlUsers {
		lowerUserEmail := strings.ToLower(userEmail)
		if _, userExists := uaaUsers[lowerUserEmail]; !userExists {
			lo.G.Debug("User", userEmail, "doesn't exist in cloud foundry, so creating user")
			if err := m.UAAMgr.CreateExternalUser(userEmail, userEmail, userEmail, m.LdapConfig.Origin); err != nil {
				lo.G.Error("Unable to create user", userEmail)
				continue
			} else {
				uaaUsers[userEmail] = &uaaclient.User{
					Username:   userEmail,
					Emails:     []uaaclient.Email{uaaclient.Email{Value: userEmail}},
					ExternalID: userEmail,
					Origin:     m.LdapConfig.Origin,
				}
			}
		}
		if _, ok := roleUsers[lowerUserEmail]; !ok {
			if err := updateUsersInput.AddUser(updateUsersInput, userEmail); err != nil {
				return err
			}
		} else {
			delete(roleUsers, lowerUserEmail)
		}
	}
	return nil
}

func (m *DefaultManager) RemoveUsers(roleUsers map[string]string, updateUsersInput UpdateUsersInput) error {
	if updateUsersInput.RemoveUsers {
		for roleUser, _ := range roleUsers {
			if err := updateUsersInput.RemoveUser(updateUsersInput, roleUser); err != nil {
				return err
			}
		}
	} else {
		if updateUsersInput.SpaceName == "" {
			lo.G.Debugf("Not removing users. Set enable-remove-users: true to orgConfig for org: %s", updateUsersInput.OrgName)
		} else {
			lo.G.Debugf("Not removing users. Set enable-remove-users: true to spaceConfig for org/space: %s/%s", updateUsersInput.OrgName, updateUsersInput.SpaceName)
		}
	}
	return nil
}

func (m *DefaultManager) InitializeLdap(ldapBindPassword string) error {
	ldapConfig, err := m.Cfg.LdapConfig(ldapBindPassword)
	if err != nil {
		return err
	}
	m.LdapConfig = ldapConfig
	if m.LdapConfig.Enabled {
		ldapMgr, err := ldap.NewManager(ldapConfig)
		if err != nil {
			return err
		}
		m.LdapMgr = ldapMgr
	}
	return nil
}

func (m *DefaultManager) DeinitializeLdap() error {
	if m.LdapMgr != nil {
		m.LdapMgr.Close()
	}
	return nil
}
