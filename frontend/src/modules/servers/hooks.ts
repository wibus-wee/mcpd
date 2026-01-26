// Input: Config hooks, tools hooks, SWR
// Output: Combined data hooks for servers module
// Position: Data layer for unified servers module

import type {
  ActiveClient,
  ConfigModeResponse,
  ServerDetail,
  ServerInitStatus,
  ServerRuntimeStatus,
  ServerSummary,
  ToolEntry,
} from '@bindings/mcpd/internal/ui'
import { ConfigService, DiscoveryService, RuntimeService, ServerService } from '@bindings/mcpd/internal/ui'
import { useSetAtom } from 'jotai'
import { useCallback, useEffect, useMemo, useState } from 'react'
import useSWR from 'swr'

import { withSWRPreset } from '@/lib/swr-config'

import {
  activeClientsAtom,
  configModeAtom,
  selectedServerAtom,
  serversAtom,
} from './atoms'
import { reloadConfig } from './lib/reload-config'

export function useConfigMode() {
  const setConfigMode = useSetAtom(configModeAtom)

  const { data, error, isLoading, mutate } = useSWR<ConfigModeResponse>(
    'config-mode',
    () => ConfigService.GetConfigMode(),
  )

  useEffect(() => {
    if (data) {
      setConfigMode(data)
    }
  }, [data, setConfigMode])

  return { data, error, isLoading, mutate }
}

export function useServers() {
  const setServers = useSetAtom(serversAtom)

  const { data, error, isLoading, mutate } = useSWR<ServerSummary[]>(
    'servers',
    () => ServerService.ListServers(),
    withSWRPreset('fastRealtime', {
      revalidateOnMount: true,
    }),
  )

  useEffect(() => {
    if (data) {
      setServers(data)
    }
  }, [data, setServers])

  return { data, error, isLoading, mutate }
}

export function useServer(name: string | null) {
  const setSelectedServer = useSetAtom(selectedServerAtom)

  const { data, error, isLoading, mutate } = useSWR<ServerDetail | null>(
    name ? ['server', name] : null,
    () => (name ? ServerService.GetServer(name) : null),
  )

  useEffect(() => {
    if (data !== undefined) {
      setSelectedServer(data)
    }
  }, [data, setSelectedServer])

  return { data, error, isLoading, mutate }
}

export function useServerDetails(servers: ServerSummary[] | undefined) {
  const serverNames = servers?.map(server => server.name) ?? []

  const { data, error, isLoading, mutate } = useSWR<ServerDetail[]>(
    serverNames.length > 0 ? ['server-details', ...serverNames] : null,
    async () => {
      const results = await Promise.all(
        serverNames.map(name => ServerService.GetServer(name)),
      )

      return results.filter(
        (server): server is ServerDetail => server !== null,
      )
    },
  )

  return { data, error, isLoading, mutate }
}

export function useClients() {
  const setActiveClients = useSetAtom(activeClientsAtom)

  const { data, error, isLoading, mutate } = useSWR<ActiveClient[]>(
    'active-clients',
    () => RuntimeService.GetActiveClients(),
    withSWRPreset('fastCached', {
      refreshInterval: 3000,
      dedupingInterval: 3000,
    }),
  )

  useEffect(() => {
    if (data) {
      setActiveClients(data)
    }
  }, [data, setActiveClients])

  return { data, error, isLoading, mutate }
}

export function useOpenConfigInEditor() {
  const [isOpening, setIsOpening] = useState(false)
  const [error, setError] = useState<Error | null>(null)

  const openInEditor = useCallback(async () => {
    setIsOpening(true)
    setError(null)
    try {
      await ConfigService.OpenConfigInEditor()
    }
    catch (err) {
      setError(err instanceof Error ? err : new Error(String(err)))
    }
    finally {
      setIsOpening(false)
    }
  }, [])

  return { openInEditor, isOpening, error }
}

export function useRuntimeStatus() {
  return useSWR<ServerRuntimeStatus[]>(
    'runtime-status',
    () => RuntimeService.GetRuntimeStatus(),
    withSWRPreset('fastCached', {
      refreshInterval: 2000,
      dedupingInterval: 2000,
    }),
  )
}

export function useServerInitStatus() {
  return useSWR<ServerInitStatus[]>(
    'server-init-status',
    () => RuntimeService.GetServerInitStatus(),
    withSWRPreset('fastCached', {
      refreshInterval: 2000,
      dedupingInterval: 2000,
    }),
  )
}

export interface ServerGroup {
  id: string
  specKey: string
  serverName: string
  tools: ToolEntry[]
  tags: string[]
  hasToolData: boolean
  specDetail?: ServerDetail
}

