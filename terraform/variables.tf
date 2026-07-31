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