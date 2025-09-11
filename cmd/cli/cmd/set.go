package cmd

import (
	"fmt"
	"os"
	"strings"

	cli "github.com/jawher/mow.cli"
	"usi/pkg/usi"

	"usi/pkg/client"

	"usi/pkg/type/environment"
)

func CmdSet(app *cli.Cmd) {
	app.Command("delegates", "set environment delegates", CmdSetDelegates)
	app.Command("environment", "set current environment", CmdSetEnvironment)
	app.Command("registry", "set registry url", CmdSetRegistryURL)
	app.Command("target", "set your default target", CmdSetTarget)
}

func CmdSetDelegates(app *cli.Cmd) {
	command := "set delegates"
	app.Spec = "DELEGATES [ -e=<environment> ]"
	description := "comma separated list of environments in (highest to lowest) priority order"
	delegates := app.StringArg("DELEGATES", "", description)
	Reporter.UsedOption("delegates", delegates)
	environment := app.StringOpt("e environment", usi.GetRequired("environment"), "environment to discover")
	Reporter.UsedOption("environment", environment)
	app.Action = func() {
		env, err := Client().UpdateDelegates(Requester(), *environment, strings.Split(*delegates, ",")...)
		HandleError(err, command)
		PrintResource(*env, command)
		Reporter.SendHoneycombEvent(command, map[string]interface{}{
			"result": "success",
		})
		Reporter.SendSnowflakeEvent("set", map[string]interface{}{
			"secondary_command_get_set": "delegates",
		})
	}
}

func CmdSetEnvironment(cmd *cli.Cmd) {
	command := "set environment"
	cmd.Spec = "ENVIRONMENT"
	var env = cmd.StringArg("ENVIRONMENT", "", "environment")
	cmd.Action = func() {
		_ = ValidateAndRetrieveEnvironment(command, env)

		var resource client.Resource
		HandleError(Workspace(nil, os.Stdout, os.Stderr, command).FromTypeAndName(environment.TypeName, *env, &resource),
			command)
		HandleError(usi.Set(client.ClientConfigPath(), *env, "environment"), command)
		fmt.Printf("environment set to %s\n", *env)
		PrintResource(resource, command)
		Reporter.SendHoneycombEvent("set environment", map[string]interface{}{
			"environment": env,
			"result":      "success",
		})
		Reporter.SendSnowflakeEvent("set", map[string]interface{}{
			"secondary_command_get_set": "environment",
			"additional_info":           "environment:" + *env,
		})
	}
}

func CmdSetRegistryURL(cmd *cli.Cmd) {
	command := "set registry url"
	cmd.Spec = "URL"
	var url = cmd.StringArg("URL", "", "registry url")
	cmd.Action = func() {
		HandleError(client.Set(client.ClientConfigPath(), *url, "registry", "url"), command)
		fmt.Printf("Registry: %s\n", *url)
		Reporter.SendHoneycombEvent(command, map[string]interface{}{
			"url":    url,
			"result": "success",
		})
		Reporter.SendSnowflakeEvent("set", map[string]interface{}{
			"secondary_command_get_set": "registry URL",
			"additional_info":           "url:" + *url,
		})
	}
}

func CmdSetTarget(cmd *cli.Cmd) {
	command := "set target"
	cmd.Spec = "TARGET"
	var target = cmd.StringArg("TARGET", "", "your default target")
	cmd.Action = func() {
		HandleError(client.Set(client.ClientConfigPath(), *target, "target"), command)
		fmt.Printf("Active target: %s\n", *target)
		Reporter.SendHoneycombEvent(command, map[string]interface{}{
			"target": target,
			"result": "success",
		})
		Reporter.SendSnowflakeEvent("set", map[string]interface{}{
			"secondary_command_get_set": "target",
			"additional_info":           "target:" + *target,
		})
	}
}
