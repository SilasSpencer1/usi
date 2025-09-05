package cmd

import (
	"fmt"
	"strings"

	cli "github.com/jawher/mow.cli"
	usi "usi/pkg/client"
	_ "usi/pkg/core"
)

type Opts struct {
	cmd                *cli.Cmd
	Annotations        *[]string
	Dir                *string
	Environment        *string
	Filter             *string
	Includes           *string
	Key                *string
	Max                *int
	Local              *bool
	Name               *string
	Props              *string
	Region             *string
	Quiet              *bool
	Selector           *string
	Target             *string
	UUID               *string
	Verbose            *bool
	Wait               *bool
	Logs               *bool
	SkipPostConditions *bool
	SkipProduces       *bool
	ClearAnnotations   *bool
	ShellEscape        *bool
	Force              *bool
	Diff               *bool
	Static             *bool
	Branch             *string
	Async              *bool
	DryRun             *bool
	Remote             *bool
	Kinds              *[]string
	Filename           *string
	OwnerTeam          *string
	OwnerUser          *string
	DeployerTeam       *string
	DeployerUser       *string
	Cluster            *string
	Global             *bool
}

func NewOpts(cmd *cli.Cmd) *Opts {
	return &Opts{
		cmd: cmd,
	}
}

func (o *Opts) AnnotationsOpt() *[]string {
	o.Annotations = o.cmd.StringsOpt("a annotations", nil, "additional annotations to add to a deployment")
	Reporter.UsedOption("annotations", o.Annotations)
	return o.Annotations
}

func (o *Opts) DirOpt() *string {
	o.Dir = o.cmd.StringOpt("r dir", "", "workspace dir of m5 file (e.g. for resolve)")
	Reporter.UsedOption("dir", o.Dir)
	return o.Dir
}

func (o *Opts) EnvironmentOpt() *string {
	o.Environment = o.cmd.StringOpt("e environment", *DefaultEnvironment(), "specify environment to work with")
	Reporter.UsedOption("environment", o.Environment)
	return o.Environment
}

func (o *Opts) FilterOpt() *string {
	o.Filter = o.cmd.StringOpt("f filter", "", "specify substring to filter resources with")
	Reporter.UsedOption("filter", o.Filter)
	return o.Filter
}

func (o *Opts) IncludesOpt() *string {
	o.Includes = o.cmd.StringOpt("i include", "", "comma seperated list of options. values are (active, data, metadata)")
	Reporter.UsedOption("include", o.Includes)
	return o.Includes
}

func (o *Opts) KeyOpt() *string {
	o.Key = o.cmd.StringOpt("k key", "", "specify configuration \"key\" to work with")
	Reporter.UsedOption("key", o.Key)
	return o.Key
}

func (o *Opts) MaxOpt() *int {
	o.Max = o.cmd.IntOpt("m max", 0, "max records to return")
	Reporter.UsedOption("max", o.Max)
	return o.Max
}

func (o *Opts) LocalOpt() *bool {
	o.Local = o.cmd.BoolOpt("l local", false, "Return resources only in the local environment")
	Reporter.UsedOption("local", o.Local)
	return o.Local
}

func (o *Opts) NameOpt() *string {
	o.Name = o.cmd.StringOpt("n name", "", "specify service name[.selector] to work with")
	Reporter.UsedOption("name", o.Name)
	return o.Name
}

func (o *Opts) PropsOpt() *string {
	o.Props = o.cmd.StringOpt("p properties", "", "dynamically specify env key=value,key=value pairs to include")
	Reporter.UsedOption("properties", o.Props)
	return o.Props
}

func (o *Opts) RegionOpt() *string {
	o.Region = o.cmd.StringOpt("r region", "", "specify region this resource should be associated with")
	Reporter.UsedOption("region", o.Region)
	return o.Region
}

func (o *Opts) QuietOpt() *bool {
	o.Quiet = o.cmd.BoolOpt("q quiet", false, "suppress extraneous text")
	Reporter.UsedOption("quiet", o.Quiet)
	return o.Quiet
}

