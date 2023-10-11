package insights

import (
	"os/exec"
)

type Insights struct {
	binaryPath string
}

func New(binaryPath string) *Insights {
	return &Insights{binaryPath: binaryPath}
}

func (i Insights) Exec(mustGatherPath string) ([]byte, error) {
	cmd := exec.Command(i.binaryPath, "run", "-p", "ccx_rules_ocp", "-f", "json", mustGatherPath)
	return cmd.Output()
}
