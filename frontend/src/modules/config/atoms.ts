// Input: jotai atom primitives, bindings types
// Output: Config state atoms
// Position: Global state for configuration management

import type {
  ConfigModeResponse,
  ProfileDetail,
  ProfileSummary,
} from '@bindings/mcpd/internal/ui'
import { atom } from 'jotai'

// Config mode and path
export const configModeAtom = atom<ConfigModeResponse | null>(null)

// Profile list
export const profilesAtom = atom<ProfileSummary[]>([])

// Selected profile name
export const selectedProfileNameAtom = atom<string | null>(null)

// Selected profile detail
export const selectedProfileAtom = atom<ProfileDetail | null>(null)

// Caller mappings
export const callersAtom = atom<Record<string, string>>({})

// Loading states
export const configLoadingAtom = atom(false)
export const profileLoadingAtom = atom(false)
