package cmd

import (
	"os"

	provider "usi/pkg/provider/client/annotations"

	"usi/pkg/workspace"

	"usi/pkg/core"

	cli "github.com/jawher/mow.cli"
)

func CmdAnnotate(app *cli.Cmd) {
	command := "annotate"
	uuid := app.StringArg("UUID", "", "resource uuid")
	Reporter.UsedOption("uuid", uuid)
	annotations := app.StringArg("ANNOTATIONS", "", "annotations in the format (k=v,k=v)")
	Reporter.UsedOption("annotations", annotations)
	app.Action = func() {
		ws := Workspace(nil, os.Stdout, os.Stderr, command)
		AnnotateAction(ws, uuid, annotations)
	}
}

func AnnotateAction(ws workspace.Workspace, uuid, annotations *string) {
	command := "annotate"
	r, err := ws.Annotate(
		Requester(),
		core.Request{UUID: uuid},
		parseAnnotationString(*annotations),
	)
	HandleError(err, command)

	PrintResource(*r, command)
	Reporter.SendHoneycombEvent(command, map[string]interface{}{})
}

func parseAnnotationString(annotations string) core.Annotations {
	command := "annotate"
	result, err := provider.ParseAnnotationString(annotations)
	HandleError(err, command)
	return result
}
package cmd

import (
"os"

provider "usi/pkg/provider/client/annotations"

"usi/pkg/workspace"

"usi/pkg/core"

cli "github.com/jawher/mow.cli"
)

func CmdAnnotate(app *cli.Cmd) {
	command := "annotate"
	uuid := app.StringArg("UUID", "", "resource uuid")
	Reporter.UsedOption("uuid", uuid)
	annotations := app.StringArg("ANNOTATIONS", "", "annotations in the format (k=v,k=v)")
	Reporter.UsedOption("annotations", annotations)
	app.Action = func() {
		ws := Workspace(nil, os.Stdout, os.Stderr, command)
		AnnotateAction(ws, uuid, annotations)
	}
}

func AnnotateAction(ws workspace.Workspace, uuid, annotations *string) {
	command := "annotate"
	r, err := ws.Annotate(
		Requester(),
		core.Request{UUID: uuid},
		parseAnnotationString(*annotations),
	)
	HandleError(err, command)

	PrintResource(*r, command)
	Reporter.SendHoneycombEvent(command, map[string]interface{}{})
}

func parseAnnotationString(annotations string) core.Annotations {
	command := "annotate"
	result, err := provider.ParseAnnotationString(annotations)
	HandleError(err, command)
	return result
}
