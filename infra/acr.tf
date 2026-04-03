# Azure Container Registry — stores your Docker images
resource "azurerm_container_registry" "main" {
  name                = "acr${var.project}${var.environment}"
  resource_group_name = azurerm_resource_group.main.name
  location            = azurerm_resource_group.main.location
  sku                 = "Basic" # cheapest tier, sufficient for learning
  admin_enabled       = true    # simpler auth for Container Apps

  tags = var.tags
}
