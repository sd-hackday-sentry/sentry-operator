SHELL := /bin/bash
GITCOMMIT=$(shell git rev-parse --short HEAD)$(shell [[ $$(git status --porcelain) = "" ]] || echo -dirty)
LDFLAGS="-X main.gitCommit=$(GITCOMMIT)"
NAMESPACE="$(USER)-dev"

OPERATOR_REGISTRY ?= quay.io
OPERATOR_REPO ?= thekad/sentry-operator
OPERATOR_IMAGE ?= $(OPERATOR_REGISTRY)/$(OPERATOR_REPO):$(GITCOMMIT)
OPERATOR_LATEST ?= $(OPERATOR_REGISTRY)/$(OPERATOR_REPO):latest

.PHONY: setup install uninstall generate image push deploy scrub port-forward

# local dev setup
setup:
	@mkdir -pv vendor
	@GO111MODULE=on go mod vendor
	@kubectl get namespace $(NAMESPACE) || ( kubectl create namespace $(NAMESPACE) && sleep 2 )
	@kubectl config set-context $(shell kubectl config current-context) --namespace=$(NAMESPACE)
	@$(SHELL) hack/create-secret.sh $(NAMESPACE)

# install services e.g. redis and postgres
install:
	@kubectl --namespace=$(NAMESPACE) apply --filename=hack/redis-service.yaml
	@kubectl --namespace=$(NAMESPACE) apply --filename=hack/postgres-service.yaml

# uninstall services
uninstall:
	@kubectl --namespace=$(NAMESPACE) delete --filename=hack/redis-service.yaml
	@kubectl --namespace=$(NAMESPACE) delete --filename=hack/postgres-service.yaml

# generate api bindings
generate:
	@GO111MODULE=on operator-sdk generate k8s
	@GO111MODULE=on operator-sdk generate openapi

# build docker image
image: generate
	@GO111MODULE=on operator-sdk build $(OPERATOR_IMAGE)

# push docker image to repo (or load to kind cluster)
push:
	docker push $(OPERATOR_IMAGE)

# deploy the operator to the k8s cluster
deploy:
	@cat deploy/operator.yaml.in | sed -e 's|REPLACE_IMAGE|$(OPERATOR_IMAGE)|g' > deploy/operator.yaml
	@kubectl --namespace=$(NAMESPACE) apply --filename=deploy/service_account.yaml
	@kubectl --namespace=$(NAMESPACE) apply --filename=deploy/role.yaml
	@kubectl --namespace=$(NAMESPACE) apply --filename=deploy/role_binding.yaml
	@kubectl --namespace=$(NAMESPACE) apply --filename=deploy/operator.yaml
	@kubectl --namespace=$(NAMESPACE) apply --filename=deploy/crds/sentry_v1alpha1_sentry_crd.yaml
	@kubectl --namespace=$(NAMESPACE) apply --filename=deploy/crds/sentry_v1alpha1_sentry_cr.yaml

# forward the port to be accessed locally
port-forward:
	@kubectl --namespace=$(NAMESPACE) port-forward svc/sentry-web-ui 9000

# remove all traces of the operator from the k8s cluster
scrub:
	@cat deploy/operator.yaml.in | sed -e 's|REPLACE_IMAGE|$(OPERATOR_IMAGE)|g' > deploy/operator.yaml
	@kubectl --namespace=$(NAMESPACE) delete --filename=deploy/crds/sentry_v1alpha1_sentry_cr.yaml
	@kubectl --namespace=$(NAMESPACE) delete --filename=deploy/crds/sentry_v1alpha1_sentry_crd.yaml
	@kubectl --namespace=$(NAMESPACE) delete --filename=deploy/operator.yaml
	@kubectl --namespace=$(NAMESPACE) delete --filename=deploy/role_binding.yaml
	@kubectl --namespace=$(NAMESPACE) delete --filename=deploy/role.yaml
	@kubectl --namespace=$(NAMESPACE) delete --filename=deploy/service_account.yaml
