# Azure Database for PostgreSQL Flexible Server

resource "azurerm_postgresql_flexible_server" "main" {
  name                          = "psql-devlab-413c15"
  resource_group_name           = azurerm_resource_group.main.name
  location                      = "centralus" # eastus restricted for PostgreSQL on this subscription
  version                       = "16"
  administrator_login           = "devopsadmin"
  administrator_password        = var.db_password
  sku_name                      = "B_Standard_B1ms" # cheapest: 1 vCPU, 2 GiB RAM
  storage_mb                    = 32768              # 32 GB minimum

  # Allow Azure services (Container Apps) to connect
  public_network_access_enabled = true
  zone                          = "2"

  tags = var.tags
}

resource "azurerm_postgresql_flexible_server_database" "main" {
  name      = "devopslab"
  server_id = azurerm_postgresql_flexible_server.main.id
  charset   = "UTF8"
  collation = "en_US.utf8"
}

# Allow Azure services (Container Apps) to reach the database
resource "azurerm_postgresql_flexible_server_firewall_rule" "azure_services" {
  name             = "AllowAllAzureServicesAndResourcesWithinAzureIps_2026-4-3_21-26-10"
  server_id        = azurerm_postgresql_flexible_server.main.id
  start_ip_address = "0.0.0.0"
  end_ip_address   = "0.0.0.0"
}
