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

resource "docker_image" "nginx" {
  name = "nginx:1.31"
}

resource "docker_container" "nginx" {
  image = docker_image.nginx.image_id
  name  = "homelab"
  ports {
    internal = 80
    external = 8080
  }
  restart = "unless-stopped"
}
