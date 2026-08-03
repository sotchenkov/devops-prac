.PHONY: terraform-check terraform-init

terraform-check:
	terraform fmt -check -recursive
	@set -e; \
	for dir in modules/service_stack environments/dev environments/prod; do \
		echo "Validating $$dir..."; \
		terraform -chdir="$$dir" validate; \
	done

terraform-init:
	@set -e; \
	for dir in modules/service_stack environments/dev environments/prod; do \
	    echo "Initialising $$dir..."; \
		terraform -chdir="$$dir" init -backend=false -input=false; \
	done
