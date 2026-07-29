terraform {
  required_version = "~> 1.15.0"
  required_providers {
    docker = {
      source  = "kreuzwerker/docker"
      version = "~> 4.5.0"
    }
  }
}

provider "docker" {
  context = "colima"
}

resource "docker_image" "traefik" {
  name = "traefik/whoami:v1.11.0"
}

resource "docker_network" "backend" {
  name = "homelab-backend"
}

resource "docker_container" "backend" {
  for_each = {
    "backend-01" = {
      "name" : "homelab-backend-01"
      "external_port" : 8080
    }
    "backend-02" = {
      "name" : "homelab-backend-02"
      "external_port" : 8081
    }
    "backend-03" = {
      "name" : "homelab-backend-03"
      "external_port" : 8082
    }
  }

  image = docker_image.traefik.image_id
  name  = each.value.name

  networks_advanced {
    name = docker_network.backend.name
  }

  ports {
    internal = 80
    external = each.value.external_port
  }

  restart = "unless-stopped"
}
