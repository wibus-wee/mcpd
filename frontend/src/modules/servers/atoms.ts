// Input: jotai atom primitives
// Output: Atoms for server selection and tab state
// Position: State management for servers module

import { atom } from 'jotai'

export const selectedServerNameAtom = atom<string | null>(null)
export const activeTabAtom = atom<'overview' | 'tools' | 'configuration'>('overview')
