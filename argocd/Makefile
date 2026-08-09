.PHONY: argo-install
	
argo-install:
	helm repo add argo https://argoproj.github.io/argo-helm; \
	helm repo update argo; \
	helm upgrade --install argocd argo/argo-cd \
        --namespace argocd \
        --create-namespace \
        --version 10.2.1 \
        --wait \
        --timeout 10m
