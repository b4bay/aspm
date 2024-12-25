package server

import "strings"

func GetProjectFromEnvironment(e map[string]string) string {
	project, ok := e["CI_PROJECT_PATH"]
	if ok {
		return project
	} else {
		return ""
	}
}

func GetAuthorFromEnvironment(e map[string]string) string {
	author, ok := e["GITLAB_USER_NAME"]
	if ok {
		return author
	} else {
		return ""
	}
}

func GetWorkerFromEnvironment(e map[string]string) string {
	worker, ok := e["CI_RUNNER_DESCRIPTION"]
	if ok {
		worker := strings.SplitN(worker, "/", 2)[0]
		return worker
	} else {
		return ""
	}
}
