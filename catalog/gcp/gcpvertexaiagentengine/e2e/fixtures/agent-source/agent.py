"""The smallest Agent Development Kit agent Agent Engine will build and run.

This is the source behind the inline archive in e2e/manifest.yaml and the
scenarios: a single LLM agent with no tools, served as an AdkApp, so the proof exercises the
build-from-source path (Cloud Build in the project, Vertex AI's Python
build) and the runtime without depending on any external service.

Rebuild the archive after editing (from this directory; the flags keep the
tarball free of owner names and macOS attributes so it is byte-stable):
    COPYFILE_DISABLE=1 tar --no-xattrs --uid 0 --gid 0 -czf /tmp/agent.tar.gz agent.py requirements.txt
    base64 -i /tmp/agent.tar.gz | tr -d '\n'
"""

from google.adk.agents import Agent
from vertexai import agent_engines

root_agent = Agent(
    name="planton_e2e_agent",
    model="gemini-3.5-flash",
    description="Answers a single question and stops.",
    instruction="You are a terse assistant. Answer in one sentence.",
)

# Agent Engine serves the entrypoint object through its query methods; a bare
# ADK agent has none, so the served object is the agent wrapped as an AdkApp.
app = agent_engines.AdkApp(agent=root_agent)
