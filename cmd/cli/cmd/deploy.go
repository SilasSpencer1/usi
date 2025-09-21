package cmd

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"time"
	"usi/pkg/type/deployment"

	"platform-go-common/pkg/errors"

	"usi/pkg/type/environment"
	"usi/pkg/type/typeconst"

	cli "github.com/jawher/mow.cli"

	"usi/pkg/core"
	"usi/pkg/model/config"
)

const deploySpec = "[ -a=<key1=value1,key2=value2>... ] [ -d ] [ -e=<environment> ] [ -n=<name> ]  [ -r=<dir> ] [ -p=<properties> ] [ -s=<selector1[,selector2]> ] [ -t=<target> ] [-w] [ -v ] [ -l ] [ -x | --skip-post-conditions ] [ --skip-produces ] [ --clear-annotations ] [ --force ]"

func CmdDeploy(cmd *cli.Cmd) {
	command := "deploy"
	cmd.Spec = deploySpec
	opts := NewOpts(cmd)

	deployOpts := NewDeployOpts(opts)

	cmd.Action = func() {
		if deployOpts.wait != nil && *deployOpts.wait && deployOpts.logs != nil && *deployOpts.logs {
			PrintWarning("Will wait for deployment before outputting logs. " +
				"You may want to remove -w to see initialization errors")
		}
		opts.Normalize(command)
		opts.Validate(command)
		deployOpts.env = ToggleEnvironment(deployOpts.env, deployOpts.name)

		var environmentResource = &environment.Resource{}
		ValidateAndRetrieveEnvironment(command, deployOpts.env).Remarshal(environmentResource)

		if environmentResource.Selector.MatchesSelector(typeconst.AdditionalTestingEnvironmentSelector) && *deployOpts.selector == "" {
			HandleError(errors.WithCode(fmt.Sprintf("You must provide an optional selector (-s) when deploying to shared team environments"), errors.BadRequest), command)
		}

		var request = deployment.ClientDeployRequest{}
		if deployOpts.force != nil {
			request.Force = *deployOpts.force
		}
		properties, err := StrToConfiguration(deployOpts.props)
		HandleError(err, command)

		var outWriter, errWriter io.Writer = outOpt, errOpt
		outOpt.SetRules(deployOutputRules)
		errOpt.SetRules(deployOutputRules)

		if *deployOpts.verbose {
			outWriter = os.Stdout
			errWriter = os.Stderr
		}

		registryOpts, source := DecorateM5(deployOpts.name, deployOpts.env, deployOpts.m5Dir, core.DeployCmd, StrsToAnnotations(deployOpts.annotations), StrToSelector(deployOpts.selector, command), properties, deployOpts.target, outWriter, errWriter)
		request.ProviderOptions = registryOpts
		if *deployOpts.selector != "" {
			source.M5.EnsureDeclaration()
			source.M5.Declaration.OptionalSelector = deployOpts.selector
		}

		HandleError(source.M5.ReMarshal(&request), command)
		request.Environment = EnvFromSelectorName(*deployOpts.env)
		request.Requester = Requester()
		request.DryRun = *deployOpts.dryRun
		if !request.DryRun {
			PrintHeader("* Deploying via usi")
		} else {
			PrintHeader("* Deploying via usi (Dry Run)")
		}

		if deployOpts.skipProduces != nil && *deployOpts.skipProduces {
			keepProperties := make([]config.Property, 0, len(request.Configuration.Properties))
			producesKeys := make(map[string]bool, len(request.Declaration.Produces))
			for _, produces := range request.Declaration.Produces {
				producesKeys[produces.Key] = true
			}
			for _, property := range request.Configuration.Properties {
				if (property.Key != nil && producesKeys[*property.Key]) || (property.EnvKey != nil && producesKeys[*property.EnvKey]) {
					continue
				}
				keepProperties = append(keepProperties, property)
			}

			request.Declaration.Produces = nil
			request.Configuration.Properties = keepProperties
			PrintHeader("* Skipping produces declaration (bypass validation)")
		}

		if deployOpts.clearAnnotations != nil && *deployOpts.clearAnnotations {
			PrintHeader("* Clearing annotations using --clear-annotations flag")
			request.ClearAnnotations = true
		}

		if *deployOpts.verbose {
			PrintSectionWarning(source.Warnings)
		}

		deployStart := time.Now()
		deployResponse, err := Workspace(deployOpts.target, outWriter, errWriter, command).Deploy(request)
		HandleResolveError(command, err)
		if *deployOpts.dryRun {
			PrintYAML(deployResponse.Deployment, command)
		}

		if deployResponse.Deployment.MetaData.Annotations != nil && len(deployResponse.Deployment.MetaData.Annotations) > 0 && *deployOpts.verbose {
			PrintHeader("Annotations")
			PrintYAML(deployResponse.Deployment.MetaData.Annotations, command)
		}

		deploymentDuration := time.Since(deployStart)
		PrintHeader("Completed Deployment: %s (%s) in %vs", deployResponse.Deployment.Name, deployResponse.Deployment.UUID, roundFloat(deploymentDuration.Seconds(), 3))

		if deployOpts.skipPostConditions == nil || (deployOpts.skipPostConditions != nil && !*deployOpts.skipPostConditions) {
			if err := Workspace(deployOpts.target, outWriter, errWriter, command).Postconditions(*deployOpts.env, source, core.DeployCmd); err != nil {
				HandleError(err, command)
			}
		}

		fmt.Println("__________________________________________________________________")

		producedKeys, found := ExtractAndPrintProducedKValuePairs(deployResponse.Deployment.Data.Configuration, deployResponse.Deployment.Data.Declaration)

		if deployResponse.Deployment.Links != nil && len(deployResponse.Deployment.Links) > 0 {
			PrintLinks(deployResponse.Deployment.Links)
		}

		if !request.DryRun {
			PrintPortDocumentation(producedKeys, found, deployResponse.Deployment, command)
		}

		PrintIngressHostsIfAvailable(deployResponse.Deployment.Kubernetes)

		if deployResponse.Warnings != nil {
			for _, warning := range deployResponse.Warnings {
				PrintWarning(warning)
			}
		}
		HandleDeployWarning(deployResponse, command)

		remote := os.Getenv("REMOTE_DEPLOY")
		if remote == "" {
			remote = "false"
		}

		honeyCombMap := map[string]interface{}{
			"annotations":           deployOpts.annotations,
			"dryrun":                strconv.FormatBool(*deployOpts.dryRun),
			"environment":           deployOpts.env,
			"name":                  deployOpts.name,
			"selectors":             deployOpts.selector,
			"target":                deployOpts.target,
			"deployment_duration_s": time.Duration.Seconds(deploymentDuration),
			"result":                "success",
			"remote":                remote,
		}

		for script, timing := range source.ScriptTimes {
			honeyCombMap[script+"_script_duration_s"] = time.Duration.Seconds(timing)
		}
		snowflakeMap := map[string]interface{}{
			"service_name": deployOpts.name,
			"additional_info": "environment:" + *deployOpts.env + " dryrun:" + strconv.FormatBool(*deployOpts.dryRun) +
				" selectors:" + *deployOpts.selector + " target:" + *deployOpts.target +
				" deployment_duration_s:" + strconv.FormatFloat(time.Duration.Seconds(deploymentDuration), 'E', -1, 64),
			"environment": *deployOpts.env,
			"remote":      remote,
		}

		fmt.Println("") // Extra new line before waiting / adding the wait warning
		// Start waiting after fully completing the deployment
		if deployOpts.wait != nil && *deployOpts.wait {
			waitDur := WaitForDeployment(deployResponse.Deployment, *deployOpts.env, command)

			// add wait durations to reported metrics
			honeyCombMap["wait_duration_s"] = waitDur.Seconds()
			snowflakeMap["additional_info"] =
				snowflakeMap["additional_info"].(string) + fmt.Sprintf(" wait_duration_s:%f", waitDur.Seconds())
			snowflakeMap["wait_duration_s"] = waitDur.Seconds()
		} else {
			PrintWarning("Note: Due to delayed application startup time of 10+ minutes in some cases," +
				" your application service URL may not be accessible immediately post usi deploy.")
		}

		Reporter.SendHoneycombEvent(command, honeyCombMap)
		Reporter.SendSnowflakeEvent(command, snowflakeMap)

		// report metrics above ^^^ because dev will Exit out of tail using ctrl-c
		if deployOpts.logs != nil && *deployOpts.logs {
			PrintWarningIfKubectlVersionBehind()

			TailDeploymentLogs(deployResponse.Deployment, *deployOpts.env, command, "deployment")
		}
	}
}

