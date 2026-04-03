# Budget alert — notifies when spending approaches the monthly limit
# Azure Cost Management budget scoped to the app resource group

data "azurerm_subscription" "current" {}

resource "azurerm_consumption_budget_resource_group" "main" {
  name              = "budget-${var.project}-${var.environment}"
  resource_group_id = azurerm_resource_group.main.id

  amount     = var.monthly_budget
  time_grain = "Monthly"

  time_period {
    start_date = "2026-04-01T00:00:00Z"
  }

  # Alert at 50% spend
  notification {
    operator       = "GreaterThanOrEqualTo"
    threshold      = 50
    threshold_type = "Actual"
    enabled        = true

    contact_roles = ["Owner"]
  }

  # Alert at 80% spend
  notification {
    operator       = "GreaterThanOrEqualTo"
    threshold      = 80
    threshold_type = "Actual"
    enabled        = true

    contact_roles = ["Owner"]
  }

  # Alert at 100% — budget exceeded
  notification {
    operator       = "GreaterThanOrEqualTo"
    threshold      = 100
    threshold_type = "Actual"
    enabled        = true

    contact_roles = ["Owner"]
  }

  # Forecast alert — warns before you actually hit the limit
  notification {
    operator       = "GreaterThanOrEqualTo"
    threshold      = 100
    threshold_type = "Forecasted"
    enabled        = true

    contact_roles = ["Owner"]
  }
}
