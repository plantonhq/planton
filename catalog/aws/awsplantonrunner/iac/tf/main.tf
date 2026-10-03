# AWS Planton Runner Terraform Module
#
# Provisions a standing Planton runner appliance on ECS Fargate: an
# always-on, outbound-only worker that executes deploy and cloud
# operations from inside the VPC -- the piece that makes private
# endpoints (most notably private Kubernetes cluster APIs) deployable
# and operable. The subnets (and any extra security groups or the
# runtime role) are referenced resources -- the module never creates or
# mutates them.

# ── Token secret ────────────────────────────────────────────────────────
#
# The container never sees the runner token in its launch configuration:
# the task definition carries only this secret's ARN, and the ECS agent
# fetches the value at task start through the execution role -- so reading
# the task definition (a common, low-sensitivity permission) reveals
# nothing. The token authorizes JOINING and is never the runner's
# identity: the runner presents it once per enrollment, registers itself,
# and receives its own individually revocable identity from the control
# plane.
#
# Zero recovery window: the token is re-mintable credential material, not
# data. Secrets Manager's default 30-day soft-delete would block
# re-creating a same-named runner for a month after a destroy, with no
# compensating benefit for a value that can simply be re-issued.
resource "aws_secretsmanager_secret" "token" {
  name                    = "${local.runner_name}-token"
  description             = "Runner token for Planton runner '${local.runner_name}'"
  recovery_window_in_days = 0
  tags                    = local.aws_tags
}

resource "aws_secretsmanager_secret_version" "token" {
  secret_id     = aws_secretsmanager_secret.token.id
  secret_string = var.spec.token
}

# ── IAM: two roles, strictly separated ──────────────────────────────────
#
# The execution role is the SETUP identity (the ECS agent pulls the
# image, writes logs, and reads the one credentials secret); the runtime
# role is the runner's OWN identity while executing work -- the seam
# keyless cloud access rides. Collapsing them would hand the runner's
# workloads the infrastructure permissions and vice versa.

data "aws_iam_policy_document" "ecs_tasks_trust" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["ecs-tasks.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "execution" {
  name               = "${local.runner_name}-exec"
  assume_role_policy = data.aws_iam_policy_document.ecs_tasks_trust.json
  description        = "ECS task execution role for Planton runner '${local.runner_name}'"
  tags               = local.aws_tags
}

# Baseline execution permissions: pull images, write CloudWatch logs.
resource "aws_iam_role_policy_attachment" "execution_managed" {
  role       = aws_iam_role.execution.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

# Secret access is an inline policy scoped to exactly the one secret ARN
# -- the managed execution policy deliberately grants no secret
# permissions.
resource "aws_iam_role_policy" "execution_secret_read" {
  name = "token-secret-read"
  role = aws_iam_role.execution.id
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect   = "Allow"
        Action   = ["secretsmanager:GetSecretValue"]
        Resource = aws_secretsmanager_secret.token.arn
      }
    ]
  })
}

# Created permissionless when no task_role is referenced, so the identity
# seam always exists: permissions can be granted later without replacing
# the runner.
resource "aws_iam_role" "runtime" {
  count              = var.spec.task_role != "" ? 0 : 1
  name               = "${local.runner_name}-runtime"
  assume_role_policy = data.aws_iam_policy_document.ecs_tasks_trust.json
  description        = "Runtime identity for Planton runner '${local.runner_name}' -- grant it the permissions keyless operations need"
  tags               = local.aws_tags
}

# ── Logs ────────────────────────────────────────────────────────────────
#
# The runner's logs are the audit trail of every operation it executed --
# retention is explicit, never CloudWatch's never-expire default.
resource "aws_cloudwatch_log_group" "runner" {
  name              = local.log_group_name
  retention_in_days = var.spec.log_retention_days
  tags              = local.aws_tags
}

# ── Networking ──────────────────────────────────────────────────────────
#
# The VPC is derived from the first referenced subnet rather than asked
# for on the spec: a separate vpc field could only ever agree with the
# subnets or contradict them.
data "aws_subnet" "first" {
  id = var.spec.subnets[0]
}

