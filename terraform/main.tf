module "service_stack" {
  source = "./modules/service_stack"

  environment     = var.environment
  project_name    = var.project_name
  backend_image   = var.backend_image
  proxy_image     = var.proxy_image
  proxy_host_port = var.proxy_host_port
  backends        = var.backends
}

moved {
  from = data.docker_registry_image.backend
  to   = module.service_stack.data.docker_registry_image.backend
}

moved {
  from = data.docker_registry_image.proxy
  to   = module.service_stack.data.docker_registry_image.proxy
}

moved {
  from = docker_image.backend
  to   = module.service_stack.docker_image.backend
}

moved {
  from = docker_image.proxy
  to   = module.service_stack.docker_image.proxy
}

moved {
  from = docker_network.backend
  to   = module.service_stack.docker_network.backend
}

moved {
  from = docker_container.backend
  to   = module.service_stack.docker_container.backend
}

moved {
  from = docker_container.proxy
  to   = module.service_stack.docker_container.proxy
}
