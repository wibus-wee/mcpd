// Input: jotai atom primitives
// Output: Navigation state atoms (activePageAtom, sidebarOpenAtom)
// Position: Global state for app navigation and sidebar

import { atom } from 'jotai'

export type PageId = 'dashboard' | 'tools' | 'logs' | 'settings'

export const activePageAtom = atom<PageId>('dashboard')
export const sidebarOpenAtom = atom<boolean>(true)
