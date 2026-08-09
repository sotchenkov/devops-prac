.PHONY: argo-install

ARGOCD_CHART_VERSION := 10.2.1
ARGOCD_NAMESPACE := argocd
KUBE_CONTEXT := kind-homelab-cluster

argo-install:
	helm repo add argo https://argoproj.github.io/argo-helm --force-update
	helm repo update argo
	helm upgrade --install argocd argo/argo-cd \
		--kube-context $(KUBE_CONTEXT) \
		--namespace $(ARGOCD_NAMESPACE) \
		--create-namespace \
		--version $(ARGOCD_CHART_VERSION) \
		--wait \
		--timeout 10m