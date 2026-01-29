// Input: Hooks, child components, UI primitives
// Output: ServersPage component - Master-Detail layout with resizable panels
// Position: Main page for servers module

import type { ServerDetail, ServerRuntimeStatus } from '@bindings/mcpd/internal/ui'
import { useNavigate } from '@tanstack/react-router'
import { useSetAtom } from 'jotai'
import { PlusIcon, SearchIcon } from 'lucide-react'
import { m } from 'motion/react'
import { useEffect, useMemo, useState } from 'react'

import { RefreshButton } from '@/components/custom/refresh-button'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Separator } from '@/components/ui/separator'
import { useResizable } from '@/hooks/use-resizable'
import { Spring } from '@/lib/spring'
import { cn } from '@/lib/utils'
import { ImportMcpServersSheet } from '@/modules/servers/components/import-mcp-servers-sheet'

import { selectedServerAtom } from './atoms'
import { ServerDetailPanel } from './components/server-detail-panel'
import { ServerEditSheet } from './components/server-edit-sheet'
import { ServersMasterList } from './components/servers-master-list'
import type { ServerTab } from './constants'
import { useConfigMode, useFilteredServers, useRuntimeStatus, useServer, useServers, useToolsByServer } from './hooks'

interface ServersPageProps {
  initialTab?: ServerTab
  initialServer?: string
}

