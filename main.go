package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/pivotalservices/cf-mgmt/cloudcontroller"
	"github.com/pivotalservices/cf-mgmt/config"
	"github.com/pivotalservices/cf-mgmt/generated"
	"github.com/pivotalservices/cf-mgmt/importconfig"
	"github.com/pivotalservices/cf-mgmt/organization"
	"github.com/pivotalservices/cf-mgmt/space"
	"github.com/pivotalservices/cf-mgmt/uaa"
	"github.com/pivotalservices/cf-mgmt/uaac"
	"github.com/xchapter7x/lo"
)

var (
	//VERSION -
	VERSION string
)

//ErrorHandler -
type ErrorHandler struct {
	ExitCode int
	Error    error
}

type flagBucket struct {
	Desc        string
	EnvVar      string
	StringSlice bool
}

//CFMgmt -
type CFMgmt struct {
	UAAManager    uaa.Manager
	OrgManager    organization.Manager
	SpaceManager  space.Manager
	ConfigManager config.Manager
	ConfigDir     string
	LdapBindPwd   string
	uaacToken     string
	systemDomain  string
	UAACManager   uaac.Manager
}

//InitializeManager -
func InitializeManager(c *cli.Context) (*CFMgmt, error) {
	var err error
	configDir := getConfigDir(c)
	if configDir == "" {
		err = fmt.Errorf("Config directory name is required")
		return nil, err
	}
	sysDomain := c.String(getFlag(systemDomain))
	user := c.String(getFlag(userID))
	pwd := c.String(getFlag(password))
	secret := c.String(getFlag(clientSecret))
	ldapPwd := c.String(getFlag(ldapPassword))

	if sysDomain == "" ||
		user == "" ||
		pwd == "" ||
		secret == "" {
		err = fmt.Errorf("Must set system-domain, user-id, password, client-secret properties")
		return nil, err
	}

	var cfToken, uaacToken string
	cfMgmt := &CFMgmt{}
	cfMgmt.LdapBindPwd = ldapPwd
	cfMgmt.UAAManager = uaa.NewDefaultUAAManager(sysDomain, user)
	if cfToken, err = cfMgmt.UAAManager.GetCFToken(pwd); err != nil {
		return nil, err
	}
	if uaacToken, err = cfMgmt.UAAManager.GetUAACToken(secret); err != nil {
		return nil, err
	}
	cfMgmt.uaacToken = uaacToken
	cfMgmt.systemDomain = systemDomain
	cfMgmt.OrgManager = organization.NewManager(sysDomain, cfToken, uaacToken)
	cfMgmt.SpaceManager = space.NewManager(sysDomain, cfToken, uaacToken)
	cfMgmt.ConfigManager = config.NewManager(configDir)
	cfMgmt.ConfigDir = configDir
	cfMgmt.UAACManager = uaac.NewManager(systemDomain, uaacToken)
	return cfMgmt, nil
}

const (
	systemDomain     string = "SYSTEM_DOMAIN"
	userID           string = "USER_ID"
	password         string = "PASSWORD"
	clientSecret     string = "CLIENT_SECRET"
	configDir        string = "CONFIG_DIR"
	orgName          string = "ORG"
	spaceName        string = "SPACE"
	ldapPassword     string = "LDAP_PASSWORD"
	orgBillingMgrGrp string = "ORG_BILLING_MGR_GRP"
	orgMgrGrp        string = "ORG_MGR_GRP"
	orgAuditorGrp    string = "ORG_AUDITOR_GRP"
	spaceDevGrp      string = "SPACE_DEV_GRP"
	spaceMgrGrp      string = "SPACE_MGR_GRP"
	spaceAuditorGrp  string = "SPACE_AUDITOR_GRP"
)

func main() {
	eh := new(ErrorHandler)
	eh.ExitCode = 0
	app := NewApp(eh)
	if err := app.Run(os.Args); err != nil {
		eh.ExitCode = 1
		eh.Error = err
		lo.G.Error(eh.Error)
	}
	os.Exit(eh.ExitCode)
}

