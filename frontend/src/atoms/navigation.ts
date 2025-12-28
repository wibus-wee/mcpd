// Input: jotai atom primitives
// Output: Navigation state atoms (sidebarOpenAtom)
// Position: Global state for app navigation and sidebar

import { atom } from 'jotai'

export const sidebarOpenAtom = atom<boolean>(true)
