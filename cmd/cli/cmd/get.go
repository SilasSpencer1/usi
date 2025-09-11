package cmd

import (
	"fmt"
	"os"
	"strconv"

	"/usi/pkg/core"

	"/platform-go-common/pkg/errors"
	"/platform-go-common/pkg/usi"
	cli "github.com/jawher/mow.cli"

	"/usi/pkg/client"

	"/usi/pkg/model/config"
	"/usi/pkg/model/kubernetes"
	"/usi/pkg/registry"
	"/usi/pkg/type/cluster"
	"/usi/pkg/type/deployment"
	"/usi/pkg/type/environment"
	"/usi/pkg/type/namespace"
	"/usi/pkg/type/service"
	"/usi/pkg/type/team"
	"/usi/pkg/type/user"
)

func CmdGet(cmd *cli.Cmd) {
	cmd.Command("annotations", "display annotations", CmdGetAnnotations)
	//cmd.Command("changelog", "change history for a resource", CmdGetChangeLog)
	cmd.Command("clusters", "list clusters", CmdListClusters)
	cmd.Command("configuration", "display configuration", CmdGetConfiguration)
	cmd.Command("dependencies", "list a deployment's dependencies", CmdDependencies)
	cmd.Command("dependents", "list a deployment's dependent deployments", CmdDependents)
	cmd.Command("deployments", "list deployments", CmdListDeployments)
	cmd.Command("environment", "get an environment", CmdGetEnvironment)
	cmd.Command("environments", "list environments", CmdListEnvironments)
	cmd.Command("links", "gets the links related to the deployment", CmdGetLinks)
	cmd.Command("names", "list names of services within current workspace", CmdNames)
	cmd.Command("namespaces", "list namespaces", CmdListNamespaces)
	cmd.Command("registry", "get current registry", CmdGetRegistry)
	cmd.Command("resource", "display resource", CmdGetResource)
	cmd.Command("services", "list services", CmdListServices)
	cmd.Command("target", "get your target preference, if any", CmdGetTarget)
	cmd.Command("teams", "list teams", CmdListTeams)
	cmd.Command("types", "list resource types", CmdListTypes)
	cmd.Command("users", "list users", CmdListUsers)
	cmd.Command("remote", "get data on remote deployments", CmdGetRemote)
}

type Data struct {
	Kubernetes kubernetes.Config `json:"kubernetes"`
}

type Deployment struct {
	Name       string                        `json:"name"`
	Selectors  *[]string                     `json:"selectors,omitempty"`
	Properties *[]config.PropertyDeclaration `json:"map,omitempty"`
}

func CmdGetChangeLog(cmd *cli.Cmd) {
	//cmd.Spec = " ( -u=<uuid> |  (-n=<name> [-e=<environment>] ) ) -i -m "
	//name := NameOpt(cmd)
	//environmentName := EnvironmentOpt(cmd)
	//uuid := UUIDOpt(cmd)
	//includes := IncludesOpt(cmd)
	//max := MaxOpt(cmd)
	//cmd.Action = func() {
	//	var request registry.ChangeLogRequest
	//}
}

func CmdGetLinks(cmd *cli.Cmd) {
	cmd.Spec = " ( -u=<uuid> | (-n=<serviceName> [-e=<environment>]) )"
	opts := NewOpts(cmd)
	environmentName := opts.EnvironmentOpt()
	name := opts.NameOpt()
	uuid := opts.UUIDOpt()
	command := "get links"

	cmd.Action = func() {
		opts.Normalize(command)
		opts.Validate(command)
		environmentName = ToggleEnvironment(environmentName, name)
		var request registry.DeploymentRequest
		if uuid != nil && *uuid != "" {
			request.Deployment.UUID = uuid
		} else {
			request.Deployment.Name = name
			if environmentName != nil && *environmentName != "" {
				Environment := EnvFromSelectorName(*environmentName)
				request.Environment = &Environment
			}
		}

		links, err := Workspace(nil, os.Stdout, os.Stderr, command).DeploymentLinks(request)
		if err != nil {
			HandleError(err, command)
		}

		if links != nil && len(links) > 0 {
			PrintLinks(links)
		} else {
			ColoredOutput.Yellow("Deployment doesn't have any links configured")
		}
	}
}

