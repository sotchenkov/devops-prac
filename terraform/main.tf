data "docker_registry_image" "backend_image_data" {
  name = var.backend_image
}

resource "docker_image" "backend" {
  name          = data.docker_registry_image.backend_image_data.name
  pull_triggers = [data.docker_registry_image.backend_image_data.sha256_digest]
}

resource "docker_network" "backend" {
  name = "${var.environment}-${var.project_name}-backend"
}

resource "docker_container" "backend" {
  for_each = var.backends

  image = docker_image.backend.image_id
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
