terraform {
  required_version = "~> 1.15.0"
  required_providers {
    docker = {
      source  = "kreuzwerker/docker"
      version = "~> 4.5.0"
    }
  }
}

variable "project_name" {
  type        = string
  description = "Project name"
  default     = "homelab"

  validation {
    condition     = length(var.project_name) > 0
    error_message = "project_name can't be empty"
  }
}

variable "environment" {
  type        = string
  description = "Deployment environment name"
  default     = "local"

  validation {
    condition     = contains(["local", "test", "prod"], var.environment)
    error_message = "Environment must be local, test, or prod."
  }
}

variable "backend_image" {
  type        = string
  description = "Image for backend containers"
  default     = "traefik/whoami:v1.11.0"
}

variable "backends" {
  type = map(object({
    external_port = number
  }))

  description = "Names and external_ports of backend containers"
  default = {
    "backend-01" = {
      "external_port" : 8080
    }
    "backend-02" = {
      "external_port" : 8081
    }
    "backend-03" = {
      "external_port" : 8082
    }
  }
  validation {
    condition     = length(var.backends) > 0
    error_message = "backends can't be empty"
  }

  validation {
    condition = alltrue([
      for backend in values(var.backends) :
      backend.external_port >= 1024 && backend.external_port <= 65535
    ])
    error_message = "external_port must be between 1024 and 65535"
  }

  validation {
    condition = length([
      for backend in values(var.backends) : backend.external_port
      ]) == length(distinct([
        for backend in values(var.backends) : backend.external_port
    ]))

    error_message = "external_port must be uniq"
  }

}

provider "docker" {
  context = "colima"
}

resource "docker_image" "traefik" {
  name = var.backend_image
}

resource "docker_network" "backend" {
  name = "${var.environment}-${var.project_name}-backend"
}

resource "docker_container" "backend" {
  for_each = var.backends

  image = docker_image.traefik.image_id
  name  = each.key

  networks_advanced {
    name = docker_network.backend.name
  }

  ports {
    internal = 80
    external = each.value.external_port
  }

  restart = "unless-stopped"
}

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
