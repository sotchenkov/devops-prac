output "backend_urls" {
  description = "Backend URLs"
  value       = { for key, container in docker_container.backend : key => "http://localhost:${container.ports[0].external}" }
}

output "backend_container_ids" {
  description = "Backend container IDs"
  value       = { for key, container in docker_container.backend : key => container.id }
}

output "network_name" {
  description = "Network name"
  value       = docker_network.backend.name
}

output "docker_image_digest" {
  description = "Docker image digest"
  value       = data.docker_registry_image.backend_image_data.sha256_digest
}