// Shared tool schema parsing utilities
import type { ToolEntry } from '@bindings/mcpd/internal/ui'

export interface ToolSchema {
  name?: string
  description?: string
  inputSchema?: {
    type?: string
    properties?: Record<string, {
      type?: string
      description?: string
      enum?: string[]
      default?: unknown
    }>
    required?: string[]
  }
}

/**
 * Parses tool JSON from ToolEntry, handling both string and object formats
 */
export function parseToolJson(tool: ToolEntry): ToolSchema {
  try {
    const parsed = typeof tool.toolJson === 'string'
      ? JSON.parse(tool.toolJson)
      : tool.toolJson
    return { name: tool.name, ...parsed }
  } catch {
    return { name: tool.name }
  }
}

/**
 * Extracts description from tool schema
 */
export function parseToolDescription(tool: ToolEntry): string | undefined {
  const schema = parseToolJson(tool)
  return schema.description
}