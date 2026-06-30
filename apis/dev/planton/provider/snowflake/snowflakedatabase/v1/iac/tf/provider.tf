variable "snowflake_credential" {
  description = "Snowflake data"
  type = object({

    # snowflake account
    # The Snowflake account identifier, which may include the full account URL or just the account name.
    account = string

    # snowflake region
    # The Snowflake region, which specifies the location of the Snowflake instance.
    # Example values include 'us-west' or 'eu-central'.
    region = string

    # snowflake username
    # The username used to authenticate with Snowflake.
    # Ensure the username follows Snowflake's naming conventions.
    username = string

    # snowflake password
    # The password used to authenticate with Snowflake.
    # It is important to store this password securely and avoid hard-coding it in source code.
    password = string
  })
}

terraform {
  required_providers {
    snowflake = {
      source = "Snowflake-Labs/snowflake"
    }
  }
}

provider "snowflake" {
}
