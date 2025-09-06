# USI CLI client

## Building and running the CLI client

Run `bazel build //cmd/cli:usi`. This will build the CLI binary. You should be able
to run it directly from `bazel-bin/cmd/cli/usi_/usi`.

Alternatively, you can run it through go with a command like `go run cmd/cli/main.go ...`

## Debugging the CLI client

TODO - add more details here

### Debugging cli errors
- The `HandleError` function in the `cmd` package will display the stacktrace to STDOUT if you set the `DEBUG` environment variable to anything but `false`:
```shell
#default error handling behavior at the cli
./cli envfile -n dealer-engagement-platform.na -e dealer.production-na-ue1
USI Workspace: /Users/testuser/projects/cg-main
Unauthorized
```
- set your `DEBUG` environment variable. Could be any value other than `false`
```shell
export DEBUG=true
```
- re-run your commands, this time you see the error along with the stacktrace. Handy for troubleshooting.
```shell
/cli envfile -n dealer-engagement-platform.na -e dealer.production-na-ue1
USI Workspace: /Users/testuser/projects/cg-main
*errors.errorString Unauthorized
/Users/sspencer/workspace/projects/glados/pkg/errors/errors.go:36 (0x1311885)
	Unmarshal: return WithCode(result.Message, result.Code)
/Users/sspencer/workspace/projects/glados/pkg/http/client.go:86 (0x1de68ce)
	GenerateClientMethodHandler.func1: err := errors.Unmarshal(b)
/Users/sspencer/workspace/projects/glados/pkg/usi/client/client.go:315 (0x1f20978)
	(*Workspace).ClearTextConfiguration: return &response, c.client.GET("/cleartextconfiguration", handlers.ClearTextConfigurationRequest{
/Users/sspencer/workspace/projects/glados/pkg/usi/workspace/workspace.go:526 (0x1f208bb)
	(*Workspace).ClearTextConfiguration: return w.client.ClearTextConfiguration(deployment)
/Users/sspencer/workspace/projects/glados/cmd/cli/cmd/envfile.go:50 (0x26e34fc)
	CmdEnvFile.func1: conf, err := Workspace(nil).ClearTextConfiguration(core.RequestFromUUID(d.UUID))
/Users/sspencer/go/pkg/mod/github.com/jawher/mow.cli@v1.2.0/internal/flow/flow.go:55 (0x1f23b37)
	(*Step).callDo: s.Do()
/Users/sspencer/go/pkg/mod/github.com/jawher/mow.cli@v1.2.0/internal/flow/flow.go:25 (0x1f23a11)
	(*Step).Run: s.callDo(p)
/Users/sspencer/go/pkg/mod/github.com/jawher/mow.cli@v1.2.0/internal/flow/flow.go:29 (0x1f23a6e)
	(*Step).Run: s.Success.Run(p)
/Users/sspencer/go/pkg/mod/github.com/jawher/mow.cli@v1.2.0/internal/flow/flow.go:29 (0x1f23a6e)
	(*Step).Run: s.Success.Run(p)
/Users/sspencer/go/pkg/mod/github.com/jawher/mow.cli@v1.2.0/internal/flow/flow.go:29 (0x1f23a6e)
	(*Step).Run: s.Success.Run(p)
/Users/sspencer/go/pkg/mod/github.com/jawher/mow.cli@v1.2.0/commands.go:693 (0x1f2d31a)
	(*Cmd).parse: entry.Run(nil)
/Users/sspencer/go/pkg/mod/github.com/jawher/mow.cli@v1.2.0/commands.go:707 (0x1f2d57d)
	(*Cmd).parse: return sub.parse(args[1:], entry, newInFlow, newOutFlow)
/Users/sspencer/go/pkg/mod/github.com/jawher/mow.cli@v1.2.0/cli.go:76 (0x1f2ad9f)
	(*Cli).parse: return cli.Cmd.parse(args, entry, inFlow, outFlow)
/Users/sspencer/go/pkg/mod/github.com/jawher/mow.cli@v1.2.0/cli.go:105 (0x1f2af97)
	(*Cli).Run: return cli.parse(args[1:], inFlow, inFlow, outFlow)
/Users/sspencer/workspace/projects/glados/cmd/cli/main.go:54 (0x26f1cb6)
	main: err = app.Run(os.Args)
/usr/local/go/src/runtime/proc.go:255 (0x10392c7)
	main: fn()
/usr/local/go/src/runtime/asm_amd64.s:1581 (0x106a121)
	goexit: BYTE	$0x90	// NOP
```