# Outbound-only: the runner initiates every connection it uses (control
# plane, work queue, image pulls), so the group carries the permissive
# egress rule and NO inbound rules at all. Private targets that admit
# traffic by source security group reference this group's id (published
# as an output) to trust the runner.
resource "aws_security_group" "runner" {
  name = local.runner_name
  # SG descriptions reject quote characters (the API's allowed set is
  # a-zA-Z0-9. _-:/()#,@[]+=&;{}!$*), so the name is embedded bare.
  description = "Planton runner ${local.runner_name} -- outbound only, no inbound"
  vpc_id      = data.aws_subnet.first.vpc_id

  egress {
    protocol    = "-1"
    from_port   = 0
    to_port     = 0
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = local.aws_tags
}

# ── Compute ─────────────────────────────────────────────────────────────
#
# A dedicated cluster per runner: clusters are free scheduling boundaries
# (no cost until tasks run), and a dedicated one keeps the appliance
# self-contained -- its teardown removes everything, and no
# shared-cluster coupling can block or complicate it.
resource "aws_ecs_cluster" "runner" {
  name = local.runner_name
  tags = local.aws_tags
}

resource "aws_ecs_task_definition" "runner" {
  family                   = local.runner_name
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"

  # Fargate takes task-level sizing as strings (they predate typed
  # sizes); the spec's CEL already guarantees a valid cpu/memory pairing.
  cpu    = tostring(var.spec.cpu)
  memory = tostring(var.spec.memory)

  execution_role_arn = aws_iam_role.execution.arn
  task_role_arn      = local.task_role_arn

  container_definitions = jsonencode([
    {
      name      = "planton-runner"
      image     = "${var.spec.image_repository}:${var.spec.runner_version}"
      command   = ["start"]
      essential = true

      # The runner's listening ports: the gRPC/CloudOps server and the
      # health server. Both are private to the task ENI -- the security
      # group admits no inbound traffic; the CloudOps channel reaches the
      # runner through its own outbound-initiated tunnel.
      portMappings = [
        { containerPort = 50051, protocol = "tcp" },
        { containerPort = 8093, protocol = "tcp" }
      ]

      # The runner's environment contract: the name it registers itself
      # under, the control-plane endpoint when one is declared (omitted,
      # the runner's built-in hosted default applies), and the base
      # process settings. No EXECUTION_MODE: only the control plane knows
      # whether its work queue is reachable, so the runner derives its
      # mode from the identity the join returns -- a mode knob here would
      # silently strip capability.
      environment = concat(
        [
          { name = "PLANTON_RUNNER_NAME", value = local.registration_name },
          { name = "PORT", value = "50051" },
          { name = "LOG_LEVEL", value = "info" }
        ],
        try(var.spec.control_plane_endpoint, "") != "" ? [
          { name = "PLANTON_RUNNER_ENDPOINT", value = var.spec.control_plane_endpoint }
        ] : []
      )

      # The runner token arrives as an env var resolved by the ECS agent
      # at task start via the execution role -- never as plaintext in
      # this document. The token is only read at join, so rotating the
      # secret needs no task restart: running tasks keep serving on their
      # minted identity, and the next task replacement joins with the new
      # value.
      secrets = [
        { name = "PLANTON_RUNNER_TOKEN", valueFrom = aws_secretsmanager_secret.token.arn }
      ]

      logConfiguration = {
        logDriver = "awslogs"
        options = {
          "awslogs-group"         = local.log_group_name
          "awslogs-region"        = var.spec.region
          "awslogs-stream-prefix" = "runner"
        }
      }
    }
  ])

  tags = local.aws_tags

  # A missing log group fails at task START (not registration) -- this
  # explicit edge is what prevents that failure mode.
  depends_on = [aws_cloudwatch_log_group.runner]
}

# Exactly one runner per registration: the registration's work queue
# serializes operations, so scaling execution capacity means more
# runners, never more copies of this one. The service deliberately does
# NOT gate on steady state and carries no deployment circuit breaker:
# ECS reports the service ACTIVE independently of task health, and a
# runner whose control plane is momentarily unreachable must still
# deploy and destroy cleanly -- its readiness contract is the work
# queue, not ECS task liveness.
resource "aws_ecs_service" "runner" {
  name            = local.runner_name
  cluster         = aws_ecs_cluster.runner.arn
  task_definition = aws_ecs_task_definition.runner.arn
  desired_count   = 1
  launch_type     = "FARGATE"

  network_configuration {
    subnets          = var.spec.subnets
    security_groups  = local.security_group_ids
    assign_public_ip = var.spec.assign_public_ip
  }

  tags = local.aws_tags
}
