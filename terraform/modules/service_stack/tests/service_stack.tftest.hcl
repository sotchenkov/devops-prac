mock_provider "docker" {}

run "dev_resource_names" {
  command = plan

  variables {
    environment  = "dev"
    project_name = "homelab"
    backends = ["api"]
  }

  assert {
    condition     = docker_container.backend["api"].name == "dev-homelab-api"
    error_message = "Wrong backend container name"
  }

  assert {
    condition     = docker_network.backend.name == "dev-homelab-backend"
    error_message = "Wrong docker network name"
  }

  assert {
    condition     = docker_container.proxy.name == "dev-homelab-proxy"
    error_message = "Wrong proxy container name"
  }
}

run "empty_backends_rejected" {
  command = plan

  variables {
    environment  = "dev"
    project_name = "homelab"
    backends     = []
  }

  expect_failures = [
    var.backends,
  ]
}

run "unsupported_environment_rejected" {
  command = plan

  variables {
    environment  = "staging"
    project_name = "homelab"
    backends     = ["api"]
  }

  expect_failures = [
    var.environment,
  ]
}