var deployOutputRules = map[string]OutputRule{
	"running_script": {
		Hook:    "Running script: ",
		Pattern: regexp.MustCompile("Running script: (.+)$"),
		Type:    Step,
		Prefixes: map[string]string{
			"/build.sh":   "BUILDING artifact ",
			"/publish.sh": "PUBLISHING artifact to Artifactory ",
			"/digest.sh":  "RETRIEVING Docker DIGEST ",
			"/sync.sh":    "SYNCING Static Assets ",
		},
		Msg: "via: %s",
	},
	"script_elapsed_time": {
		Hook:    "Script elapsed time: ",
		Pattern: regexp.MustCompile("Script elapsed time: (.+)$"),
		Type:    StepTiming,
		Msg:     "Elapsed: %vs",
	},
	"bazel_build": {
		Hook:    "bazel build --build_tag_filters= ",
		Pattern: regexp.MustCompile(`bazel build --build_tag_filters= (.+)$`),
		Type:    SubStep,
		Msg:     "  - Building your application via bazel [%s]",
	},
	"streaming_build_results": {
		Hook:    "Streaming build results to: ",
		Pattern: regexp.MustCompile(`Streaming build results to: (.+)$`),
		Type:    SubStepFollower,
		Msg:     "    Streaming build results to: %s",
	},
	"starting_bazel_server": {
		Hook:    "Starting local Bazel server and connecting to it",
		Pattern: nil,
		Type:    SubStep,
		Msg:     "  - Starting local Bazel server and connecting to it...",
	},
	"rsyncing_freemarker_templates": {
		Hook:       "RSyncing Freemarker Templates in",
		Pattern:    nil,
		Type:       SubStep,
		Msg:        "  - RSyncing Freemarker Templates",
		NeedTiming: true,
	},
	"elapsed_time": {
		Hook:    "INFO: Elapsed time: ",
		Pattern: regexp.MustCompile(`INFO: Elapsed time: (\d*\.?\d*[a-z]+)`),
		Type:    SubStepTiming,
		Msg:     "  Elapsed: %vs",
	},
	"updating_building_message_bundles": {
		Hook:    "Updating and Building Message Bundles...",
		Pattern: nil,
		Type:    SubStep,
		Msg:     "  - Updating and Building Message Bundles...",
	},
	"tarballing_site_static_bundles": {
		Hook:    "Building and tarballing usi-site-static bundles...",
		Pattern: nil,
		Type:    SubStep,
		Msg:     "  - Building and tarballing usi-site-static bundles...",
	},
	"running_cli": {
		Hook:    "Running command line: ",
		Pattern: regexp.MustCompile(`Running command line: (.+)$`),
		Type:    SubStep,
		Msg:     "  - Running command line: [%s]",
	},
	"running_webpack": {
		Hook:       "Running webpack",
		Pattern:    nil,
		Type:       SubStep,
		Msg:        "  - Executing webpack",
		NeedTiming: true,
	},
	"rsyncing_usi_site_static": {
		Hook:       "RSyncing usi-site-static packages",
		Pattern:    nil,
		Type:       SubStep,
		Msg:        "  - RSyncing usi-site-static packages",
		NeedTiming: true,
	},

	"requires_proxinator_deploy": {
		Hook:       "we require proxinator be running to sync content",
		Pattern:    nil,
		Type:       Msg,
		Msg:        "",
		NeedTiming: false,
	},

	"requires_service_deploy": {
		Hook:       "is not deployed to",
		Pattern:    nil,
		Type:       Msg,
		Msg:        "",
		NeedTiming: false,
	},

	"usi_deploy_yn": {
		Hook:       "Would you like usi to deploy it for you?",
		Pattern:    nil,
		Type:       Msg,
		Msg:        "",
		NeedTiming: false,
	},
}
