package cmd

import (
	"fmt"
	"os"

	cli "github.com/jawher/mow.cli"

	"usi/pkg/registry"
)

func CmdDependents(app *cli.Cmd) {
	command := "dependents"
	app.Spec = " ( -u=<uuid> |  (-n=<name> [-e=<environment>] ) ) "
	opts := NewOpts(app)
	name := opts.NameOpt()
	environmentName := opts.EnvironmentOpt()
	uuid := opts.UUIDOpt()

	app.Action = func() {
		opts.Normalize(command)
		opts.Validate(command)
		environmentName = ToggleEnvironment(environmentName, name)
		var request registry.DependentsRequest
		if uuid != nil && *uuid != "" {
			request.Deployment.UUID = uuid
		} else {
			request.Deployment.Name = name
			if environmentName != nil && *environmentName != "" {
				Environment := EnvFromSelectorName(*environmentName)
				request.Environment = &Environment
			}
		}

		Dependents, err := Workspace(nil, os.Stdout, os.Stderr, command).Dependents(request)

		PrintHeader("Fetching Deployment's Dependent Deployments")
		if len(Dependents) > 0 {
			PrintDeployments(Dependents, command)
		} else {
			if err != nil {
				HandleError(err, command)
			}
			fmt.Fprintf(os.Stdout, "No Dependents found for the given deployment.")
		}
		Reporter.SendHoneycombEvent(command, map[string]interface{}{
			"name":        name,
			"environment": environmentName,
			"uuid":        uuid,
			"result":      "success",
		})
		Reporter.SendSnowflakeEvent(command, map[string]interface{}{
			"service_name":    name,
			"additional_info": "environment:" + *environmentName + " uuid:" + *uuid,
			"environment":     *environmentName,
		})
	}
}
