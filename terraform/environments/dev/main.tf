module "service_stack" {
  source = "../../modules/service_stack"

  environment     = var.environment
  project_name    = var.project_name
  backend_image   = var.backend_image
  proxy_image     = var.proxy_image
  proxy_host_port = var.proxy_host_port
  backends        = var.backends
}