func (o *Opts) SelectorOpt() *string {
	o.Selector = o.cmd.StringOpt("s selector", "", "specify an alphanumeric selector to add")
	Reporter.UsedOption("selector", o.Selector)
	return o.Selector
}

func (o *Opts) TargetOpt(command string) *string {
	o.Target = o.cmd.StringOpt("t target", DefaultTarget(command), "specify m5 target to work with")
	Reporter.UsedOption("target", o.Target)
	return o.Target
}

func (o *Opts) UUIDOpt() *string {
	o.UUID = o.cmd.StringOpt("u uuid", "", "specify UUID to work with")
	Reporter.UsedOption("uuid", o.UUID)
	return o.UUID
}

func (o *Opts) VerboseOpt() *bool {
	o.Verbose = o.cmd.BoolOpt("v verbose", false, "enable verbose logging")
	Reporter.UsedOption("verbose", o.Verbose)
	return o.Verbose
}

func (o *Opts) WaitOpt() *bool {
	o.Wait = o.cmd.BoolOpt("w wait", false, "wait for the deployment to become available after deploying. Limited to user namespace deploys and requires Rancher CLI.")
	Reporter.UsedOption("wait", o.Wait)
	return o.Wait
}

func (o *Opts) LogsOpt() *bool {
	o.Logs = o.cmd.BoolOpt("l logs", false, "tail the deployment's logs after deploying. Limited to user namespace deploys and requires Rancher CLI.")
	Reporter.UsedOption("logs", o.Logs)
	return o.Logs
}

func (o *Opts) SkipPostConditionsOpt() *bool {
	o.SkipPostConditions = o.cmd.BoolOpt("x skip-post-conditions", false, "skip post-condition scripts after deploying")
	Reporter.UsedOption("skip_post_conditions", o.SkipPostConditions)
	return o.SkipPostConditions
}

func (o *Opts) SkipProducesOpt() *bool {
	o.SkipProduces = o.cmd.BoolOpt("skip-produces", false, "[Not recommended] skip producing keys not explicitly defined under the target")
	Reporter.UsedOption("skip_produces", o.SkipProduces)
	return o.SkipProduces
}

func (o *Opts) ClearAnnotationsOpt() *bool {
	o.ClearAnnotations = o.cmd.BoolOpt("clear-annotations", false, "clear all annotations from a deployment")
	Reporter.UsedOption("clear_annotations", o.ClearAnnotations)
	return o.ClearAnnotations
}

func (o *Opts) ShellEscapeOpt() *bool {
	o.ShellEscape = o.cmd.BoolOpt("k shellescape", false, "shellescape values (convenient when values are intended to be used as command arguments)")
	Reporter.UsedOption("shellescape", o.ShellEscape)
	return o.ShellEscape
}

func (o *Opts) ForceOpt() *bool {
	o.Force = o.cmd.BoolOpt("force", false, "skip undeploy validation; requires client authentication")
	Reporter.UsedOption("force", o.Force)
	return o.Force
}

func (o *Opts) DiffOpt() *bool {
	o.Diff = o.cmd.BoolOpt("diff", false, "show the difference between what is the service currently resolving to and what would the service resolve to if redeployed")
	Reporter.UsedOption("diff", o.Diff)
	return o.Diff
}

func (o *Opts) StaticOpt() *bool {
	o.Static = o.cmd.BoolOpt("S static", true, "is this a static environment?")
	Reporter.UsedOption("static", o.Static)
	return o.Static
}

func (o *Opts) BranchOpt() *string {
	o.Branch = o.cmd.StringOpt("b branch", "", "specify git branch for remote deploy")
	Reporter.UsedOption("branch", o.Branch)
	return o.Branch
}

func (o *Opts) AsyncOpt() *bool {
	o.Async = o.cmd.BoolOpt("async", false, "option to not tail the output of the remote deploy job")
	Reporter.UsedOption("async", o.Async)
	return o.Async
}