export function ServersPage({ initialTab = 'overview', initialServer }: ServersPageProps) {
  const navigate = useNavigate()
  const { data: servers, isLoading, mutate } = useServers()
  const { data: configMode } = useConfigMode()
  const { data: selectedServer, isLoading: isServerLoading } = useServer(initialServer || null)
  const { data: runtimeStatus } = useRuntimeStatus()
  const { serverMap } = useToolsByServer()
  const setSelectedServer = useSetAtom(selectedServerAtom)

  const [editSheetOpen, setEditSheetOpen] = useState(false)
  const [editingServer, setEditingServer] = useState<ServerDetail | null>(null)
  const [isRefreshing, setIsRefreshing] = useState(false)
  const [searchQuery, setSearchQuery] = useState('')
  const [currentTab, setCurrentTab] = useState<ServerTab>(initialTab)

  const { size: masterWidth, resizeHandleProps, isDragging } = useResizable({
    defaultSize: 400,
    minSize: 280,
    maxSize: 600,
    storageKey: 'servers-master-width',
    direction: 'horizontal',
    handle: 'right',
  })

  const isWritable = configMode?.isWritable ?? false
  const serverCount = servers?.length ?? 0

  const filteredServers = useFilteredServers(servers ?? [], searchQuery)

  // Build runtime status map
  const runtimeStatusMap = useMemo(() => {
    const map = new Map<string, ServerRuntimeStatus>()
    if (runtimeStatus) {
      runtimeStatus.forEach((status) => {
        map.set(status.specKey, status)
      })
    }
    return map
  }, [runtimeStatus])

  // Build tool count map
  const toolCountMap = useMemo(() => {
    const map = new Map<string, number>()
    serverMap.forEach((group, specKey) => {
      map.set(specKey, group.tools?.length ?? 0)
    })
    return map
  }, [serverMap])

  const currentToolCount = selectedServer
    ? toolCountMap.get(selectedServer.specKey) ?? 0
    : 0

  // Update selectedServerAtom when selectedServer changes
  useEffect(() => {
    setSelectedServer(selectedServer || null)
  }, [selectedServer, setSelectedServer])

  // Sync tab with URL
  useEffect(() => {
    setCurrentTab(initialTab)
  }, [initialTab])

  // Sync tab with URL
  useEffect(() => {
    setCurrentTab(initialTab)
  }, [initialTab])

  const handleAddServer = () => {
    setEditingServer(null)
    setEditSheetOpen(true)
  }

  const handleEditServer = () => {
    if (selectedServer) {
      setEditingServer(selectedServer as ServerDetail)
      setEditSheetOpen(true)
    }
  }

  const handleRefresh = async () => {
    setIsRefreshing(true)
    try {
      await mutate()
    }
    finally {
      setIsRefreshing(false)
    }
  }

  const handleSelectServer = (name: string) => {
    navigate({
      to: '/servers',
      search: { tab: currentTab, server: name },
    })
  }

  const handleSelectServerTab = (name: string, tab: ServerTab) => {
    navigate({
      to: '/servers',
      search: { tab, server: name },
    })
  }

  const handleTabChange = (tab: ServerTab) => {
    setCurrentTab(tab)
    navigate({
      to: '/servers',
      search: { tab, server: initialServer },
    })
  }

  const handleDeleted = () => {
    navigate({
      to: '/servers',
      search: { tab: currentTab, server: undefined },
    })
  }

  return (
    <div className="flex flex-col h-full">
      {/* Header */}
      <div className="px-6 pt-6 pb-4">
        <m.div
          className="flex items-center justify-between"
          initial={{ opacity: 0, y: -8 }}
          animate={{ opacity: 1, y: 0 }}
          transition={Spring.presets.smooth}
        >
          <div className="space-y-1">
            <div className="flex items-center gap-2">
              <h1 className="text-xl font-semibold tracking-tight">Servers</h1>
              {serverCount > 0 && (
                <Badge variant="secondary" size="sm">
                  {serverCount}
                </Badge>
              )}
            </div>
            <p className="text-sm text-muted-foreground">
              Manage MCP servers, tools, and configurations
            </p>
          </div>
          <div className="flex items-center gap-1">
            <Button
              variant="default"
              size="sm"
              onClick={handleAddServer}
              disabled={!isWritable}
            >
              <PlusIcon className="size-4" />
              Add Server
            </Button>
            <ImportMcpServersSheet />
            <RefreshButton
              onClick={handleRefresh}
              isLoading={isRefreshing}
              tooltip="Refresh servers"
            />
          </div>
        </m.div>
      </div>

      <Separator />

      {/* Toolbar */}
      <div className="px-6 py-4 flex items-center gap-3">
        <div className="relative flex-1 max-w-md">
          <SearchIcon className="absolute left-3 top-1/2 -translate-y-1/2 size-4 text-muted-foreground pointer-events-none" />
          <Input
            type="search"
            placeholder="Search servers..."
            value={searchQuery}
            onChange={e => setSearchQuery(e.target.value)}
            className="pl-9"
          />
        </div>
        {searchQuery && (
          <Badge variant="secondary" size="sm">
            {filteredServers?.length ?? 0} results
          </Badge>
        )}
      </div>

      <Separator />

      {/* Master-Detail Layout */}
      <div className={cn('flex flex-1 min-h-0', isDragging && 'select-none')}>
        {/* Master Panel (Left) */}
        <div className="relative" style={{ width: `${masterWidth}px` }}>
          <ServersMasterList
            servers={filteredServers ?? []}
            selectedServer={initialServer || null}
            onSelectServer={handleSelectServer}
            onSelectServerTab={handleSelectServerTab}
            isLoading={isLoading}
            runtimeStatusMap={runtimeStatusMap}
            toolCountMap={toolCountMap}
          />
          <div
            {...resizeHandleProps}
            className={cn(
              'border-primary/10 hover:border-primary/20 active:border-primary/30 border h-full absolute top-0 right-0 cursor-col-resize z-10',
              isDragging && 'bg-primary/50',
            )}
          />
        </div>

        {/* Detail Panel (Right) */}
        <div className="flex-1 min-w-0">
          <ServerDetailPanel
            server={selectedServer || null}
            isLoading={isServerLoading}
            tab={currentTab}
            toolCount={currentToolCount}
            onTabChange={handleTabChange}
            onEdit={handleEditServer}
            onDeleted={handleDeleted}
          />
        </div>
      </div>

      {/* Edit Sheet */}
      <ServerEditSheet
        open={editSheetOpen}
        onOpenChange={setEditSheetOpen}
        server={editingServer}
        onSaved={() => mutate()}
      />
    </div>
  )
}
