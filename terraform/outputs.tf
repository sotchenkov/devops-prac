output "proxy_url" {
  description = "Reverse proxy URL"
  value       = module.service_stack.proxy_url
}

output "backend_container_ids" {
  description = "Backend container IDs"
  value       = module.service_stack.backend_container_ids
}

output "proxy_container_id" {
  description = "Proxy container ID"
  value       = module.service_stack.proxy_container_id
}

output "network_name" {
  description = "Network name"
  value       = module.service_stack.network_name
}

output "docker_image_digest" {
  description = "Docker image digest"
  value       = module.service_stack.docker_image_digest
}
