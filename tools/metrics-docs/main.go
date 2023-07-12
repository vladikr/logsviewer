package main

import (
	"fmt"

	"github.com/machadovilaca/operator-observability/pkg/docs"

	"logsviewer/pkg/backend/monitoring/metrics"
)

const tpl = `# LogsViewer Metrics

{{- range . }}

### {{.Name}}
{{.Help}}.

Type: {{.Type}}.
{{- end }}

`

func main() {
	metrics.SetupMetrics()

	docsString := docs.BuildMetricsDocsWithCustomTemplate(metrics.ListMetrics(), nil, tpl)
	fmt.Println(docsString)
}
