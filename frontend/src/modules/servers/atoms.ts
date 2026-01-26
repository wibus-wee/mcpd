// Input: jotai atom primitives, bindings types
// Output: Combined atoms for servers module
// Position: Global state for servers management

import type {
  ActiveClient,
  ConfigModeResponse,
  ServerDetail,
  ServerSummary,
} from '@bindings/mcpd/internal/ui'
import { atom } from 'jotai'

// Config mode and path
export const configModeAtom = atom<ConfigModeResponse | null>(null)

// Server list
export const serversAtom = atom<ServerSummary[]>([])

// Selected server name
export const selectedServerNameAtom = atom<string | null>(null)

// Selected server detail
export const selectedServerAtom = atom<ServerDetail | null>(null)

// Active clients
export const activeClientsAtom = atom<ActiveClient[]>([])

// Loading states
export const configLoadingAtom = atom(false)
export const serverLoadingAtom = atom(false)

// Server selection and tab state
export const activeTabAtom = atom<'overview' | 'tools' | 'configuration'>('overview')