export function useToolsByServer() {
  const {
    data: tools,
    isLoading: toolsLoading,
    error: toolsError,
  } = useSWR<ToolEntry[]>(
    'tools',
    () => DiscoveryService.ListTools(),
    withSWRPreset('cached', {
      refreshInterval: 10000,
      dedupingInterval: 10000,
    }),
  )

  const {
    data: runtimeStatus,
    isLoading: runtimeLoading,
    error: runtimeError,
  } = useRuntimeStatus()
  const {
    data: servers,
    isLoading: serversLoading,
    error: serversError,
  } = useServers()
  const {
    data: serverDetails,
    isLoading: detailsLoading,
    error: detailsError,
  } = useServerDetails(servers)

  const toolsBySpecKey = useMemo(() => {
    const map = new Map<string, ToolEntry[]>()
    if (!tools) return map

    tools.forEach((tool) => {
      const specKey = tool.specKey || tool.serverName || tool.name
      if (!specKey) return
      const bucket = map.get(specKey)
      if (bucket) {
        bucket.push(tool)
      }
      else {
        map.set(specKey, [tool])
      }
    })

    return map
  }, [tools])

  const serversFromSummaries = useMemo(() => {
    const map = new Map<string, { summary: ServerSummary, tags: string[] }>()
    if (!servers) return map

    servers.forEach((summary) => {
      if (!summary.specKey) return
      map.set(summary.specKey, {
        summary,
        tags: summary.tags ?? [],
      })
    })

    return map
  }, [servers])

  const serverMap = useMemo(() => {
    const map = new Map<string, ServerGroup>()

    const ensureServer = (
      specKey: string,
      serverName?: string,
      specDetail?: ServerDetail,
      tags?: string[],
    ) => {
      if (!specKey) return null
      const existing = map.get(specKey)
      if (existing) {
        if (!existing.serverName && serverName) {
          existing.serverName = serverName
        }
        if (!existing.specDetail && specDetail) {
          existing.specDetail = specDetail
        }
        if (tags && tags.length > 0 && existing.tags.length === 0) {
          existing.tags = tags
        }
        return existing
      }
      const entry: ServerGroup = {
        id: specKey,
        specKey,
        serverName: serverName || specKey,
        tools: [],
        tags: tags ?? [],
        hasToolData: false,
        specDetail,
      }
      map.set(specKey, entry)
      return entry
    }

    serversFromSummaries.forEach(({ summary, tags }, specKey) => {
      ensureServer(specKey, summary.name, undefined, tags)
    })

    serverDetails?.forEach((detail) => {
      ensureServer(detail.specKey, detail.name, detail, detail.tags ?? [])
    })

    runtimeStatus?.forEach((status) => {
      ensureServer(status.specKey, status.serverName)
    })

    toolsBySpecKey.forEach((toolList, specKey) => {
      const entry = ensureServer(specKey)
      if (entry) {
        entry.tools = toolList
        entry.hasToolData = true
      }
    })

    return map
  }, [runtimeStatus, serverDetails, serversFromSummaries, toolsBySpecKey])

  const groupedServers = useMemo(() => {
    return Array.from(serverMap.values()).sort((a, b) =>
      a.serverName.localeCompare(b.serverName),
    )
  }, [serverMap])

  const isLoading
    = toolsLoading || serversLoading || detailsLoading || runtimeLoading
  const error = toolsError || serversError || detailsError || runtimeError

  return {
    servers: groupedServers,
    serverMap,
    isLoading,
    error,
    runtimeStatus,
  }
}

type ErrorHandler = (title: string, description: string) => void
type SuccessHandler = (title: string, description: string) => void

export function useServerOperation(
  canEdit: boolean,
  mutateServers: () => Promise<any>,
  mutateServer?: () => Promise<any>,
  onDeleted?: (serverName: string) => void,
  errorHandler?: ErrorHandler,
  successHandler?: SuccessHandler,
) {
  const [isWorking, setIsWorking] = useState(false)

  const executeOperation = useCallback(async (
    operation: 'toggle' | 'delete',
    server: { name: string; disabled?: boolean },
  ) => {
    if (!canEdit || isWorking) return
    setIsWorking(true)

    try {
      if (operation === 'toggle') {
        await ServerService.SetServerDisabled({
          server: server.name,
          disabled: !server.disabled,
        })
      } else if (operation === 'delete') {
        await ServerService.DeleteServer({ server: server.name })
      }

      const reloadResult = await reloadConfig()
      if (!reloadResult.ok) {
        errorHandler?.('Reload failed', reloadResult.message)
        return
      }

      await Promise.all([
        mutateServers(),
        mutateServer?.(),
      ])

      if (operation === 'toggle') {
        successHandler?.(
          server.disabled ? 'Server enabled' : 'Server disabled',
          'Changes applied.',
        )
      } else if (operation === 'delete') {
        successHandler?.('Server deleted', 'Changes applied.')
        onDeleted?.(server.name)
      }
    } catch (err) {
      const message = err instanceof Error ? err.message : `${operation} failed.`
      errorHandler?.(`${operation === 'toggle' ? 'Update' : 'Delete'} failed`, message)
    } finally {
      setIsWorking(false)
    }
  }, [canEdit, isWorking, mutateServers, mutateServer, onDeleted, errorHandler, successHandler])

  const toggleDisabled = useCallback((server: { name: string; disabled?: boolean }) =>
    executeOperation('toggle', server), [executeOperation])

  const deleteServer = useCallback((server: { name: string; disabled?: boolean }) =>
    executeOperation('delete', server), [executeOperation])

  return {
    isWorking,
    toggleDisabled,
    deleteServer,
  }
}
