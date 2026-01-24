// Input: RuntimeService bindings, SWR
// Output: Active callers SWR hook
// Position: Shared hook for active caller registrations

import type { ActiveCaller } from '@bindings/mcpd/internal/ui'
import { RuntimeService } from '@bindings/mcpd/internal/ui'
import useSWR from 'swr'

import { swrPresets } from '@/lib/swr-config'

export const activeCallersKey = 'active-callers'

export function useActiveCallers() {
  return useSWR<ActiveCaller[]>(activeCallersKey, () => RuntimeService.GetActiveCallers(), swrPresets.fastCached)
}
