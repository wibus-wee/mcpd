// Input: WailsService bindings, SWR
// Output: Active callers SWR hook
// Position: Shared hook for active caller registrations

import type { ActiveCaller } from '@bindings/mcpd/internal/ui'
import { WailsService } from '@bindings/mcpd/internal/ui'
import useSWR from 'swr'

export const activeCallersKey = 'active-callers'

export function useActiveCallers() {
  return useSWR<ActiveCaller[]>(activeCallersKey, () => WailsService.GetActiveCallers(), {
    revalidateOnFocus: false,
    dedupingInterval: 5000,
  })
}
