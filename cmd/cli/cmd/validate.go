package cmd

import (
	"fmt"
	"os"

	cli "github.com/jawher/mow.cli"
	"usi/pkg/core"
	"usi/pkg/registry"
)

func CmdValidate(cmd *cli.Cmd) {
	command := "validate"
	cmd.Spec = "[ -e=<environment> ] [ -n=<name> ] [ -r=<dir> ] [ -t=<target> ] "
	opts := NewOpts(cmd)
	environment := opts.EnvironmentOpt()
	name := opts.NameOpt()
	m5Dir := opts.DirOpt()
	target := opts.TargetOpt(command)
	cmd.Action = func() {
		opts.Normalize(command)
		opts.Validate(command)
		environment = ToggleEnvironment(environment, name)
		_ = ValidateAndRetrieveEnvironment(command, environment)

		_, source := DecorateM5(name, environment, m5Dir, core.ValidateCmd, StrsToAnnotations(nil), nil, nil, target, os.Stdout, os.Stderr) // TODO: handle errors inside DecM5
		PrintHeader("Validating")
		var request registry.ValidateRequest
		HandleError(source.M5.ReMarshal(&request), command)
		request.Environment = EnvFromSelectorName(*environment)
		fmt.Fprintf(os.Stdout, "validating [%s] against [%s]\n", *source.Path(), *core.JoinNameAndSelector(*request.Environment.Name, request.Environment.Selector))
		request.Requester = Requester()
		HandleResolveError(command, Workspace(target, os.Stdout, os.Stderr, command).Validate(request))
		fmt.Fprintf(os.Stdout, "m5.yaml validation succeeded! (%s)\n", *source.Path())
		PrintFooter()
		Reporter.SendHoneycombEvent(command, map[string]interface{}{
			"environment":  environment,
			"service_name": name,
			"target":       target,
			"result":       "success",
		})
		Reporter.SendSnowflakeEvent(command, map[string]interface{}{
			"service_name":    name,
			"environment":     *environment,
			"additional_info": "target:" + *target + " environment:" + *environment,
		})
	}
}
