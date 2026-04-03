output "resource_group_name" {
  description = "Name of the resource group"
  value       = azurerm_resource_group.main.name
}

output "acr_login_server" {
  description = "ACR login server URL (used in CI/CD to push images)"
  value       = azurerm_container_registry.main.login_server
}

output "acr_name" {
  description = "ACR name"
  value       = azurerm_container_registry.main.name
}

output "backend_url" {
  description = "Backend Container App URL"
  value       = "https://${azurerm_container_app.backend.ingress[0].fqdn}"
}

output "frontend_url" {
  description = "Frontend Container App URL"
  value       = "https://${azurerm_container_app.frontend.ingress[0].fqdn}"
}

output "container_app_environment_id" {
  description = "Container Apps Environment ID"
  value       = azurerm_container_app_environment.main.id
}
