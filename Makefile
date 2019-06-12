NAMESPACE=unifiedpush
CODE_COMPILE_OUTPUT = build/_output/bin/unifiedpush-operator
TEST_COMPILE_OUTPUT = build/_output/bin/unifiedpush-operator-test

.PHONY: setup/travis
setup/travis:
	@echo Installing dep
	curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
	@echo Installing Operator SDK
	curl -Lo ${GOPATH}/bin/operator-sdk https://github.com/operator-framework/operator-sdk/releases/download/v0.7.0/operator-sdk-v0.7.0-x86_64-linux-gnu
	chmod +x ${GOPATH}/bin/operator-sdk
	@echo setup complete

.PHONY: code/compile
code/compile:
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o=$(CODE_COMPILE_OUTPUT) ./cmd/manager/main.go

.PHONY: code/run
code/run: code/gen
	operator-sdk up local --namespace=$(NAMESPACE)

.PHONY: code/gen
code/gen: code/fix
	operator-sdk generate k8s
	go generate ./...

.PHONY: code/fix
code/fix:
	gofmt -w `find . -type f -name '*.go' -not -path "./vendor/*"`

.PHONY: test/unit
test/unit:
	@echo Running tests:
	CGO_ENABLED=1 go test -v -race -cover ./pkg/...

.PHONY: test/compile
test/compile:
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go test -c -o=$(TEST_COMPILE_OUTPUT) ./test/e2e/...

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
