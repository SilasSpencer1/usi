package cmd

import (
	"os"

	cli "github.com/jawher/mow.cli"

	usi "usi/pkg/client"

	"usi/pkg/core"
)

func CmdCreateEnvironment(app *cli.Cmd) {
	command := "create environment"
	app.Spec = " ENVIRONMENT [ -s=<selector1[,selector2]> ] [ --static ] [ -r=<region> ] "
	environment := app.StringArg("ENVIRONMENT", "", "environment name")
	Reporter.UsedOption("environment", environment)
	opts := NewOpts(app)
	region := opts.RegionOpt()
	sel := opts.SelectorOpt()
	static := opts.StaticOpt()
	app.Action = func() {
		opts.Normalize(command)
		opts.Validate(command)
		var resource usi.Resource
		var nsRequest *core.Request

		selector := StrToSelector(sel, command)
		selectors := map[string]interface{}{core.DevelopmentScope: nil}
		selectors[*region] = nil
		if *static {
			selectors["static"] = nil
		}
		for s := range selectors {
			selector.Selectors = append(selector.Selectors, s)
		}

		HandleError(Workspace(nil, os.Stdout, os.Stderr, command).CreateEnvironment(Requester(), *environment, core.DevelopmentScope, *selector, nsRequest, &resource),
			command)
		PrintResource(resource, command)
		Reporter.SendHoneycombEvent(command, map[string]interface{}{
			"environment": environment,
			"scope":       core.DevelopmentScope,
			"namespace":   nil,
			"selectors":   sel,
			"result":      "success",
		})
		Reporter.SendSnowflakeEvent(command, map[string]interface{}{
			"additional_info": "environment:" + *environment + " scope:" + core.DevelopmentScope +
				" selectors:" + *sel,
			"environment": *environment,
		})
	}
}
