.PHONY: cluster-create cluster-status check-context check-environment cluster-delete bootstrap

CLUSTER := homelab-cluster
CONTEXT := kind-homelab-cluster
ENV ?= dev
ENVIRONMENTS := dev stage prod
ENVIRONMENT_DIR := environments/$(ENV)

cluster-create:
	kind create cluster --config ./cluster/kind.yaml

cluster-status:
	kubectl cluster-info --context $(CONTEXT)

check-context:
	@context="$$(kubectl config current-context 2>/dev/null)"; \
		status=$$?; \
		if [ $$status -ne 0 ] || [ -z "$$context" ]; then \
			echo "Error: unable to get current Kubernetes context"; \
			exit 1; \
		fi; \
		awk -v actual="$$context" -v expected="$(CONTEXT)" 'BEGIN { \
			if (actual != expected) { \
				print "Error: wrong context: " actual; \
				print "Expected: " expected; \
				exit 2; \
			} \
			print "OK: " actual; \
		}'

check-environment:
	@case " $(ENVIRONMENTS) " in \
		*" $(ENV) "*) ;; \
		*) \
			echo "Error: unsupported ENV: $(ENV)"; \
			echo "Expected one of: $(ENVIRONMENTS)"; \
			exit 2; \
		;; \
	esac

cluster-delete:
	kind delete cluster -n $(CLUSTER)

bootstrap: check-context check-environment
	kubectl apply --context $(CONTEXT) -f $(ENVIRONMENT_DIR)/namespace.yaml
	kubectl apply --context $(CONTEXT) -f $(ENVIRONMENT_DIR)/limits.yaml
