data "docker_registry_image" "backend" {
  name = var.backend_image
}

data "docker_registry_image" "proxy" {
  name = var.proxy_image
}

resource "docker_image" "backend" {
  name          = data.docker_registry_image.backend.name
  pull_triggers = [data.docker_registry_image.backend.sha256_digest]
}

resource "docker_image" "proxy" {
  name          = data.docker_registry_image.proxy.name
  pull_triggers = [data.docker_registry_image.proxy.sha256_digest]
}

resource "docker_network" "backend" {
  name = "${var.environment}-${var.project_name}-backend"
}

resource "docker_container" "backend" {
  for_each = var.backends

  image = docker_image.backend.image_id
  name  = each.key

  restart = "unless-stopped"

  networks_advanced {
    name = docker_network.backend.name
  }
}

resource "docker_container" "proxy" {
  image = docker_image.proxy.image_id
  name  = "proxy"

  restart = "unless-stopped"

  depends_on = [docker_container.backend]

  wait         = true
  wait_timeout = 60

  upload {
    file    = "/etc/nginx/conf.d/default.conf"
    content = templatefile("${path.module}/nginx/default.conf.tftpl", { backends = sort(tolist(var.backends)) })
  }

  networks_advanced {
    name = docker_network.backend.name
  }

  ports {
    internal = 80
    external = 8080
    ip       = "127.0.0.1"
  }

  healthcheck {
    test         = ["CMD", "service", "nginx", "status"]
    interval     = "10s"
    timeout      = "3s"
    retries      = 3
    start_period = "5s"
  }
}