func CmdGetAnnotations(cmd *cli.Cmd) {
	cmd.Spec = " ( -u=<uuid> | (-n=<serviceName> [-e=<environment>]) ) [ -k=<key>] [ -q=<quiet> ]"
	opts := NewOpts(cmd)
	environmentName := opts.EnvironmentOpt()
	key := opts.KeyOpt()
	name := opts.NameOpt()
	uuid := opts.UUIDOpt()
	quiet := opts.QuietOpt()
	command := "get annotations"

	cmd.Action = func() {
		opts.Normalize(command)
		opts.Validate(command)

		if *key != "" || *quiet {
			Quiet = true
		} else {
			Quiet = *quiet
		}

		environmentName = ToggleEnvironment(environmentName, name)
		_ = ValidateAndRetrieveEnvironment(command, environmentName)

		if *uuid != "" {
			var resource client.Resource
			HandleError(Workspace(nil, os.Stdout, os.Stderr, "get annotations").FromUUID(*uuid, &resource),
				command)
			switch *resource.TypeName {
			case deployment.TypeName:
				var d deployment.Resource
				HandleError(resource.Remarshal(&d), command)
				PrintHeader("Annotations")
				PrintYAML(d.MetaData.Annotations, command)
			case environment.TypeName:
				var e environment.Resource
				HandleError(resource.Remarshal(&e), command)
				PrintHeader("Annotations")
				PrintYAML(e.MetaData.Annotations, command)
			default:
				HandleError(errors.WithCode(fmt.Sprintf("Unsupported type: (%s)", *resource.TypeName), errors.NotImplemented),
					command)
			}
		} else if *name != "" {
			AssertDeployment(command, *environmentName, name)
			deployment := GetServiceDeployment(command, *environmentName, *name, nil)
			if deployment == nil {
				HandleError(errors.WithCode(fmt.Sprintf("Unable to find deployment (%s) in environment {%s}", *name, *environmentName), errors.NotFound),
					command)
			} else if deployment.MetaData.Annotations == nil {
				HandleError(errors.WithCode("deployment has no annotations", errors.NotFound), command)
			}

			if *key != "" {
				fmt.Print(deployment.MetaData.Annotations[*key])
			} else {
				PrintHeader("Annotations")
				PrintYAML(deployment.MetaData.Annotations, command)
			}
		} else {
			HandleError(errors.WithCode("must pass -u or -n", errors.BadRequest), command)
		}
		Reporter.SendHoneycombEvent("get annotations", map[string]interface{}{
			"environment":  environmentName,
			"key":          key,
			"service_name": name,
			"uuid":         uuid,
			"quiet":        strconv.FormatBool(*quiet),
			"result":       "success",
		})
		Reporter.SendSnowflakeEvent("get", map[string]interface{}{
			"secondary_command_get_set": "annotations",
			"service_name":              *name,
			"additional_info":           " key:" + *key + " uuid:" + *uuid + " quiet:" + strconv.FormatBool(*quiet),
			"environment":               *environmentName,
		})
	}
}

