# Custom Container Function

This preset creates a Function Compute v3 function that runs a custom Docker image. The container listens on port 8080 for HTTP requests with a health check endpoint, and includes lifecycle hooks for warm-up (initializer) and cleanup (pre-stop). This is the right choice when your function needs a runtime not natively supported by FC, has complex native dependencies, or uses a framework that expects a long-running HTTP server.

## When to Use

- Functions requiring runtimes not available as FC built-in runtimes (Rust, C++, Ruby, etc.)
- Applications with complex native library dependencies that are difficult to package as a ZIP
- Migrating existing containerized HTTP services to serverless with minimal changes
- AI/ML inference workloads with pre-trained models baked into the container image

## Key Configuration Choices

- **custom-container runtime** (`runtime: custom-container`) -- FC pulls and runs the specified container image. The container must start an HTTP server that FC forwards invocations to.
- **Handler set to "not-applicable"** (`handler: not-applicable`) -- The FC provider requires the handler field, but for container functions the actual entry point is the container's ENTRYPOINT/CMD. Use any non-empty string.
- **Higher compute** (`cpu: 2.0`, `memorySize: 4096`) -- Container functions typically need more resources than built-in runtime functions due to framework overhead and larger runtime footprints. Adjust based on profiling.
- **120-second timeout** (`timeout: 120`) -- Longer than the event handler default to accommodate container startup time and heavier processing. The first invocation after a cold start includes image pull and container initialization.
- **Health check** (`healthCheckConfig`) -- FC sends HTTP GET requests to `/healthz` to verify the container is ready before routing traffic. The 5-second initial delay accounts for application startup.
- **Lifecycle hooks** (`instanceLifecycleConfig`) -- The initializer runs once per instance for warm-up tasks (loading models, opening database connections). The pre-stop hook runs before an idle instance is reclaimed for cleanup (flushing buffers, closing connections).
- **No VPC or logging** -- Kept minimal to focus on the container deployment model. Add `vpcConfig` for private resource access and `logConfig` for SLS logging as needed.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<alibaba-cloud-region>` | Region code (e.g., `cn-hangzhou`) | Your deployment region |
| `<your-function-name>` | Function name (1-128 chars) | Your naming convention |
| `<your-container-image-uri>` | Full container image URI (e.g., `registry.cn-hangzhou.aliyuncs.com/my-ns/my-func:v1`) | Your container registry (ACR) |
| `<your-team>` | Team or business unit | Your organizational structure |

## Related Presets

- **01-event-handler** -- Use for simple event processing with a built-in runtime (Python, Node.js)
- **02-vpc-api-function** -- Use for API functions with VPC access using a built-in runtime
