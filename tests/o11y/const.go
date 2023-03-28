package o11y

import "time"

const (
	O11yUser = "o11y-e2e"
	O11ySA   = "pipeline"

	o11yUserSecret string = "o11y-tests-token"

	pipelinerun_timeout  = time.Second * 300
	pipelinerun_interval = time.Second * 10
)
