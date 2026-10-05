terraform {
  required_version = ">= 1.0"

  required_providers {
    aws = {
      # One pessimistic pin, catalog-wide: every AWS module tracks the same
      # provider line, floored at the newest minor already released when the
      # monthly pin sweep last advanced it. The `~>` cap makes the next major
      # a deliberate catalog-wide decision, and floor-at-latest-released-minor
      # means the constraint never understates what any module's newest
      # argument needs. Only the sweep moves this line — never a single kind.
      #
      # Feature floor: aws_bedrockagent_knowledge_base's Kendra/SQL types
      # plus the Neptune Analytics / MongoDB Atlas / managed-OpenSearch /
      # S3 Vectors backends land in v6.27.0; Bedrock Data Automation
      # parsing in v6.29.0; the MANAGED type, the managed connector, and
      # the audio/video embedding blocks in v6.56.0 -- and v6.58.0 fixes
      # the v6.56.0 parsing-strategy validator regression, so an older
      # floor would reject valid BEDROCK_DATA_AUTOMATION configurations.
      source  = "hashicorp/aws"
      version = "~> 6.58"
    }
  }
}

provider "aws" {
  # Region and credentials are injected by the runtime as environment variables
  # (AWS_REGION + AWS_ACCESS_KEY_ID / AWS_SECRET_ACCESS_KEY / AWS_SESSION_TOKEN), resolved
  # from the IaC input's provider_config. For keyless (oidc)
  # connections the runtime performs the STS web-identity exchange and injects the resulting
  # short-lived credentials. Keep this block empty -- do not wire region or static keys here.
}
