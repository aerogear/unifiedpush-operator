APP_NAME            ?= unifiedpush-operator
ORG_NAME            ?= aerogear
PKG                 ?= github.com/$(ORG_NAME)/$(APP_NAME)
TOP_SRC_DIRS        ?= pkg
PACKAGES            ?= $(shell sh -c "find $(TOP_SRC_DIRS) -name \\*_test.go \
                         -exec dirname {} \\; | sort | uniq")
TEST_PKGS           ?= $(addprefix $(PKG)/,$(PACKAGES))
APP_FILE            ?= ./cmd/manager/main.go
NAMESPACE           ?= unifiedpush
APP_NAMESPACE       ?= unifiedpush-apps
CODE_COMPILE_OUTPUT ?= build/_output/bin/unifiedpush-operator
TEST_COMPILE_OUTPUT ?= build/_output/bin/unifiedpush-operator-test
DEV_TAG             ?= $(shell sh -c "git rev-parse --short HEAD")

##############################
# Local Development          #
##############################

.PHONY: code/run
code/run: export SERVICE_NAMESPACE = $(NAMESPACE)
code/run: export APP_NAMESPACES    = $(NAMESPACE),$(APP_NAMESPACE)
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

##############################
# Application                #
##############################

.PHONY: install
install:
	@echo ....... Install .......
	- make cluster/prepare
	@echo ....... Applying the Operator .......
	- kubectl apply -n $(NAMESPACE) -f deploy/operator.yaml
	@echo ....... Applying UnifiedPush Server .......
	- kubectl apply -n $(NAMESPACE) -f deploy/crds/push_v1alpha1_unifiedpushserver_cr.yaml

.PHONY: monitoring/install
monitoring/install:
	@echo Installing service monitor in ${NAMESPACE} :
	- oc project ${NAMESPACE}
	- kubectl label namespace ${NAMESPACE} monitoring-key=middleware
	- kubectl apply -n $(NAMESPACE) -f deploy/monitor/service_monitor.yaml
	- kubectl apply -n $(NAMESPACE) -f deploy/monitor/operator_service.yaml
	- kubectl apply -n $(NAMESPACE) -f deploy/monitor/prometheus_rule.yaml
	- kubectl apply -n $(NAMESPACE) -f deploy/monitor/grafana_dashboard.yaml
	- kubectl apply -n $(NAMESPACE) -f deploy/monitor/push_service_monitor.yaml
	- kubectl apply -n $(NAMESPACE) -f deploy/monitor/push_prometheus_rule.yaml
	- kubectl apply -n $(NAMESPACE) -f deploy/monitor/push_grafana_dashboard.yaml

.PHONY: monitoring/uninstall
monitoring/uninstall:
	@echo Uninstalling monitor service from ${NAMESPACE} :
	- oc project ${NAMESPACE}
	- kubectl delete -n $(NAMESPACE) -f deploy/monitor/service_monitor.yaml
	- kubectl delete -n $(NAMESPACE) -f deploy/monitor/prometheus_rule.yaml
	- kubectl delete -n $(NAMESPACE) -f deploy/monitor/grafana_dashboard.yaml
	- kubectl delete -n $(NAMESPACE) -f deploy/monitor/push_service_monitor.yaml
	- kubectl delete -n $(NAMESPACE) -f deploy/monitor/push_prometheus_rule.yaml
	- kubectl delete -n $(NAMESPACE) -f deploy/monitor/push_grafana_dashboard.yaml
	- kubectl delete -n $(NAMESPACE) -f deploy/monitor/operator_service.yaml

.PHONY: example-pushapplication/apply
example-pushapplication/apply:
	@echo ....... Applying the PushApplication example in the current namespace  ......
	- kubectl apply -n $(APP_NAMESPACE) -f deploy/crds/examples/push_v1alpha1_pushapplication_cr.yaml

.PHONY: example-pushapplication/delete
example-pushapplication/delete:
	@echo ....... Deleting the PushApplication example in the current namespace  ......
	- kubectl delete -n $(APP_NAMESPACE) -f deploy/crds/examples/push_v1alpha1_pushapplication_cr.yaml

.PHONY: cluster/prepare
cluster/prepare:
	- kubectl create namespace $(NAMESPACE)
	- kubectl label namespace $(NAMESPACE) monitoring-key=middleware
	- kubectl create namespace $(APP_NAMESPACE)
	- kubectl create -n $(NAMESPACE) -f deploy/service_account.yaml
	- kubectl create -f deploy/role.yaml
	- kubectl create -n $(NAMESPACE) -f deploy/role_binding.yaml
	- kubectl apply -f deploy/crds/push_v1alpha1_pushapplication_crd.yaml
	- kubectl apply -f deploy/crds/push_v1alpha1_androidvariant_crd.yaml
	- kubectl apply -f deploy/crds/push_v1alpha1_webpushvariant_crd.yaml
	- kubectl apply -f deploy/crds/push_v1alpha1_iosvariant_crd.yaml
	- kubectl apply -f deploy/crds/push_v1alpha1_iostokenvariant_crd.yaml
	- kubectl apply -f deploy/crds/push_v1alpha1_unifiedpushserver_crd.yaml

.PHONY: cluster/clean
cluster/clean:
	- kubectl delete -n $(NAMESPACE) iosVariant --all
	- kubectl delete -n $(NAMESPACE) iosTokenVariant --all	
	- kubectl delete -n $(NAMESPACE) androidVariant --all
	- kubectl delete -n $(NAMESPACE) webpushVariant --all
	- kubectl delete -n $(NAMESPACE) pushApplication --all
	- kubectl delete -n $(NAMESPACE) unifiedpushServer --all
	- kubectl delete -f deploy/role.yaml
	- kubectl delete -n $(NAMESPACE) -f deploy/role_binding.yaml
	- kubectl delete -n $(NAMESPACE) -f deploy/service_account.yaml
	- kubectl delete -f deploy/crds/push_v1alpha1_pushapplication_crd.yaml
	- kubectl delete -f deploy/crds/push_v1alpha1_androidvariant_crd.yaml
	- kubectl delete -f deploy/crds/push_v1alpha1_webpushvariant_crd.yaml
	- kubectl delete -f deploy/crds/push_v1alpha1_iostokenvariant_crd.yaml
	- kubectl delete -f deploy/crds/push_v1alpha1_iosvariant_crd.yaml
	- kubectl delete -f deploy/crds/push_v1alpha1_unifiedpushserver_crd.yaml
	- make monitoring/uninstall
	- kubectl delete namespace $(NAMESPACE)
	- kubectl delete namespace $(APP_NAMESPACE)

.PHONY: image/build
image/build:
	operator-sdk build quay.io/${ORG_NAME}/${APP_NAME}:${DEV_TAG}

.PHONY: image/push
image/push: image/build
	docker push quay.io/${ORG_NAME}/${APP_NAME}:${DEV_TAG}