func CmdGetConfiguration(cmd *cli.Cmd) {
	cmd.Spec = " ( -u=<uuid> | (-n=<serviceName> [-e=<environment>]) ) [ -k=<key>] [ -q=<quiet> ] [ -s=<selector> ]"
	opts := NewOpts(cmd)
	environmentName := opts.EnvironmentOpt()
	key := opts.KeyOpt()
	name := opts.NameOpt()
	uuid := opts.UUIDOpt()
	quiet := opts.QuietOpt()
	selectorString := opts.SelectorOpt()
	command := "get configuration"

	cmd.Action = func() {
		opts.Normalize(command)
		opts.Validate(command)

		if *key != "" {
			Quiet = true
		} else {
			Quiet = *quiet
		}

		environmentName = ToggleEnvironment(environmentName, name)
		_ = ValidateAndRetrieveEnvironment(command, environmentName)

		if *uuid != "" {
			var resource client.Resource
			HandleError(Workspace(nil, os.Stdout, os.Stderr, command).FromUUID(*uuid, &resource), command)
			switch *resource.TypeName {
			case deployment.TypeName:
				var d deployment.Resource
				HandleError(resource.Remarshal(&d), command)
				producedKeys, found := ExtractAndPrintProducedKValuePairs(d.Configuration, d.Declaration)
				PrintLinks(d.Links)
				PrintPortDocumentation(producedKeys, found, &d, command)
				PrintIngressHostsIfAvailable(d.Kubernetes)

			case environment.TypeName:
				var e environment.Resource
				HandleError(resource.Remarshal(&e), command)
				ExtractAndPrintProducedKValuePairs(e.Configuration, e.Declaration)
			default:
				HandleError(errors.WithCode(fmt.Sprintf("Unsupported type: (%s)", *resource.TypeName), errors.NotImplemented),
					command)
			}
		} else if *name != "" {
			selector := core.ParseSelector(*selectorString)
			serviceName, serviceSelector := core.ParseSelectorNameAndAddCliSelector(*name, *selectorString) // already normalizes
			normName := core.JoinNameAndSelector(serviceName, serviceSelector)
			AssertDeployment(command, *environmentName, normName)
			deployment := GetServiceDeployment(command, *environmentName, *name, selector)
			if deployment != nil {
				conf, err := Workspace(nil, os.Stdout, os.Stderr, command).ClearTextConfiguration(core.RequestFromUUID(deployment.UUID))
				HandleError(err, command)
				if conf != nil {
					if *key != "" {
						PrintKey(*conf, *key, command)
					} else {
						producedKeys, found := ExtractAndPrintProducedKValuePairs(conf, deployment.Declaration)
						PrintLinks(deployment.Links)
						PrintPortDocumentation(producedKeys, found, deployment, command)
						PrintIngressHostsIfAvailable(deployment.Kubernetes)
					}
				} else {
					HandleError(errors.WithCode("deployment has no configuration", errors.NotFound), command)
				}
			} else {
				HandleError(errors.WithCode(errors.InvalidServiceDeploymentErrorMessage, errors.NotFound), command)
			}
		} else {
			HandleError(errors.WithCode("must pass -u or -n", errors.BadRequest), command)
		}
		Reporter.SendHoneycombEvent("get configuration", map[string]interface{}{
			"environment":  environmentName,
			"key":          key,
			"service_name": name,
			"uuid":         uuid,
			"quiet":        strconv.FormatBool(*quiet),
			"result":       "success",
		})
		Reporter.SendSnowflakeEvent("get", map[string]interface{}{
			"secondary_command_get_set": "configuration",
			"service_name":              *name,
			"additional_info":           " key:" + *key + " uuid:" + *uuid + " quiet:" + strconv.FormatBool(*quiet),
			"environment":               *environmentName,
		})
	}
}

func CmdGetEnvironment(cmd *cli.Cmd) {
	command := "get environment"
	opts := NewOpts(cmd)
	environmentName := opts.EnvironmentOpt()
	cmd.Action = func() {

		PrintEnvironment(*environmentName, command)
		Reporter.SendHoneycombEvent(command, map[string]interface{}{})
		Reporter.SendSnowflakeEvent("get", map[string]interface{}{
			"secondary_command_get_set": "environment",
		})
	}
}

func CmdGetRegistry(cmd *cli.Cmd) {
	command := "get registry"
	cmd.Action = func() {
		PrintHeader("Registry: %s", usi.GetOrDefault("unset", "registry", "url"))
		Reporter.SendHoneycombEvent(command, map[string]interface{}{
			"option_registry": usi.GetOrDefault("unset", "registry", "url"),
		})
		Reporter.SendSnowflakeEvent("get", map[string]interface{}{
			"secondary_command_get_set": "registry",
			"option_registry":           usi.GetOrDefault("unset", "registry", "url"),
		})
	}
}

func CmdGetResource(cmd *cli.Cmd) {
	command := "get resource"
	opts := NewOpts(cmd)
	uuid := opts.UUIDOpt()
	cmd.Action = func() {
		var resource client.Resource
		HandleError(Workspace(nil, os.Stdout, os.Stderr, command).FromUUID(*uuid, &resource), command)
		PrintResource(resource, command)
		Reporter.SendHoneycombEvent("get resource", map[string]interface{}{})
		Reporter.SendSnowflakeEvent("get", map[string]interface{}{
			"secondary_command_get_set": "resource",
		})
	}
}

func CmdGetTarget(cmd *cli.Cmd) {
	command := "get target"
	cmd.Action = func() {
		PrintHeader("Active target: %s", usi.GetOrDefault("unset", "target"))
		Reporter.SendHoneycombEvent(command, map[string]interface{}{
			"target": usi.GetOrDefault("unset", "target"),
		})
		Reporter.SendSnowflakeEvent("get", map[string]interface{}{
			"secondary_command_get_set": "target",
			"additional_info":           "target:" + usi.GetOrDefault("unset", "target"),
		})
	}
}

