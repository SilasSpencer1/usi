package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"usi/pkg/errors"
	cli "github.com/jawher/mow.cli"
)

func HandleError(err error) {
	if err != nil {
		if _, ok := err.(*errors.Error); !ok {
			err = errors.New(err)
		}
		err := err.(*errors.Error)
		if err.Code != errors.Unexpected {
			println(err.Message)
		} else {
			println(err.ErrorStack())
		}
		os.Exit(1)
	}
}

func CmdUpload(app *cli.Cmd) {
	path := app.StringArg("PATH", "", "artifact file path")
	url := app.StringArg("URL", "", "artifactory url")
	app.Action = func() {
		f, err := os.Open(*path)
		if err != nil {
			HandleError(err)
		}
		defer func(f *os.File) {
			_ = f.Close()
		}(f)

		payload := bufio.NewReader(f)
		req, err := http.NewRequest("PUT", *url, payload)
		if err != nil {
			HandleError(err)
		}
		token := os.Getenv("ARTIFACTORY_TOKEN")
		if token == "" {
			HandleError(errors.New("env var ARTIFACTORY_TOKEN is required"))
		}
		req.Header.Add("Content-Type", "application/octet-stream")
		req.Header.Add("Authorization", "Bearer "+token)
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			HandleError(err)
		}

		defer func(Body io.ReadCloser) {
			_ = Body.Close()
		}(res.Body)
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			HandleError(err)
		}

		fmt.Println(res)
		fmt.Println(string(body))
	}
}

func CmdDelete(app *cli.Cmd) {
	url := app.StringArg("URL", "", "artifactory url")
	app.Action = func() {
		req, err := http.NewRequest("DELETE", *url, nil)
		if err != nil {
			HandleError(err)
		}
		token := os.Getenv("ARTIFACTORY_TOKEN")
		if token == "" {
			HandleError(errors.New("env var ARTIFACTORY_TOKEN is required"))
		}
		req.Header.Add("Content-Type", "application/octet-stream")
		req.Header.Add("Authorization", "Bearer "+token)
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			HandleError(err)
		}

		defer func(Body io.ReadCloser) {
			_ = Body.Close()
		}(res.Body)
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			HandleError(err)
		}

		fmt.Println(res)
		fmt.Println(string(body))
	}
}

func main() {
	app := cli.App("Artifactory", "Artifactory Client")
	app.Command("upload", "upload an artifact to artifactory", CmdUpload)
	app.Command("delete", "delete artifact on artifactory", CmdDelete)
	_ = app.Run(os.Args)
}
