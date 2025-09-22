package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/gen2brain/beeep"
	cli "github.com/jawher/mow.cli"
	"platform-go-common/pkg/errors"

	"usi/pkg/kubernetes"

	"usi/pkg/core"
	"usi/pkg/registry"
	"usi/pkg/type/deployment"
)

func CmdWait(cmd *cli.Cmd) {
	command := "wait"
	cmd.Spec = "(-n=<serviceName> [-s=<selector>] [-e=<environment>])"
	opts := NewOpts(cmd)

	environmentName := opts.EnvironmentOpt()
	name := opts.NameOpt()
	selectorStr := opts.SelectorOpt()

	var waitDur time.Duration
	cmd.Action = func() {
		opts.Normalize(command)
		opts.Validate(command)

		environmentName = ToggleEnvironment(environmentName, name)
		selector := StrToSelector(selectorStr, command)
		normName := core.JoinNameAndSelector(*name, selector)
		AssertDeployment(command, *environmentName, normName)
		deployment := GetServiceDeployment(command, *environmentName, *name, selector)
		if deployment != nil {
			waitDur = WaitForDeployment(deployment, *environmentName, command)
		} else {
			HandleError(
				errors.WithCode(errors.InvalidServiceDeploymentErrorMessage, errors.NotFound),
				command,
			)
		}

		Reporter.SendHoneycombEvent(command, map[string]interface{}{
			"environment":     environmentName,
			"service_name":    name,
			"selectors":       selectorStr,
			"wait_duration_s": waitDur.Seconds(),
			"result":          "success",
		})
		Reporter.SendSnowflakeEvent(command, map[string]interface{}{
			"service_name":    *name,
			"additional_info": " selectors:" + *selectorStr + fmt.Sprintf(" wait_duration_s:%f", waitDur.Seconds()),
			"environment":     *environmentName,
			"wait_duration_s": waitDur.Seconds(),
		})
	}
}

func WaitForDeployment(deployment *deployment.Resource, environment, command string) time.Duration {
	// Validate that wait can be run
	validateRancherCli(command)
	// Wait can only be run against deployments in the user cluster
	devDeploy := IsDevEnvironmentDeployment(deployment, command)
	if !devDeploy {
		HandleError(
			errors.WithCode(
				fmt.Sprintf("Unable to wait for deployment within environment %s", environment),
				errors.BadRequest,
			),
			command,
		)
	}

	PrintHeader(fmt.Sprintf("Waiting for %s deployment ...", deployment.Name))
	PrintWarning("The service must have a startup probe in order to wait for application startup. Otherwise waiting will just return when the container starts up!\n")

	var deploymentK8sName, deploymentK8sKind string
	var found bool
	if deploymentK8sName, found = deployment.Annotations[kubernetes.AnnotationK8sNameKey]; !found {
		HandleError(
			errors.WithCode(
				"Unable to determine deployment name from annotations",
				errors.BadRequest,
			),
			command,
		)
	}
	if deploymentK8sKind, found = deployment.Annotations[kubernetes.AnnotationK8sKindKey]; !found {
		HandleError(
			errors.WithCode(
				"Unable to determine deployment kind from annotations",
				errors.BadRequest,
			),
			command,
		)
	}

	namespace := fmt.Sprintf("user-%s", os.Getenv("USER"))
	envK8s := GetEnvironmentKubernetes(command, environment)
	if envK8s != nil && envK8s.Namespace != nil {
		namespace = envK8s.Namespace.Namespace.Name
	}

	_, _ = ColoredOutput.HiBlue("Setting correct rancher kubectl context ...")
	setRancherContext(environment, command, envK8s)

	label := deployment.ShortName()
	if deployment.HasLegacyShortname() {
		label = deployment.ShortNameLegacy()
	}
	cmdArgs := []string{
		"rancher",
		"kubectl",
		"rollout",
		"status",
		deploymentK8sKind,
		"-n",
		namespace,
		"-l",
		fmt.Sprintf("app=%s", label),
	}
	cmdArgsString := strings.Join(cmdArgs, " ")
	_, _ = ColoredOutput.HiBlue("Running wait command ...")
	_, _ = fmt.Fprintf(os.Stdout, "> ")
	_, _ = ColoredOutput.Green("%s\n", cmdArgsString)
	execStart := time.Now()
	cmd := exec.Command("bash", "-c", cmdArgsString)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		_ = beeep.Notify("usi wait", fmt.Sprintf("Failed to wait for %s deployment.", deployment.Name), "")
		HandleError(
			errors.WithCode(
				fmt.Sprintf(
					"Failed to wait for %s deployment. Check the application's logs for errors.",
					deploymentK8sName,
				),
				errors.BadRequest,
			),
			command,
		)
	}
	dur := time.Since(execStart)
	dur = dur.Round(time.Second)
	completionMsg := fmt.Sprintf("%s has successfully started up after %s", deployment.Name, dur)
	fmt.Println(completionMsg)
	_ = beeep.Notify("usi wait", completionMsg, "")

	return dur
}

func validateRancherCli(command string) {
	cmd := exec.Command("rancher")
	err := cmd.Run()
	if err != nil {
		HandleError(
			errors.WithCode(
				"Rancher CLI must be setup on your machine. See http://cg/rancher-cli",
				errors.BadRequest,
			),
			command,
		)
	}
}

func setRancherContext(environment, command string, envK8s *registry.EnvironmentKubernetesResponse) {
	rancherProject := ""
	if envK8s != nil && envK8s.Cluster != nil && envK8s.Cluster.ProjectId != nil && envK8s.Cluster.ClusterId != nil {
		rancherProject = fmt.Sprintf("%s:%s", *envK8s.Cluster.ClusterId, *envK8s.Cluster.ProjectId)
	}

	if rancherProject == "" {
		HandleError(
			errors.WithCode(
				fmt.Sprintf("Unable to set rancher context for environment %s", environment),
				errors.NotImplemented,
			),
			command,
		)
	}

	cmdArgs := []string{"context", "switch", rancherProject}
	_, _ = fmt.Fprintf(os.Stdout, "> ")
	_, _ = ColoredOutput.Green("rancher %s\n", strings.Join(cmdArgs, " "))
	cmd := exec.Command("rancher", cmdArgs...)
	err := cmd.Run()
	if err != nil {
		HandleError(
			errors.WithCode(
				fmt.Sprintf(
					"Unable to set rancher context for environment %s to %s",
					environment,
					rancherProject,
				),
				errors.NotImplemented,
			),
			command,
		)
	}
}

//test