func CmdListClusters(cmd *cli.Cmd) {
	command := "get clusters"
	cmd.Action = ListMetaDataNames(command, cluster.TypeName, true)
	Reporter.SendHoneycombEvent(command, map[string]interface{}{})
	Reporter.SendSnowflakeEvent("get", map[string]interface{}{
		"secondary_command_get_set": "clusters",
	})
}

func CmdListDeployments(cmd *cli.Cmd) {
	command := "get deployments"
	cmd.Spec = "[ -e=<environment> ] [ -n=<service> [ -s=<selector> ] ] [ -f=<filter> [ -l ] ] [ -l ] [ --global ] [ --owner-team=<name> ] [ --owner-user=<name> ] [ --deployer-team=<name> ] [ --deployer-user=<name> ] [ --cluster=<name> ]"
	opts := NewOpts(cmd)
	environmentName := opts.EnvironmentOpt()
	name := opts.NameOpt()
	selectorString := opts.SelectorOpt()
	filter := opts.FilterOpt()
	local := opts.LocalOpt()
	ownerTeam := opts.OwnerTeamOpt()
	ownerUser := opts.OwnerUserOpt()
	deployerTeam := opts.DeployerTeamOpt()
	deployerUser := opts.DeployerUserOpt()
	cluster := opts.ClusterOpt()
	global := opts.GlobalOpt()
	cmd.Action = func() {
		opts.Normalize(command)
		opts.Validate(command)

		if global != nil && *global {
			PrintHeader("Deployments (global)")
			ot, on := "", ""
			dt, dn := "", ""
			if *ownerTeam != "" && *ownerUser != "" {
				HandleError(errors.WithCode("--owner-team and --owner-user are mutually exclusive", errors.BadRequest), command)
			}
			if *deployerTeam != "" && *deployerUser != "" {
				HandleError(errors.WithCode("--deployer-team and --deployer-user are mutually exclusive", errors.BadRequest), command)
			}
			if *ownerTeam != "" {
				ot, on = "team", *ownerTeam
			} else if *ownerUser != "" {
				ot, on = "user", *ownerUser
			}
			if *deployerTeam != "" {
				dt, dn = "team", *deployerTeam
			} else if *deployerUser != "" {
				dt, dn = "user", *deployerUser
			}
			selectors := ""
			if selectorString != nil {
				selectors = *selectorString
			}
			cli := Client()
			deployments, err := cli.DeploymentsFiltered(ot, on, dt, dn, *cluster, selectors)
			HandleError(err, command)
			PrintDeployments(deployments, command)
			PrintFooter()
			return
		}

		environmentName = ToggleEnvironment(environmentName, name)
		PrintHeader("Deployments: %s", *environmentName)
		if *name != "" {
			if *local {
				PrintReferenceDeployments(GetLocalDeployments(command, *environmentName))
			} else {
				optionalSelector := core.ParseSelector(*selectorString)
				PrintServiceDeployment(*environmentName, *name, optionalSelector, command)
			}

		} else if *filter != "" {
			if *local {
				PrintReferenceDeployments(FilterReferencedDeployments(*filter, GetLocalDeployments(command, *environmentName)))
			} else {
				PrintDeployments(FilterDeployments(*filter, GetDeployments(command, *environmentName)), command)
			}
		} else if *ownerTeam != "" || *ownerUser != "" || *deployerTeam != "" || *deployerUser != "" || *cluster != "" {
			// Env-scoped client-side filtering
			deployments := GetDeployments(command, *environmentName)
			filtered := make([]deployment.Resource, 0, len(deployments))
			for _, d := range deployments {
				if *cluster != "" {
					if d.Cluster == nil || d.Cluster.Name != *cluster {
						continue
					}
				}
				if *ownerTeam != "" || *ownerUser != "" {
					so := d.ServiceOwner
					if *ownerTeam != "" && !(so.TypeName == "team" && so.Name == *ownerTeam) {
						continue
					}
					if *ownerUser != "" && !(so.TypeName == "user" && so.Name == *ownerUser) {
						continue
					}
				}
				if *deployerTeam != "" || *deployerUser != "" {
					if *deployerTeam != "" && !(d.Deployer.TypeName == "team" && d.Deployer.Name == *deployerTeam) {
						continue
					}
					if *deployerUser != "" && !(d.Deployer.TypeName == "user" && d.Deployer.Name == *deployerUser) {
						continue
					}
				}
				filtered = append(filtered, d)
			}
			PrintDeployments(filtered, command)
		} else {
			if *local {
				PrintReferenceDeployments(GetLocalDeployments(command, *environmentName))
			} else {
				PrintDeploymentsForEnvironment(*environmentName, command)
			}
		}
		PrintFooter()
		Reporter.SendHoneycombEvent(command, map[string]interface{}{
			"environment":  environmentName,
			"service_name": "",
			"filter":       filter,
		})
		Reporter.SendSnowflakeEvent("get", map[string]interface{}{
			"secondary_command_get_set": "deployments",
			"additional_info":           "environment:" + *environmentName + " filter:" + *filter,
			"service_name":              "",
		})
	}
}

