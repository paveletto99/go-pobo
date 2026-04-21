SHELL := /usr/bin/env bash

.PHONY: test-template
test-template:
	cd templates/service/sample-service && go test ./...

.PHONY: gofmt-template
gofmt-template:
	cd templates/service/sample-service && gofmt -w $$(find . -name '*.go')

.PHONY: skaffold-render
skaffold-render:
	skaffold render --filename skaffold.yaml

.PHONY: yaml-check
yaml-check:
	ruby -e 'require "yaml"; Dir["**/*.yaml"].each { |f| YAML.load_stream(File.read(f)); puts f }'

.PHONY: scaffold
scaffold:
	@test -n "$(NAME)" || (echo "usage: make scaffold NAME=my-service OUT=../my-service" >&2; exit 1)
	./scripts/scaffold-service.sh "$(NAME)" "$(or $(OUT),../$(NAME))"
