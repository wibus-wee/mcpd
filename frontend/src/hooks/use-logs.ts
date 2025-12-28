// Input: SWR hooks
// Output: LogEntry type and useLogs hook for log stream cache
// Position: Shared log state accessors for app-wide logging UI

import useSWR from 'swr'

export type LogSource = 'core' | 'downstream' | 'ui' | 'unknown'

export interface LogEntry {
  id: string
  timestamp: Date
  level: 'debug' | 'info' | 'warn' | 'error'
  message: string
  source: LogSource
  logger?: string
  serverType?: string
  stream?: string
  fields: Record<string, unknown>
}

export const logsKey = 'logs'
export const maxLogEntries = 1000

export function useLogs() {
  const swr = useSWR<LogEntry[]>(logsKey, null, {
    fallbackData: [],
    revalidateIfStale: false,
    revalidateOnFocus: false,
    revalidateOnReconnect: false,
  })

  return {
    ...swr,
    logs: swr.data ?? [],
  }
}