func (o *Opts) DryRunOpt() *bool {
	o.DryRun = o.cmd.BoolOpt("d dryrun", false, "do deployment dry run (returns data relevant to the deploy)")
	Reporter.UsedOption("dryrun", o.DryRun)
	return o.DryRun
}

func (o *Opts) RemoteOpt() *bool {
	o.Remote = o.cmd.BoolOpt("r remote", false, "specify if the command references a remote job")
	Reporter.UsedOption("remote", o.Remote)
	return o.Remote
}

func (o *Opts) KindsOpt() *[]string {
	o.Kinds = o.cmd.StringsOpt("k kinds", nil, "specify the Kubernetes resource kinds to filter by, e.g. -k=Service -k=Deployment. All resources are returned if not specified.")
	Reporter.UsedOption("kinds", o.Kinds)
	return o.Kinds
}

func (o *Opts) Normalize(command string) {
	o.normalizeName()
	o.normalizeEnvironment()
	o.normalizeSelector(command)
	o.normalizeRegion()
}

func (o *Opts) Validate(command string) {
	o.validateName(command)
	o.validateSelector(command)
	o.validateRegion(command)
	o.validateEnvironment(command)
}

func (o *Opts) normalizeName() {
	if o.Name == nil || *o.Name == "" {
		return
	}
	*o.Name = core.NormalizeSelectorName(*o.Name)
}

func (o *Opts) normalizeEnvironment() {
	if o.Environment == nil || *o.Environment == "" {
		return
	}
	*o.Environment = core.NormalizeSelectorName(*o.Environment)
}

func (o *Opts) normalizeSelector(command string) {
	if o.Selector == nil || *o.Selector == "" {
		return
	}

	selector := StrToSelector(o.Selector, command)
	selector.Normalize()
	*o.Selector = strings.Join(selector.Selectors, ",")
}

func (o *Opts) normalizeRegion() {
	if o.Region == nil || *o.Region == "" {
		return
	}
	*o.Region = *core.NormalizeName(*o.Region)
}

func (o *Opts) validateEnvironment(command string) {
	if o.Environment == nil || *o.Environment == "" {
		return
	}
	_, envSelector, _ := core.ParseSelectorName(*o.Environment)
	err := envSelector.Validate()
	HandleError(err, command)
}

func (o *Opts) validateName(command string) {
	if o.Name == nil || *o.Name == "" {
		return
	}
	_, serviceSelector, _ := core.ParseSelectorName(*o.Name)
	err := serviceSelector.Validate()
	HandleError(err, command)
}

func (o *Opts) validateSelector(command string) {
	if o.Selector == nil || *o.Selector == "" {
		return
	}
	selector := StrToSelector(o.Selector, command)
	err := selector.Validate()
	HandleError(err, command)
}

func (o *Opts) validateRegion(command string) {
	if o.Region == nil || *o.Region == "" {
		return
	}
	if *o.Region != "na" && *o.Region != "eu" {
		HandleError(errors.WithCode(fmt.Sprintf("%s is not currently an accepted region.", *o.Region), errors.BadRequest),
			command)
	}
}

func (o *Opts) FilenameOpt() *string {
	o.Filename = o.cmd.StringOpt("f filename", mach5.ConfigFilename, "destination file name for resolved env vars")
	Reporter.UsedOption("filename", o.Filename)
	return o.Filename
}