// NewApp creates a new cli app
func NewApp(eh *ErrorHandler) *cli.App {
	//cli.AppHelpTemplate = CfopsHelpTemplate
	app := cli.NewApp()
	app.Version = VERSION
	app.Name = "cf-mgmt"
	app.Usage = "cf-mgmt"
	app.Commands = []cli.Command{
		{
			Name:  "version",
			Usage: "shows the application version currently in use",
			Action: func(c *cli.Context) (err error) {
				cli.ShowVersion(c)
				return
			},
		},
		CreateInitCommand(eh),
		CreateAddOrgCommand(eh),
		CreateAddSpaceCommand(eh),
		CreateImportConfigCommand(eh),
		CreateGeneratePipelineCommand(runGeneratePipeline, eh),
		CreateCommand("create-orgs", runCreateOrgs, defaultFlags(), eh),
		CreateCommand("update-org-quotas", runCreateOrgQuotas, defaultFlags(), eh),
		CreateCommand("update-org-users", runUpdateOrgUsers, defaultFlagsWithLdap(), eh),
		CreateCommand("create-spaces", runCreateSpaces, defaultFlagsWithLdap(), eh),
		CreateCommand("update-spaces", runUpdateSpaces, defaultFlags(), eh),
		CreateCommand("update-space-quotas", runCreateSpaceQuotas, defaultFlags(), eh),
		CreateCommand("update-space-users", runUpdateSpaceUsers, defaultFlagsWithLdap(), eh),
		CreateCommand("update-space-security-groups", runCreateSpaceSecurityGroups, defaultFlags(), eh),
	}

	return app
}

// CreateImportConfigCommand -
func CreateImportConfigCommand(eh *ErrorHandler) (command cli.Command) {
	flags := defaultFlags()
	flag := cli.StringSliceFlag{
		Name:  "excluded-org",
		Usage: "Orgs to be excluded from import",
	}
	flags = append(flags, flag)
	command = cli.Command{
		Name:        "import-config",
		Usage:       "import-config --excluded-org <orgname> (Repeat the flag for specifying multiple org names)  ",
		Description: "Imports org and space configurations from an existing Cloud Foundry instance. [Warning: This operation will delete existing config folder]",
		Action:      runImportConfig,
		Flags:       flags,
	}
	return
}

//CreateInitCommand -
func CreateInitCommand(eh *ErrorHandler) (command cli.Command) {
	flagList := map[string]flagBucket{
		configDir: {
			Desc:   "Name of the config directory. Default config directory is `config`",
			EnvVar: configDir,
		},
	}

	command = cli.Command{
		Name:        "init-config",
		Usage:       "Initializes folder structure for configuration",
		Description: "Initializes folder structure for configuration",
		Action:      runInit,
		Flags:       buildFlags(flagList),
	}
	return
}

func runInit(c *cli.Context) (err error) {
	configDir := getConfigDir(c)
	configManager := config.NewManager(configDir)
	err = configManager.CreateConfigIfNotExists("ldap")
	return err
}

//CreateAddOrgCommand -
func CreateAddOrgCommand(eh *ErrorHandler) (command cli.Command) {
	flagList := map[string]flagBucket{
		configDir: {
			Desc:   "Config directory name.  Default is config",
			EnvVar: configDir,
		},
		orgName: {
			Desc:   "Org name to add",
			EnvVar: orgName,
		},
		orgBillingMgrGrp: {
			Desc:   "LDAP group for Org Billing Manager",
			EnvVar: orgBillingMgrGrp,
		},
		orgMgrGrp: {
			Desc:   "LDAP group for Org Manager",
			EnvVar: orgMgrGrp,
		},
		orgAuditorGrp: {
			Desc:   "LDAP group for Org Auditor",
			EnvVar: orgAuditorGrp,
		},
	}

	command = cli.Command{
		Name:        "add-org-to-config",
		Usage:       "Adds specified org to configuration",
		Description: "Adds specified org to configuration",
		Action:      runAddOrg,
		Flags:       buildFlags(flagList),
	}
	return
}

func runAddOrg(c *cli.Context) error {
	inputOrg := c.String(getFlag(orgName))
	var cfMgmt *CFMgmt
	orgConfig := &config.OrgConfig{OrgName: inputOrg,
		OrgBillingMgrLDAPGrp: c.String(getFlag(orgBillingMgrGrp)),
		OrgMgrLDAPGrp:        c.String(getFlag(orgMgrGrp)),
		OrgAuditorLDAPGrp:    c.String(getFlag(orgAuditorGrp)),
	}
	return cfMgmt.ConfigManager.AddOrgToConfig(orgConfig)
}

//CreateAddSpaceCommand -
func CreateAddSpaceCommand(eh *ErrorHandler) (command cli.Command) {
	flagList := map[string]flagBucket{
		configDir: {
			Desc:   "config dir.  Default is config",
			EnvVar: configDir,
		},
		orgName: {
			Desc:   "org name of space",
			EnvVar: orgName,
		},
		spaceName: {
			Desc:   "space name to add",
			EnvVar: spaceName,
		},
		spaceDevGrp: {
			Desc:   "LDAP group for Space Developer",
			EnvVar: spaceDevGrp,
		},
		spaceMgrGrp: {
			Desc:   "LDAP group for Space Manager",
			EnvVar: spaceMgrGrp,
		},
		spaceAuditorGrp: {
			Desc:   "LDAP group for Space Auditor",
			EnvVar: spaceAuditorGrp,
		},
	}

	command = cli.Command{
		Name:        "add-space-to-config",
		Usage:       "adds specified space to configuration for org",
		Description: "adds specified space to configuration for org",
		Action:      runAddSpace,
		Flags:       buildFlags(flagList),
	}
	return
}

