package cmd

import (
	"os"

	cli "github.com/jawher/mow.cli"

	"usi/pkg/client"
)

func CmdApplyResource(app *cli.Cmd) {
	command := "upload"
	filename := app.StringArg("FILENAME", "", "file to upload (apply)")
	Reporter.UsedOption("filename", filename)
	app.Action = func() {
		resource := client.Resource{}
		ReadResource(*filename, &resource)
		HandleError(Workspace(nil, os.Stdout, os.Stderr, command).SaveResource(Requester(), resource), command)
		PrintYAML(resource, command)
		Reporter.SendHoneycombEvent(command, map[string]interface{}{
			"result": "success",
		})
		Reporter.SendSnowflakeEvent(command, map[string]interface{}{})

	}
}
