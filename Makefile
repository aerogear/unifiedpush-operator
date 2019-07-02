APP_NAME = unifiedpush-operator
ORG_NAME = aerogear
PKG = github.com/$(ORG_NAME)/$(APP_NAME)
TOP_SRC_DIRS = pkg
PACKAGES ?= $(shell sh -c "find $(TOP_SRC_DIRS) -name \\*_test.go \
              -exec dirname {} \\; | sort | uniq")
TEST_PKGS = $(addprefix $(PKG)/,$(PACKAGES))
APP_FILE=./cmd/manager/main.go

NAMESPACE=unifiedpush
CODE_COMPILE_OUTPUT = build/_output/bin/unifiedpush-operator
TEST_COMPILE_OUTPUT = build/_output/bin/unifiedpush-operator-test

.PHONY: code/run
code/run: code/gen
	operator-sdk up local

.PHONY: code/gen
code/gen: code/fix
	operator-sdk generate k8s
	operator-sdk generate openapi
	go generate ./...

.PHONY: code/fix
code/fix:
	gofmt -w `find . -type f -name '*.go' -not -path "./vendor/*"`

.PHONY: cluster/prepare
cluster/prepare:
	-kubectl create namespace $(NAMESPACE)
	-kubectl label namespace $(NAMESPACE) monitoring-key=middleware
	-kubectl create -n $(NAMESPACE) -f deploy/service_account.yaml
	-kubectl create -n $(NAMESPACE) -f deploy/role.yaml
	-kubectl create -n $(NAMESPACE) -f deploy/role_binding.yaml
	-kubectl apply -n $(NAMESPACE) -f deploy/crds/push_v1alpha1_unifiedpushserver_crd.yaml
	-kubectl apply -n $(NAMESPACE) -f deploy/crds/push_v1alpha1_pushapplication_crd.yaml
	-kubectl apply -n $(NAMESPACE) -f deploy/crds/push_v1alpha1_androidvariant_crd.yaml
	-kubectl apply -n $(NAMESPACE) -f deploy/crds/push_v1alpha1_iosvariant_crd.yaml

.PHONY: cluster/clean
cluster/clean:
	-kubectl delete -n $(NAMESPACE) iosVariant --all
	-kubectl delete -n $(NAMESPACE) androidVariant --all
	-kubectl delete -n $(NAMESPACE) pushApplication --all
	-kubectl delete -n $(NAMESPACE) unifiedpushServer --all
	-kubectl delete -f deploy/role.yaml
	-kubectl delete -n $(NAMESPACE) -f deploy/role_binding.yaml
	-kubectl delete -n $(NAMESPACE) -f deploy/service_account.yaml
	-kubectl delete -n $(NAMESPACE) -f deploy/crds/push_v1alpha1_unifiedpushserver_crd.yaml
	-kubectl delete -n $(NAMESPACE) -f deploy/crds/push_v1alpha1_pushapplication_crd.yaml
	-kubectl delete -n $(NAMESPACE) -f deploy/crds/push_v1alpha1_androidvariant_crd.yaml
	-kubectl delete -n $(NAMESPACE) -f deploy/crds/push_v1alpha1_iosvariant_crd.yaml
	-kubectl delete namespace $(NAMESPACE)

##############################
# Jenkins                    #
##############################

.PHONY: test/compile
test/compile:
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go test -c -o=$(TEST_COMPILE_OUTPUT) ./test/e2e/...


.PHONY: code/compile
code/compile: code/gen
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o=$(CODE_COMPILE_OUTPUT) ./cmd/manager/main.go

##############################
# Tests / CI                 #
##############################

.PHONY: test/integration-cover
test/integration-cover:
	echo "mode: count" > coverage-all.out
	GOCACHE=off $(foreach pkg,$(PACKAGES),\
		go test -failfast -tags=integration -coverprofile=coverage.out -covermode=count $(addprefix $(PKG)/,$(pkg)) || exit 1;\
		tail -n +2 coverage.out >> coverage-all.out;)

.PHONY: test/unit
test/unit:
	@echo Running tests:
	CGO_ENABLED=1 go test -v -race -cover $(TEST_PKGS)

.PHONY: code/build/linux
code/build/linux:
	env GOOS=linux GOARCH=amd64 go build $(APP_FILE)
