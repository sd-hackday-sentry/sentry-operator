SHELL := /bin/bash
GITCOMMIT=$(shell git rev-parse --short HEAD)$(shell [[ $$(git status --porcelain) = "" ]] || echo -dirty)
LDFLAGS="-X main.gitCommit=$(GITCOMMIT)"

ifndef QUAY_REPO
	QUAY_REPO:=thekad/sentry-operator
endif

OPERATOR_IMAGE ?= quay.io/$(QUAY_REPO):$(GITCOMMIT)

.PHONY: setup generate image push deploy scrub

setup:
	mkdir -pv vendor
	GO111MODULE=on go mod vendor

generate:
	operator-sdk generate k8s
	operator-sdk generate openapi

image: generate
	operator-sdk build $(OPERATOR_IMAGE)

push: image
	docker push $(OPERATOR_IMAGE)

deploy:
	cat deploy/operator.yaml.in | sed -e 's|REPLACE_IMAGE|$(OPERATOR_IMAGE)|g' > deploy/operator.yaml
	kubectl apply -f deploy/service_account.yaml
	kubectl apply -f deploy/role.yaml
	kubectl apply -f deploy/role_binding.yaml
	kubectl apply -f deploy/crds/sentry_v1alpha1_sentry_crd.yaml
	kubectl apply -f deploy/operator.yaml
	#kubectl apply -f deploy/crds/sentry_v1alpha1_sentry_cr.yaml

scrub:
	cat deploy/operator.yaml.in | sed -e 's|REPLACE_IMAGE|$(OPERATOR_IMAGE)|g' > deploy/operator.yaml
	#kubectl apply -f deploy/crds/sentry_v1alpha1_sentry_cr.yaml
	kubectl delete -f deploy/operator.yaml
	kubectl delete -f deploy/role.yaml
	kubectl delete -f deploy/role_binding.yaml
	kubectl delete -f deploy/service_account.yaml
	kubectl delete -f deploy/crds/sentry_v1alpha1_sentry_crd.yaml
