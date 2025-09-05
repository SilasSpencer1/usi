# Universal Supercharged Infastructure

***See [cmd/server/README.md](cmd/server/README.md) for USI Registry docs***

***See [cmd/cli/README.md](cmd/cli/README.md) for USI CLI client docs***

## Concourse Pipeline (for building and deploying to stage and production)

TO BE UPDATED: link here

## Repo tools and scripts

Prior to merging code into the master branch, you should make sure you run
`./scripts/pre-commit.sh -u`. This script will run our generation, formatting, and linting
tools followed by building all targets and running unit and integration tests.

Running the pre-commit.sh script will require that you have some tools installed om
your local machine. You can use `./scripts/install-repo-tools.sh` in order to install
them.



### Linting and formatting Go code

First, make sure you are using the correct Golang version since the formatter differs
across version. You can see the current version at the top of the `go.mod` file.

```
go fmt code.cargurus.com/platform/glados/...
```

### Generating Go code

This repo includes some mocks and protobuf messages that may need to be regenerated if
there is a change to them. In order to run generation, use:

```
go generate ./...
```

## Testing

See [TESTING.md](TESTING.md)