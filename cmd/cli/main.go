package main

import (
	"os"

	cli "github.com/jawher/mow.cli"

	usi "usi/pkg/client"

	"usi/cmd/cli/cmd"
)

func main() {
	err := usi.LoadClientConfig()
	if err != nil {
		cmd.PrintError("Unable to load usi client config %s", err.Error())
	}
	cmd.InitUsageReporter()

	defer cmd.Reporter.Close()

	app := cli.App("usi", "Supercharged Infrastructure - For more details on using usi, see our use case guide at [link coming soon]")
	app.BoolOpt("h help", false, "show usage information and help")
	app.Command("annotate", "add annotations to a resource by UUID", cmd.CmdAnnotate)
	app.Command("bounce", "bounce a deployment", cmd.CmdBounce)
	app.Command("configure", "configure a resource", cmd.Configure)
	// app.Command("create", "create things", func(app *cli.Cmd) {
	// 	app.Command("environment", "create a new environment", cmd.CmdCreateEnvironment)
	// })
	app.Command("clean", "clean up a resource", func(app *cli.Cmd) {
		app.Command("environment", "clean up an environment (undeploy all)", cmd.CmdCleanEnvironment)
	})
	app.Command("deploy", "deploy a service", cmd.CmdDeploy)
	app.Command("download", "download a resource", cmd.CmdDownloadResource)
	app.Command("envfile", "Extract an envfile from a deployment", cmd.CmdEnvFile)
	app.Command("get", "get options", cmd.CmdGet)
	app.Command("help", "show usage information and help", func(cmd *cli.Cmd) { cmd.Action = app.PrintLongHelp })
	app.Command("init", "new workspace", cmd.CmdInit)
	app.Command("resolve", "resolve configuration for a service", cmd.CmdResolve)
	app.Command("run", "run a predefined script", cmd.CmdRun)
	app.Command("debug", "debug mode", func(app *cli.Cmd) {
		app.Command("logs", "print out local logs", cmd.CmdDebugLogs)
	})
	app.Command("set", "set options", cmd.CmdSet)
	app.Command("undeploy", "undeploy a service", cmd.CmdUndeploy)
	app.Command("unset", "unset preferences", cmd.CmdUnset)
	app.Command("unstable", "unstable commands (Do not use without consent). WARNING: These commands may be unstable or experimental.", func(app *cli.Cmd) {
		app.Hidden = true
		app.Command("upload", "upload (apply) resource changes", cmd.CmdApplyResource)
		app.Command("manifests", "Get Kubernetes manifest for a specified deployment", cmd.CmdK8sManifestsDeprecated)
	})
	app.Command("validate", "validate m5.yaml against an environment", cmd.CmdValidate)
	app.Command("wait", "wait for a dev deployed service to become available", cmd.CmdWait)
	app.Command("logs", "tail logs for a dev deployed service", cmd.CmdLogs)
	app.Command("remote", "remote usi command execution", func(app *cli.Cmd) {
		app.Command("deploy", "remote usi deploy execution", cmd.CmdRemoteDeploy)
		app.Command("delete", "delete remote usi job", cmd.CmdDeleteRemoteJob)
	})

	app.Command("goto", "[Experimental] launch a service link in a browser", cmd.CmdGoTo)
	app.Command("k8s", "[Experimental] inspect and troubleshoot CGService Kubernetes resources", cmd.CmdK8s)

	err = app.Run(os.Args)
	if err != nil {
		cmd.PrintError("Unable to execute command %s", err.Error())
	}

	cmd.PrintWarningIfRegistryNotProd()
}
