package workitems

import (
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients/workitems"
	"strings"
)

func GetAreaOrIterationPath(node *workitems.WorkItemClassificationNode) string {
	components := strings.Split(*node.Path, "\\")
	var pathComponents []string
	if len(components) > 3 {
		pathComponents = components[3:]
	} else {
		pathComponents = []string{""}
	}

	finalPath := strings.Join(pathComponents, "/")
	return finalPath
}

func PlanAreaOrIterationPath(path string, name string, isMove bool) string {
	if isMove {
		return strings.TrimPrefix(strings.Join([]string{path, name}, "/"), "/")
	} else {
		components := strings.Split(path, "/")
		components[len(components)-1] = name
		return strings.Join(components, "/")
	}
}
