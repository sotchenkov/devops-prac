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

variable "proxy_image" {
  type        = string
  description = "Image for reverse proxy"
  default     = "nginx:1.31"
}

variable "backends" {
  type        = set(string)
  description = "Names of backend containers"

  default = [
    "backend-01",
    "backend-02",
    "backend-03",
  ]

  validation {
    condition     = length(var.backends) > 0
    error_message = "backends can't be empty"
  }
}
