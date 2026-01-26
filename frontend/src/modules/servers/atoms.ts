// Input: jotai atom primitives, bindings types
// Output: Combined atoms for servers module
// Position: Global state for servers management

import type {
  ConfigModeResponse,
  ServerDetail,
  ToolEntry,
} from '@bindings/mcpd/internal/ui'
import { atom } from 'jotai'

export interface ServerGroup {
  id: string
  specKey: string
  serverName: string
  tools: ToolEntry[]
  tags: string[]
  hasToolData: boolean
  specDetail?: ServerDetail
}

// Config mode and path
export const configModeAtom = atom<ConfigModeResponse | null>(null)

// Selected server detail
export const selectedServerAtom = atom<ServerDetail | null>(null)

// Loading states
export const configLoadingAtom = atom(false)
export const serverLoadingAtom = atom(false)
