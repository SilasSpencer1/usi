package cmd

import (
	"fmt"
	"os"
	"path"

	cli "github.com/jawher/mow.cli"

	usi "usi/pkg/client"

	"usi/pkg/core"
	"usi/pkg/registry"
)

func CmdResolve(cmd *cli.Cmd) {
	command := "resolve"
	dir, err := os.Getwd()
	if err != nil {
		HandleError(err, command)
	}
	cmd.Spec = "[ -a=<key1=value1,key2=value2> ] [ -e=<environment> ] [ -k ] [ -n=<name> ] [ -r=<dir> ] [ -o ] [ -p=<properties> ] [ -s=<selector1[,selector2]> ] [ -t=<target> ] [ --diff ] [ -f=<filename> ]"
	opts := NewOpts(cmd)
	annotations := opts.AnnotationsOpt()
	environment := opts.EnvironmentOpt()
	name := opts.NameOpt()
	m5Dir := opts.DirOpt()
	selector := opts.SelectorOpt()
	filename := opts.FilenameOpt()
	omitDotEnvFile := cmd.BoolOpt("o omitfile", false, "omit "+usi.ConfigFilename+" file")
	props := opts.PropsOpt()
	shellEscape := opts.ShellEscapeOpt()
	target := opts.TargetOpt(command)
	diff := opts.DiffOpt()
	cmd.Action = func() {
		opts.Normalize(command)
		opts.Validate(command)
		environment = ToggleEnvironment(environment, name)
		_ = ValidateAndRetrieveEnvironment(command, environment)

		properties, err := StrToConfiguration(props)
		HandleError(err, command)
		_, source := DecorateM5(name, environment, m5Dir, core.ResolveCmd, StrsToAnnotations(annotations),
			StrToSelector(selector, command), properties, target, os.Stdout, os.Stderr)
		if *selector != "" {
			source.M5.EnsureDeclaration()
			source.M5.Declaration.OptionalSelector = selector
		}
		PrintHeader("Resolving")
		var request registry.ResolveRequest
		HandleError(source.M5.ReMarshal(&request), command)
		request.Environment = EnvFromSelectorName(*environment)
		HandleError(err, command)
		fmt.Fprintf(os.Stdout, "resolving [%s] to [%s]\n", *source.Path(), *core.JoinNameAndSelector(*request.Environment.Name, request.Environment.Selector))
		request.Requester = Requester()
		configuration, err := Workspace(target, os.Stdout, os.Stderr, command).Resolve(request)
		HandleResolveError(command, err)
		switch {
		case diff != nil && *diff:
			deployment := GetServiceDeployment(command, *environment, *name, StrToSelector(selector, command))
			HandleDiffInDotEnv(deployment, configuration, environment)
		case omitDotEnvFile != nil && *omitDotEnvFile:
			PrintHeader("Results")
			PrintYAML(configuration, command)
		default:
			WriteDotEnv(path.Join(dir, *filename), false, *configuration, *shellEscape, command)
		}
		PrintFooter()
		Reporter.SendHoneycombEvent(command, map[string]interface{}{
			"name":        name,
			"environment": environment,
			"target":      target,
			"result":      "success",
		})
		Reporter.SendSnowflakeEvent(command, map[string]interface{}{
			"service_name":    name,
			"additional_info": "environment:" + *environment + " target:" + *target,
			"environment":     *environment,
		})
	}
}
