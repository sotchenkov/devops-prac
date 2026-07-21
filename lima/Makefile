VM_NAME := ansible-node-1
VM_TEMPLATE := ubuntu-24.04

.PHONY: homelab-up homelab-down homelab-clear

homelab-up:
	@if limactl list -q | grep -qx "$(VM_NAME)"; then \
		echo "Starting existing VM $(VM_NAME)..."; \
		limactl start "$(VM_NAME)"; \
	else \
		echo "Creating VM $(VM_NAME)..."; \
		limactl start --yes \
			--name="$(VM_NAME)" \
			--arch=aarch64 \
			--vm-type=vz \
			--cpus=2 \
			--memory=2 \
			--disk=20 \
			--containerd=none \
			--mount-none \
			--ssh-port=60022 \
			template:$(VM_TEMPLATE); \
	fi
	@limactl list "$(VM_NAME)"

homelab-down:
	limactl stop "$(VM_NAME)"

homelab-clear:
	limactl delete --yes --force "$(VM_NAME)"