func runAddSpace(c *cli.Context) (err error) {

	inputOrg := c.String(getFlag(orgName))
	inputSpace := c.String(getFlag(spaceName))

	spaceConfig := &config.SpaceConfig{OrgName: inputOrg,
		SpaceName:       inputSpace,
		SpaceDevGrp:     c.String(getFlag(spaceDevGrp)),
		SpaceMgrGrp:     c.String(getFlag(spaceMgrGrp)),
		SpaceAuditorGrp: c.String(getFlag(spaceAuditorGrp)),
	}

	configDr := getConfigDir(c)
	if inputOrg == "" || inputSpace == "" {
		err = fmt.Errorf("Must provide org name and space name")
	} else {
		err = config.NewManager(configDr).AddSpaceToConfig(spaceConfig)
	}
	return
}

//CreateGeneratePipelineCommand -
func CreateGeneratePipelineCommand(action func(c *cli.Context) (err error), eh *ErrorHandler) (command cli.Command) {
	command = cli.Command{
		Name:        "generate-concourse-pipeline",
		Usage:       "generates a concourse pipline based on convention of org/space metadata",
		Description: "generate-concourse-pipeline",
		Action:      action,
	}
	return
}

func runGeneratePipeline(c *cli.Context) (err error) {
	const varsFileName = "vars.yml"
	const pipelineFileName = "pipeline.yml"
	const cfMgmtYml = "cf-mgmt.yml"
	const cfMgmtSh = "cf-mgmt.sh"
	var targetFile string
	fmt.Println("Generating pipeline....")
	if err = createFile(pipelineFileName, pipelineFileName); err != nil {
		lo.G.Error("Error creating pipeline.yml", err)
		return
	}
	if err = createFile(varsFileName, varsFileName); err != nil {
		lo.G.Error("Error creating vars.yml", err)
		return
	}
	if err = os.MkdirAll("ci/tasks", 0755); err == nil {
		targetFile = fmt.Sprintf("./ci/tasks/%s", cfMgmtYml)
		lo.G.Debug("Creating", targetFile)
		if err = createFile(cfMgmtYml, targetFile); err != nil {
			lo.G.Error("Error creating cf-mgmt.yml", err)
			return
		}
		targetFile = fmt.Sprintf("./ci/tasks/%s", cfMgmtSh)
		lo.G.Debug("Creating", targetFile)
		if err = createFile(cfMgmtSh, targetFile); err != nil {
			lo.G.Error("Error creating cf-mgmt.sh", err)
			return
		}
	}
	fmt.Println("1) Update vars.yml with the appropriate values")
	fmt.Println("2) Using following command to set your pipeline in concourse after you have checked all files in to GIT")
	fmt.Println("fly -t lite set-pipeline -p cf-mgmt -c pipeline.yml --load-vars-from=vars.yml")
	return
}

func createFile(assetName, fileName string) (err error) {
	var f *os.File
	var fileBytes []byte
	if fileBytes, err = generated.Asset(fmt.Sprintf("files/%s", assetName)); err == nil {
		if f, err = os.Create(fileName); err == nil {
			defer f.Close()
			_, err = f.Write(fileBytes)
		}
	}
	return
}

//CreateCommand -
func CreateCommand(commandName string, action func(c *cli.Context) (err error), flags []cli.Flag, eh *ErrorHandler) (command cli.Command) {
	desc := fmt.Sprintf(commandName)
	command = cli.Command{
		Name:        commandName,
		Usage:       fmt.Sprintf("%s with what is defined in config", commandName),
		Description: desc,
		Action:      action,
		Flags:       flags,
	}
	return
}

func runCreateOrgs(c *cli.Context) error {
	var cfMgmt *CFMgmt
	var err error
	if cfMgmt, err = InitializeManager(c); err == nil {
		err = cfMgmt.OrgManager.CreateOrgs(cfMgmt.ConfigDir)
	}
	return err
}

func runCreateOrgQuotas(c *cli.Context) error {
	var cfMgmt *CFMgmt
	var err error
	if cfMgmt, err = InitializeManager(c); err == nil {
		err = cfMgmt.OrgManager.CreateQuotas(cfMgmt.ConfigDir)
	}
	return err
}

