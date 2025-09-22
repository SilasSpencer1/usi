package cmd

import (
	"regexp"
	"time"

	"usi/pkg/type/deployment"
)

func IsDevEnvironmentDeployment(deployment *deployment.Resource, command string) bool {
	// necessary as top level environments don't have namespaces associated with them
	if deployment.Cluster == nil || len(deployment.Cluster.Name) == 0 {
		return false
	}
	clusterName := deployment.Cluster.Name

	matched, err := regexp.MatchString("user", clusterName)
	HandleError(err, command, "InternalError-DevEnvironmentValidation")

	return matched
}

func ConvertDateToLocalTZ(date time.Time) time.Time {
	userTimezone := time.Local

	userLocalTime := date.In(userTimezone)

	return userLocalTime
}
