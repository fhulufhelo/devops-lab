variable "project" {
  description = "Project name used in resource naming"
  type        = string
  default     = "devopslab"
}

variable "environment" {
  description = "Environment name (dev, staging, prod)"
  type        = string
  default     = "dev"
}

variable "location" {
  description = "Azure region for all resources"
  type        = string
  default     = "eastus"
}

variable "subscription_id" {
  description = "Azure subscription ID"
  type        = string
}

variable "tags" {
  description = "Tags applied to all resources"
  type        = map(string)
  default = {
    project     = "devops-lab"
    environment = "dev"
    managed_by  = "terraform"
    team        = "platform"
  }
}

variable "monthly_budget" {
  description = "Monthly budget in USD for cost alerts"
  type        = number
  default     = 50
}

variable "alert_email" {
  description = "Email address for budget alerts"
  type        = string
  default     = ""
}

variable "db_password" {
  description = "PostgreSQL administrator password"
  type        = string
  sensitive   = true
}
