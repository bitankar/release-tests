package cli

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"regexp"

	"github.com/getgauge-contrib/gauge-go/gauge"
	m "github.com/getgauge-contrib/gauge-go/models"
	"github.com/getgauge-contrib/gauge-go/testsuit"
	"github.com/openshift-pipelines/release-tests/pkg/pipelines"
	"github.com/openshift-pipelines/release-tests/pkg/store"
	"github.com/openshift-pipelines/release-tests/pkg/tkn"
)

var _ = gauge.Step("Start and verify pipeline <pipelineName> with param <paramName> with values stored in variable <variableName> with workspace <workspaceValue>", func(pipelineName, paramName, variableName, workspaceValue string) {
	values := store.GetScenarioDataSlice(variableName)
	params := make(map[string]string)
	workspaces := make(map[string]string)
	workspaces[strings.Split(workspaceValue, ",")[0]] = strings.Split(workspaceValue, ",")[1]
	for _, value := range values {
		if value == "latest" {
			continue
		}
		log.Printf("Starting pipeline %s with param %s=%s", pipelineName, paramName, value)
		params[paramName] = value
		pipelineRunName := tkn.StartPipeline(pipelineName, params, workspaces, "--use-param-defaults")
		pipelines.ValidatePipelineRun(store.Clients(), pipelineRunName, "successful", "no", store.Namespace())
	}
})

var _ = gauge.Step("Start and verify dotnet pipeline <pipelineName> with values stored in variable <variableName> with workspace <workspaceValue>", func(pipelineName, variableName, workspaceValue string) {
	values := store.GetScenarioDataSlice(variableName)
	params := make(map[string]string)
	workspaces := make(map[string]string)
	workspaces[strings.Split(workspaceValue, ",")[0]] = strings.Split(workspaceValue, ",")[1]
	paramName := "VERSION"
	re := regexp.MustCompile(`\d+\.\d+`)
	var exampleRevision string
	for _, value := range values {
		if value == "latest" {
			continue
		}
		exampleRevision = re.FindString(value)
		params[paramName] = value
		versionInt, _ := strconv.ParseFloat(exampleRevision, 64)
		if versionInt >= 5.0 {
			params["EXAMPLE_REVISION"] = "dotnet-" + exampleRevision
		} else {
			params["EXAMPLE_REVISION"] = "dotnetcore-" + exampleRevision
		}
		log.Printf("Starting pipeline %s with param %s=%s and EXAMPLE_REVISION=%s", pipelineName, paramName, value, params["EXAMPLE_REVISION"])
		pipelineRunName := tkn.StartPipeline(pipelineName, params, workspaces, "--use-param-defaults")
		pipelines.ValidatePipelineRun(store.Clients(), pipelineRunName, "successful", "no", store.Namespace())
	}
})

var _ = gauge.Step("Verify component versions <table>", func(tbl *m.Table) {
	tkn_version_map := tkn.UpdateVersionsMap()
	for _, row := range tbl.Rows {
		tool := strings.ToLower(row.Cells[0])
		expectedVersion := strings.ReplaceAll(strings.ToLower(row.Cells[1]), `v`, ``)
		if tkn_version_map[tool] != expectedVersion {
			testsuit.T.Fail(fmt.Errorf("version mismatch for tool :  %v - got: %v, want: %v", tool, tkn_version_map[tool], expectedVersion))
		}
	}
})