func runCreateSpaceQuotas(c *cli.Context) error {
	var cfMgmt *CFMgmt
	var err error
	if cfMgmt, err = InitializeManager(c); err == nil {
		err = cfMgmt.SpaceManager.CreateQuotas(cfMgmt.ConfigDir)
	}
	return err
}

func runCreateSpaceSecurityGroups(c *cli.Context) error {
	var cfMgmt *CFMgmt
	var err error
	if cfMgmt, err = InitializeManager(c); err == nil {
		err = cfMgmt.SpaceManager.CreateApplicationSecurityGroups(cfMgmt.ConfigDir)
	}
	return err
}

func runCreateSpaces(c *cli.Context) error {
	var cfMgmt *CFMgmt
	var err error
	if cfMgmt, err = InitializeManager(c); err == nil {
		err = cfMgmt.SpaceManager.CreateSpaces(cfMgmt.ConfigDir, cfMgmt.LdapBindPwd)
	}
	return err
}

func runUpdateSpaces(c *cli.Context) error {
	var cfMgmt *CFMgmt
	var err error
	if cfMgmt, err = InitializeManager(c); err == nil {
		err = cfMgmt.SpaceManager.UpdateSpaces(cfMgmt.ConfigDir)
	}
	return err
}

func runUpdateSpaceUsers(c *cli.Context) error {
	var cfMgmt *CFMgmt
	var err error
	if cfMgmt, err = InitializeManager(c); err == nil {
		err = cfMgmt.SpaceManager.UpdateSpaceUsers(cfMgmt.ConfigDir, cfMgmt.LdapBindPwd)
	}
	return err
}

func runUpdateOrgUsers(c *cli.Context) error {
	var cfMgmt *CFMgmt
	var err error
	if cfMgmt, err = InitializeManager(c); err == nil {
		err = cfMgmt.OrgManager.UpdateOrgUsers(cfMgmt.ConfigDir, cfMgmt.LdapBindPwd)
	}
	return err
}

func defaultFlagsWithLdap() (flags []cli.Flag) {
	flags = defaultFlags()
	flag := cli.StringFlag{
		Name:   getFlag(ldapPassword),
		EnvVar: ldapPassword,
		Usage:  "Ldap password for binding",
	}
	flags = append(flags, flag)
	return
}

func defaultFlags() (flags []cli.Flag) {
	var flagList = buildDefaultFlags()
	flags = buildFlags(flagList)
	return
}

func buildDefaultFlags() (flagList map[string]flagBucket) {
	flagList = map[string]flagBucket{
		systemDomain: {
			Desc:   "system domain",
			EnvVar: systemDomain,
		},
		userID: {
			Desc:   "user id that has admin priv",
			EnvVar: userID,
		},
		password: {
			Desc:   "password for user account that has admin priv",
			EnvVar: password,
		},
		clientSecret: {
			Desc:   "secret for user account that has admin priv",
			EnvVar: clientSecret,
		},
		configDir: {
			Desc:   "config dir.  Default is config",
			EnvVar: configDir,
		},
	}
	return
}
func buildFlags(flagList map[string]flagBucket) (flags []cli.Flag) {
	for _, v := range flagList {
		if v.StringSlice {
			flags = append(flags, cli.StringSliceFlag{
				Name:   getFlag(v.EnvVar),
				Usage:  v.Desc,
				EnvVar: v.EnvVar,
			})
		} else {
			flags = append(flags, cli.StringFlag{
				Name:   getFlag(v.EnvVar),
				Value:  "",
				Usage:  v.Desc,
				EnvVar: v.EnvVar,
			})
		}
	}
	return
}

func getFlag(input string) string {
	return strings.ToLower(strings.Replace(input, "_", "-", -1))
}

func getConfigDir(c *cli.Context) (cDir string) {
	cDir = c.String(getFlag(configDir))
	if cDir == "" {
		return "config"
	}
	return cDir
}

func runImportConfig(c *cli.Context) error {
	var cfMgmt *CFMgmt
	var err error
	cfMgmt, err = InitializeManager(c)
	if cfMgmt != nil {
		cloudController := cloudcontroller.NewManager(fmt.Sprintf("https://api.%s", cfMgmt.systemDomain), cfMgmt.uaacToken)
		importManager := importconfig.NewManager(cfMgmt.ConfigDir, cfMgmt.UAACManager, cfMgmt.OrgManager, cfMgmt.SpaceManager, cloudController)
		ignoredOrgs := make(map[string]string)
		ignoredOrgs["system"] = "system"
		ignoreOrgs := c.StringSlice(getFlag(configDir))
		for _, org := range ignoreOrgs {
			ignoredOrgs[org] = org
		}
		err = importManager.ImportConfig(ignoredOrgs)
	}
	return err
}
