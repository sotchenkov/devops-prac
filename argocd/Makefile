.PHONY: argo-install argo-ready argo-uninstall

ARGOCD_CHART_VERSION := 10.2.1
ARGOCD_NAMESPACE := argocd
KUBE_CONTEXT := kind-homelab-cluster
ARGOCD_VALUES := values.yaml

argo-install:
	helm repo add argo https://argoproj.github.io/argo-helm --force-update
	helm repo update argo
	helm upgrade --install argocd argo/argo-cd \
		--kube-context $(KUBE_CONTEXT) \
		--namespace $(ARGOCD_NAMESPACE) \
		--values $(ARGOCD_VALUES) \
		--create-namespace \
		--version $(ARGOCD_CHART_VERSION) \
		--wait \
		--timeout 10m

argo-ready:
	kubectl --context $(KUBE_CONTEXT) wait \
		--namespace $(ARGOCD_NAMESPACE) \
		--for=condition=Available \
		deployment \
		--all \
		--timeout 5m
	kubectl --context $(KUBE_CONTEXT) rollout status \
		statefulset/argocd-application-controller \
		--namespace $(ARGOCD_NAMESPACE) \
		--timeout 5m

argo-uninstall:
	helm uninstall argocd \
        --kube-context $(KUBE_CONTEXT) \
        --namespace $(ARGOCD_NAMESPACE) \
        --wait \
        --timeout 2m