func (o *Opts) OwnerTeamOpt() *string {
	o.OwnerTeam = o.cmd.StringOpt("owner-team", "", "filter deployments by service owner team name")
	Reporter.UsedOption("owner_team", o.OwnerTeam)
	return o.OwnerTeam
}
func (o *Opts) OwnerUserOpt() *string {
	o.OwnerUser = o.cmd.StringOpt("owner-user", "", "filter deployments by service owner user name")
	Reporter.UsedOption("owner_user", o.OwnerUser)
	return o.OwnerUser
}
func (o *Opts) DeployerTeamOpt() *string {
	o.DeployerTeam = o.cmd.StringOpt("deployer-team", "", "filter deployments by deployer team name")
	Reporter.UsedOption("deployer_team", o.DeployerTeam)
	return o.DeployerTeam
}
func (o *Opts) DeployerUserOpt() *string {
	o.DeployerUser = o.cmd.StringOpt("deployer-user", "", "filter deployments by deployer user name")
	Reporter.UsedOption("deployer_user", o.DeployerUser)
	return o.DeployerUser
}
func (o *Opts) ClusterOpt() *string {
	o.Cluster = o.cmd.StringOpt("cluster", "", "filter deployments by cluster name")
	Reporter.UsedOption("cluster", o.Cluster)
	return o.Cluster
}

func (o *Opts) GlobalOpt() *bool {
	o.Global = o.cmd.BoolOpt("global", false, "list across all environments (server-side filtering)")
	Reporter.UsedOption("global", o.Global)
	return o.Global
}

type DeployOpts struct {
	annotations        *[]string
	dryRun             *bool
	env                *string
	name               *string
	m5Dir              *string
	props              *string
	selector           *string
	target             *string
	verbose            *bool
	wait               *bool
	logs               *bool
	skipPostConditions *bool
	skipProduces       *bool
	clearAnnotations   *bool
	force              *bool
}

func NewDeployOpts(opts *Opts) DeployOpts {
	return DeployOpts{
		annotations:        opts.AnnotationsOpt(),
		dryRun:             opts.DryRunOpt(),
		env:                opts.EnvironmentOpt(),
		name:               opts.NameOpt(),
		m5Dir:              opts.DirOpt(),
		props:              opts.PropsOpt(),
		selector:           opts.SelectorOpt(),
		target:             opts.TargetOpt("deploy"),
		verbose:            opts.VerboseOpt(),
		wait:               opts.WaitOpt(),
		logs:               opts.LogsOpt(),
		skipPostConditions: opts.SkipPostConditionsOpt(),
		skipProduces:       opts.SkipProducesOpt(),
		clearAnnotations:   opts.ClearAnnotationsOpt(),
		force:              opts.ForceOpt(),
	}
}

func (o DeployOpts) String() string {
	res := ""

	if o.annotations != nil && len(*o.annotations) > 0 {
		fullAnnotationString := strings.Join(*o.annotations, ",")
		res = res + "-a=" + fullAnnotationString + " "
	}

	if o.dryRun != nil && *o.dryRun == true {
		res = res + "-d" + " "
	}

	if o.env != nil && *o.env != "" {
		res = res + "-e=" + *o.env + " "
	}

	if o.name != nil && *o.name != "" {
		res = res + "-n=" + *o.name + " "
	}

	if o.m5Dir != nil && *o.m5Dir != "" {
		res = res + "-r=" + *o.m5Dir + " "
	}

	if o.props != nil && *o.props != "" {
		res = res + "-p=" + *o.props + " "
	}

	if o.selector != nil && *o.selector != "" {
		res = res + "-s=" + *o.selector + " "
	}

	if o.target != nil && *o.target != "" {
		res = res + "-t=" + *o.target + " "
	}

	if o.verbose != nil && *o.verbose == true {
		res = res + "-v" + " "
	}

	if o.wait != nil && *o.wait == true {
		res = res + "-w" + " "
	}

	if o.logs != nil && *o.logs == true {
		res = res + "-l" + " "
	}

	if o.skipPostConditions != nil && *o.skipPostConditions == true {
		res = res + "--skip-post-conditions" + " "
	}

	if o.skipProduces != nil && *o.skipProduces == true {
		res = res + "--skip-produces" + " "
	}

	if o.clearAnnotations != nil && *o.clearAnnotations == true {
		res = res + "--clear-annotations" + " "
	}

	if o.force != nil && *o.force == true {
		res = res + "--force" + " "
	}

	return res
}
