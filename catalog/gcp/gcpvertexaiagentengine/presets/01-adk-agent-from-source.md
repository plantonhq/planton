# ADK Agent from Source

## Use Case

Deploy an Agent Development Kit agent from its source code: Vertex AI builds the source with its own Python build (installing `requirements.txt`, importing `root_agent` from `agent.py`), hosts it as an autoscaled service, and gives it the project's Vertex AI service agent as identity. The archive here is a working one-file agent; replace it with your own.

## When to Use

- Your first agent on Agent Engine
- ADK, LangChain, LangGraph, or LlamaIndex agents kept as source in a repository
- Teams that want a build from source rather than a container pipeline

## What This Creates

- An Agent Engine instance in `us-central1` built from the inline source archive on Python 3.12
- One to five instances, with `LOG_LEVEL=info` in the environment
- The default Vertex AI Reasoning Engine service agent as the agent's identity

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `spec.sourceCodeSpec.inlineSource.sourceArchive` | the sample agent | Your archive: `tar czf - -C <source-root> . \| base64`. |
| `spec.sourceCodeSpec.pythonSpec.entrypointModule` / `entrypointObject` | `agent` / `root_agent` | Where your agent object lives. |
| `spec.agentFramework` | `google-adk` | `langchain`, `langgraph`, `llama-index`, `ag2`, or your own. |
| `spec.deploymentSpec.minInstances` | `1` | `0` to scale to nothing between requests, at the cost of cold starts. |
| `spec.serviceAccount` | none | A `GcpServiceAccount` reference to run as a custom identity. |
| `deletionPolicy` | `DELETE` | `PREVENT` for an agent in production. |

A new archive redeploys the agent's code in place; the location and the encryption key are the only immutable fields. Builds run as Cloud Build in the project.
