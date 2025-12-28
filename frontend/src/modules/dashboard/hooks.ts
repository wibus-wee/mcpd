// Input: SWR, WailsService bindings
// Output: Dashboard data fetching hooks (useAppInfo, useTools, useResources, usePrompts)
// Position: Data fetching hooks for dashboard module

import { WailsService } from '@bindings/mcpd/internal/ui'
import useSWR from 'swr'

export function useAppInfo() {
  const swr = useSWR(
    'app-info',
    () => WailsService.GetInfo(),
    {
      revalidateOnFocus: false,
      dedupingInterval: 30000,
    },
  )
  return {
    ...swr,
    appInfo: swr.data ?? null,
  }
}

export function useTools() {
  const swr = useSWR(
    'tools',
    () => WailsService.ListTools(),
    {
      revalidateOnFocus: false,
      dedupingInterval: 10000,
    },
  )
  return {
    ...swr,
    tools: swr.data ?? [],
  }
}

export function useResources() {
  const swr = useSWR(
    'resources',
    async () => {
      const page = await WailsService.ListResources('')
      return page?.resources ?? []
    },
    {
      revalidateOnFocus: false,
      dedupingInterval: 10000,
    },
  )
  return {
    ...swr,
    resources: swr.data ?? [],
  }
}

export function usePrompts() {
  const swr = useSWR(
    'prompts',
    async () => {
      const page = await WailsService.ListPrompts('')
      return page?.prompts ?? []
    },
    {
      revalidateOnFocus: false,
      dedupingInterval: 10000,
    },
  )
  return {
    ...swr,
    prompts: swr.data ?? [],
  }
}
