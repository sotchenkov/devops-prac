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
  default     = "dev"

  validation {
    condition     = var.environment == "dev"
    error_message = "Environment must be dev"
  }
}

variable "backend_image" {
  type        = string
  description = "Image for backend containers"
  default     = "traefik/whoami:v1.11.0"
}

variable "proxy_image" {
  type        = string
  description = "Image for reverse proxy"
  default     = "nginx:1.31"
}

variable "proxy_host_port" {
  type        = number
  description = "Host port for proxy container"
  default     = 8080
}

variable "backends" {
  type        = set(string)
  description = "Names of backend containers"

  default = [
    "backend-01",
    "backend-canary"
  ]

  validation {
    condition     = length(var.backends) > 0
    error_message = "backends can't be empty"
  }
}
