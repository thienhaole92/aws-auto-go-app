# -----------------------------------------------------------------
# TERRAFORM AND PROVIDER CONFIGURATION
# -----------------------------------------------------------------

terraform {
  required_providers {
    postgresql = {
      source  = "cyrilgdn/postgresql"
      version = ">= 1.26.0"
    }
  }
}

provider "postgresql" {
  host      = var.host
  port      = var.port
  username  = var.admin_user
  password  = var.admin_password
  sslmode   = "disable"
  superuser = true # Must be true to use the admin credentials for creation
}

# -----------------------------------------------------------------
# INPUT VARIABLES
# -----------------------------------------------------------------

variable "host" {
  description = "The hostname or IP address of the PostgreSQL server."
  type        = string
  default     = "postgres"
}

variable "port" {
  description = "The port of the PostgreSQL server."
  type        = number
  default     = 5432
}

variable "admin_user" {
  description = "The administrative user (super_user) used by Terraform to provision resources."
  type        = string
  default     = "postgres"
}

variable "admin_password" {
  description = "The password for the administrative user."
  type        = string
  default     = "password"
  sensitive   = true
}

variable "db_name" {
  description = "The name for the new PostgreSQL database."
  type        = string
}

variable "db_user" {
  description = "The name for the new PostgreSQL application user/role."
  type        = string
}

variable "db_password" {
  description = "The password for the new PostgreSQL application user/role."
  type        = string
  sensitive   = true
}

variable "schema_name" {
  description = "The name of the new schema to create within the database."
  type        = string
  default     = "app_schema"
}

# -----------------------------------------------------------------
# RESOURCES
# -----------------------------------------------------------------

# 1. Create the dedicated database
resource "postgresql_database" "app_db" {
  name  = var.db_name
  owner = var.admin_user # Initially owned by the admin user
}

# 2. Create the application user/role
resource "postgresql_role" "app_user" {
  name      = var.db_user
  password  = var.db_password
  login     = true
  superuser = false
}

# 3. Create the schema within the new database
resource "postgresql_schema" "app_schema" {
  database = postgresql_database.app_db.name
  name     = var.schema_name
  owner    = postgresql_role.app_user.name # Make the new app user the owner of the schema
}

# 4. Grant schema usage and connect privileges to the new user

# Grant USAGE and CREATE privilege on the new schema to the new user
resource "postgresql_grant" "schema_usage_grant" {
  database    = postgresql_database.app_db.name
  role        = postgresql_role.app_user.name
  schema      = postgresql_schema.app_schema.name
  object_type = "schema"
  privileges  = ["USAGE", "CREATE"]
}

# Grant CONNECT privilege on the database to the new user
resource "postgresql_grant" "db_connect_grant" {
  database    = postgresql_database.app_db.name
  role        = postgresql_role.app_user.name
  object_type = "database"
  privileges  = ["CONNECT", "CREATE"]

  # Ensure this grant is applied only after the database is created
  depends_on = [
    postgresql_database.app_db,
  ]
}

# 5. Grant default privileges (for future objects) to the new user (CRUD)
# These grants ensure the user can interact with objects created later by the same user.

# Grant CRUD privileges on future tables created in the schema
resource "postgresql_default_privileges" "table_crud_privileges" {
  database    = postgresql_database.app_db.name
  role        = postgresql_role.app_user.name
  schema      = postgresql_schema.app_schema.name
  owner       = postgresql_role.app_user.name
  object_type = "table"
  privileges  = ["SELECT", "INSERT", "UPDATE", "DELETE"]
}

# Grant necessary privileges on future sequences (for auto-incrementing IDs)
resource "postgresql_default_privileges" "sequence_privileges" {
  database    = postgresql_database.app_db.name
  role        = postgresql_role.app_user.name
  schema      = postgresql_schema.app_schema.name
  owner       = postgresql_role.app_user.name
  object_type = "sequence"
  privileges  = ["USAGE", "SELECT"]
}

# Grant EXECUTE privileges on future functions (stored procedures)
resource "postgresql_default_privileges" "function_execute_privileges" {
  database    = postgresql_database.app_db.name
  role        = postgresql_role.app_user.name
  schema      = postgresql_schema.app_schema.name
  owner       = postgresql_role.app_user.name
  object_type = "function"
  privileges  = ["EXECUTE"]
}

# -----------------------------------------------------------------
# OUTPUTS
# -----------------------------------------------------------------

output "db_name" {
  description = "The name of the created database."
  value       = postgresql_database.app_db.name
}

output "db_user" {
  description = "The name of the created database user."
  value       = postgresql_role.app_user.name
}

output "db_password" {
  description = "The password for the created database user."
  value       = postgresql_role.app_user.password
  sensitive   = true
}

output "schema_name" {
  description = "The name of the created schema."
  value       = postgresql_schema.app_schema.name
}

output "connection_string" {
  description = "A standard connection string for the new user and database, including the default schema in the search path."
  value       = "postgresql://${postgresql_role.app_user.name}:${postgresql_role.app_user.password}@${var.host}:${var.port}/${postgresql_database.app_db.name}?sslmode=disable&options=-c%20search_path%3D${postgresql_schema.app_schema.name}"
  sensitive   = true
}
