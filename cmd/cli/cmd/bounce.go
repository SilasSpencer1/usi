package cmd

import (
	"fmt"
	"os"

	cli "github.com/jawher/mow.cli"

	"usi/pkg/core"
	"usi/pkg/type/deployment"
)

func CmdBounce(cmd *cli.Cmd) {
	command := "bounce"
	cmd.Spec = "[-n=<service name>] [ -s=<selector> ] [ -e=<environment> ] [-r]"
	opts := NewOpts(cmd)
	opts.NameOpt()
	opts.EnvironmentOpt()
	opts.SelectorOpt()
	resolveBool := cmd.BoolOpt("r resolve", false, "preform a configuration resolution prior to bounce")
	cmd.Action = func() {
		opts.Normalize(command)
		opts.Validate(command)

		environmentName := ToggleEnvironment(opts.Environment, opts.Name)
		serviceName, serviceSelector, err := core.ParseSelectorName(*opts.Name)
		mergedServiceSelectors := core.MergeSelectors(serviceSelector, StrToSelector(opts.Selector, command))
		HandleError(err, command)
		AssertDeployment(command, *environmentName, core.JoinNameAndSelector(serviceName, mergedServiceSelectors))
		fmt.Fprintf(os.Stdout, "Bounce for the  %s deployment in environment %s is in progress...", serviceName, *environmentName)
		fmt.Println("\n__________________________________________________________________")

		deploymentRequest := core.RequestFromTypeAndNameAndSelector(
			deployment.TypeName, *deployment.Name(*environmentName, serviceName, mergedServiceSelectors), nil)
		_, err = Workspace(nil, os.Stdout, os.Stderr, command).Bounce(Requester(), deploymentRequest, *resolveBool)

		if !*resolveBool {
			PrintHeader("%s deployment was successfully bounced.", serviceName)
			Reporter.SendHoneycombEvent(command, map[string]interface{}{
				"result": "success",
			})
			Reporter.SendSnowflakeEvent(command, map[string]interface{}{})
		} else {
			PrintHeader("%s deployment was successfully bounced and resolved.", serviceName)
			Reporter.SendHoneycombEvent("bounce-resolve", map[string]interface{}{
				"environment":  *environmentName,
				"service_name": serviceName,
				"result":       "success",
			})
			Reporter.SendSnowflakeEvent("bounce-resolve", map[string]interface{}{
				"environment":  *environmentName,
				"service_name": serviceName,
				"result":       "success",
			})
		}
	}
}
