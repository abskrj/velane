package agentdocs

// Doc is the canonical MCP-facing guide for AI agent workflows on Velane executors.
const Doc = `# Velane agent frameworks

Velane Bun and Python executors ship with agent frameworks pre-installed. **Use them for any LLM/agent workflow** (chat, tool use, multi-step reasoning). Do **not** hand-roll OpenAI/Anthropic fetch loops or custom agent classes.

| Language | Framework | Import |
|---|---|---|
| bun | Mastra | ` + "`import { Agent } from '@mastra/core/agent'`" + ` |
| python | LangGraph | ` + "`from langgraph.graph import StateGraph`" + ` |

Call **get_agent_framework_docs** before writing agent workflow code.

## Bun — Mastra (required for agent workflows)

` + "```typescript" + `
import { Agent } from '@mastra/core/agent'

export default async function handler(input: Record<string, unknown>) {
  const agent = new Agent({
    name: 'workflow-agent',
    instructions: 'You are a helpful assistant.',
    model: 'openai/gpt-4o-mini',
  })
  const result = await agent.generate('Summarize: ' + String(input.topic ?? ''))
  return { text: result.text }
}
` + "```" + `

Provider API keys (e.g. OPENAI_API_KEY) must be tenant secrets — injected as env vars at invoke time.

Wire OAuth integrations inside Mastra tool ` + "`execute`" + ` functions using ` + "`import { integration } from '@velane/integrations'`" + `.

## Python — LangGraph (required for agent workflows)

` + "```python" + `
import os
from langchain_openai import ChatOpenAI
from langgraph.graph import StateGraph

def handler(input: dict) -> dict:
    model = ChatOpenAI(model="gpt-4o-mini", api_key=os.environ["OPENAI_API_KEY"])
  # Build and run your graph; return structured output.
    return {"message": "agent result"}
` + "```" + `

## Runtime limits for agent workflows

Agent imports need more memory. When calling update_draft, set for example:
- timeout_ms: 120000
- max_memory_mb: 512 (or 1024 for large graphs)
- max_cpu_percent: 50

## MCP workflow for agents

1. Call get_agent_framework_docs (this document).
2. create_workflow with language bun (Mastra) or python (LangGraph).
3. update_draft with framework-based code — not custom agent loops.
4. invoke_workflow in dev; inspect output and stderr.
5. publish_workflow after validation.
`
