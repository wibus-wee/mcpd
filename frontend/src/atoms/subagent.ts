// Input: jotai for state management
// Output: SubAgent configuration atoms for runtime settings
// Position: Global state atoms for SubAgent feature

import type { SubAgentConfigDetail } from '@bindings/mcpd/internal/ui'
import { SubAgentService } from '@bindings/mcpd/internal/ui'
import { atomWithRefresh } from 'jotai/utils'

// Atom to fetch runtime-level SubAgent config
export const subAgentConfigAtom = atomWithRefresh(async () => {
  try {
    const config = await SubAgentService.GetSubAgentConfig()
    return config as SubAgentConfigDetail
  }
  catch (error) {
    console.error('Failed to fetch SubAgent config:', error)
    return null
  }
})

// Atom to check if SubAgent infrastructure is available
export const isSubAgentAvailableAtom = atomWithRefresh(async () => {
  try {
    return await SubAgentService.IsSubAgentAvailable()
  }
  catch (error) {
    console.error('Failed to check SubAgent availability:', error)
    return false
  }
})
