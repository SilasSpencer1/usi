package cmd

import (
	"fmt"
	"os"

	"usi/pkg/registry"
	"usi/pkg/type/deployment"

	"platform-go-common/pkg/errors"
	cli "github.com/jawher/mow.cli"

	"usi/pkg/core"
)

func CmdUndeploy(cmd *cli.Cmd) {
	command := "undeploy"
	cmd.Spec = "[ -e=<environment> ] [ -n=<name> ] [ -s=<selector> ] [ --force ]"
	opts := NewOpts(cmd)
	environment := opts.EnvironmentOpt()
	name := opts.NameOpt()
	selectorString := opts.SelectorOpt()
	force := opts.ForceOpt()

	cmd.Action = func() {
		opts.Normalize(command)
		opts.Validate(command)
		if name != nil && *name != "" {
			environment = ToggleEnvironment(environment, name)
			_ = ValidateAndRetrieveEnvironment(command, environment)
			serviceName, serviceSelector := core.ParseSelectorNameAndAddCliSelector(*name, *selectorString) // already normalizes
			normName := core.JoinNameAndSelector(serviceName, serviceSelector)
			AssertDeployment(command, *environment, normName)
			PrintHeader("Removing Deployment")
			depReq := core.RequestFromTypeAndNameAndSelector(
				deployment.TypeName,
				*deployment.Name(*environment, serviceName, serviceSelector),
				nil)
			request := registry.UndeployRequest{Deployment: depReq, Requester: Requester()}
			if force != nil {
				request.Force = *force
			}
			undeployResponse, err := Workspace(nil, os.Stdout, os.Stderr, command).Undeploy(request)
			HandleResolveError("undeploy", err)
			PrintYAML(undeployResponse.Environment, command)
			HandleUndeployWarning(undeployResponse, command)
		} else if selectorString != nil && *selectorString != "" {
			_ = ValidateAndRetrieveEnvironment(command, environment)
			var request registry.CleanEnvironmentRequest
			request.Selector = core.ParseAndNormalizeSelector(*selectorString)
			request.Requester = Requester()
			request.Environment = EnvFromSelectorName(*environment)
			if force != nil {
				request.Force = *force
			}
			undeployedList, err := Workspace(nil, os.Stdout, os.Stderr, command).CleanEnvironment(request)
			if len(undeployedList) > 0 {
				ColoredOutput.Green(fmt.Sprintf("%d deployments were undeployed:", len(undeployedList)))
				PrintUndeployedList(undeployedList)
			} else {
				PrintHeader("Nothing to undeploy on %s environment with selector: %s", *environment, *selectorString)
			}
			if err != nil {
				HandleError(err, "undeploy", "undeploy deployments failed")
			}
		} else {
			HandleError(errors.WithCode(
				"service must be provides by name with the -n flag or by selector with the -s flag", errors.BadRequest),
				"undeploy")
		}

		PrintFooter()
		Reporter.SendHoneycombEvent(command, map[string]interface{}{
			"environment":  environment,
			"service_name": name,
			"result":       "success",
		})
		Reporter.SendSnowflakeEvent(command, map[string]interface{}{
			"service_name":    name,
			"environment":     *environment,
			"additional_info": "environment:" + *environment,
		})
	}
}
