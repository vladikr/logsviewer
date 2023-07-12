LDFLAGS ?= -w -s

metrics-docs:
	mkdir -p docs
	go run -ldflags="${LDFLAGS}" ./tools/metrics-docs > docs/metrics.md

.PHONY: metrics-docs