func CmdListEnvironments(cmd *cli.Cmd) {
	command := "get environments"
	cmd.Spec = "[ -f=<filter> ] [ --owner-team=<name> ] [ --owner-user=<name> ] [ --cluster=<name> ] [ --global ]"
	opts := NewOpts(cmd)
	filter := opts.FilterOpt()
	ownerTeam := opts.OwnerTeamOpt()
	ownerUser := opts.OwnerUserOpt()
	cluster := opts.ClusterOpt()
	global := opts.GlobalOpt()
	cmd.Action = func() {
		opts.Normalize(command)
		opts.Validate(command)
		// client-side: if environment is specified, we already fetch environments globally today; just filter client-side
		if global != nil && *global {
			ot, on := "", ""
			if ownerTeam != nil && *ownerTeam != "" {
				ot, on = "team", *ownerTeam
			} else if ownerUser != nil && *ownerUser != "" {
				ot, on = "user", *ownerUser
			}
			cli := Client()
			envs, err := cli.ListEnvironmentsFiltered(ot, on, func() string {
				if cluster == nil {
					return ""
				}
				return *cluster
			}(), func() string {
				if filter == nil {
					return ""
				}
				return *filter
			}())
			HandleError(err, command)
			PrintEnvironments(envs, command)
			PrintFooter()
			return
		}
		// default: fetch and client-filter
		envs := GetEnvironments()
		filtered := FilterEnvironments(envs, ownerTeam, ownerUser, cluster)
		PrintEnvironments(filtered, command)
		PrintFooter()
	}
}

func CmdListNamespaces(cmd *cli.Cmd) {
	command := "get namespaces"
	cmd.Action = ListMetaDataNames(command, namespace.TypeName, true)
	Reporter.SendHoneycombEvent(command, map[string]interface{}{})
	Reporter.SendSnowflakeEvent("get", map[string]interface{}{
		"secondary_command_get_set": "namespaces",
	})
}

func CmdListServices(cmd *cli.Cmd) {
	command := "get services"
	cmd.Action = ListMetaDataNames(command, service.TypeName, false)
	Reporter.SendHoneycombEvent(command, map[string]interface{}{})
	Reporter.SendSnowflakeEvent("get", map[string]interface{}{
		"secondary_command_get_set": "services",
	})
}

func CmdListTeams(cmd *cli.Cmd) {
	command := "get teams"
	cmd.Action = ListMetaDataNames(command, team.TypeName, false)
	Reporter.SendHoneycombEvent(command, map[string]interface{}{})
	Reporter.SendSnowflakeEvent("get", map[string]interface{}{
		"secondary_command_get_set": "teams",
	})
}

func CmdListTypes(cmd *cli.Cmd) {
	command := "get types"
	cmd.Action = func() {
		typeNames, err := Workspace(nil, os.Stdout, os.Stderr, command).Types()
		HandleError(err, command)
		PrintHeader("Resource Types")
		for _, typeName := range typeNames {
			fmt.Println(typeName)
		}
		PrintFooter()
		Reporter.SendHoneycombEvent(command, map[string]interface{}{})
		Reporter.SendSnowflakeEvent("get", map[string]interface{}{
			"secondary_command_get_set": "types",
		})
	}
}

func CmdListUsers(cmd *cli.Cmd) {
	command := "get users"
	cmd.Action = ListMetaDataNames(command, user.TypeName, false)
	Reporter.SendHoneycombEvent(command, map[string]interface{}{})
	Reporter.SendSnowflakeEvent("get", map[string]interface{}{
		"secondary_command_get_set": "users",
	})
}
