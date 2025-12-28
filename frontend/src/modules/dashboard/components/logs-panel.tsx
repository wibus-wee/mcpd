// Input: Card, Badge, Button, ScrollArea, Checkbox, Select components, logs hook
// Output: LogsPanel component displaying real-time logs
// Position: Dashboard logs section with filtering

import { useSetAtom } from 'jotai'
import {
  AlertCircleIcon,
  AlertTriangleIcon,
  BugIcon,
  InfoIcon,
  RefreshCwIcon,
  ScrollTextIcon,
  TrashIcon,
} from 'lucide-react'
import { useEffect, useRef, useState } from 'react'

import { logStreamTokenAtom } from '@/atoms/logs'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Checkbox } from '@/components/ui/checkbox'
import { Label } from '@/components/ui/label'
import { ScrollArea } from '@/components/ui/scroll-area'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Separator } from '@/components/ui/separator'
import { useCoreState } from '@/hooks/use-core-state'
import type { LogEntry, LogSource } from '@/hooks/use-logs'
import { useLogs } from '@/hooks/use-logs'
import { cn } from '@/lib/utils'

const levelConfig = {
  debug: {
    icon: BugIcon,
    color: 'text-muted-foreground',
    badge: 'secondary' as const,
  },
  info: {
    icon: InfoIcon,
    color: 'text-info',
    badge: 'info' as const,
  },
  warn: {
    icon: AlertTriangleIcon,
    color: 'text-warning',
    badge: 'warning' as const,
  },
  error: {
    icon: AlertCircleIcon,
    color: 'text-destructive',
    badge: 'error' as const,
  },
}

const sourceConfig: Record<LogSource, { label: string, badge: 'secondary' | 'info' | 'success' | 'outline' }> = {
  core: {
    label: 'Core',
    badge: 'secondary',
  },
  downstream: {
    label: 'Downstream',
    badge: 'info',
  },
  ui: {
    label: 'Wails UI',
    badge: 'success',
  },
  unknown: {
    label: 'Unknown',
    badge: 'outline',
  },
}

const hiddenFieldKeys = new Set(['log_source', 'logger', 'serverType', 'stream', 'timestamp'])

const formatFieldValue = (value: unknown) => {
  if (value === null) return 'null'
  if (value === undefined) return 'undefined'
  if (value instanceof Date) return value.toISOString()
  if (typeof value === 'object') {
    try {
      return JSON.stringify(value)
    }
    catch {
      return String(value)
    }
  }
  return String(value)
}

function LogItem({ log }: { log: LogEntry }) {
  const config = levelConfig[log.level] ?? levelConfig.info
  const sourceMeta = sourceConfig[log.source] ?? sourceConfig.unknown
  const Icon = config.icon
  const detailEntries = Object.entries(log.fields ?? {}).filter(
    ([key]) => !hiddenFieldKeys.has(key),
  )

  return (
    <div className="flex items-start gap-3 py-2 px-3 hover:bg-muted/50 transition-colors">
      <Icon className={cn('size-4 mt-0.5 shrink-0', config.color)} />
      <div className="flex-1 min-w-0 space-y-1">
        <div className="flex flex-wrap items-center gap-2">
          <Badge variant={config.badge} size="sm">
            {log.level}
          </Badge>
          <Badge variant={sourceMeta.badge} size="sm">
            {sourceMeta.label}
          </Badge>
          {log.serverType && (
            <Badge variant="outline" size="sm">
              {log.serverType}
            </Badge>
          )}
          {log.stream && (
            <Badge variant="outline" size="sm">
              {log.stream}
            </Badge>
          )}
          {log.logger && (
            <span className="text-muted-foreground text-xs font-mono">
              @{log.logger}
            </span>
          )}
          <span className="text-muted-foreground text-xs ml-auto">
            {log.timestamp.toLocaleTimeString()}
          </span>
        </div>
        <p className="text-sm break-words whitespace-pre-wrap">{log.message}</p>
        {detailEntries.length > 0 && (
          <div className="mt-2 grid grid-cols-1 gap-2 rounded-md bg-muted/50 p-2 text-xs text-muted-foreground md:grid-cols-2">
            {detailEntries.map(([key, value]) => (
              <div key={key} className="flex gap-2">
                <span className="min-w-[120px] text-foreground font-medium">{key}</span>
                <span className="break-words font-mono">{formatFieldValue(value)}</span>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}

export function LogsPanel() {
  const { logs, mutate } = useLogs()
  const { coreStatus } = useCoreState()
  const [levelFilter, setLevelFilter] = useState<string>('all')
  const [sourceFilter, setSourceFilter] = useState<LogSource | 'all'>('all')
  const [serverFilter, setServerFilter] = useState<string>('all')
  const [autoScroll, setAutoScroll] = useState(true)
  const bumpLogStreamToken = useSetAtom(logStreamTokenAtom)
  const topAnchorRef = useRef<HTMLDivElement | null>(null)
  const levelLabels: Record<string, string> = {
    all: 'All levels',
    debug: 'Debug',
    info: 'Info',
    warn: 'Warning',
    error: 'Error',
  }

  const sourceLabels: Record<string, string> = {
    all: 'All sources',
    core: 'Core',
    downstream: 'Downstream',
    ui: 'Wails UI',
    unknown: 'Unknown',
  }

  const serverOptions = Array.from(
    new Set(
      logs
        .map(log => log.serverType)
        .filter((server): server is string => typeof server === 'string'),
    ),
  ).sort()

  const filteredLogs = logs.filter((log) => {
    if (levelFilter !== 'all' && log.level !== levelFilter) {
      return false
    }
    if (sourceFilter !== 'all' && log.source !== sourceFilter) {
      return false
    }
    if (serverFilter !== 'all' && log.serverType !== serverFilter) {
      return false
    }
    return true
  })

  const clearLogs = () => {
    mutate([], { revalidate: false })
  }

  const handleLevelChange = (value: string | null) => {
    setLevelFilter(value ?? 'all')
  }

  const forceRefresh = () => {
    bumpLogStreamToken(value => value + 1)
  }

  const isConnected = coreStatus === 'running' && logs.length > 0
  const isDisconnected = coreStatus === 'stopped' || coreStatus === 'error'
  const isWaiting = coreStatus === 'running' && logs.length === 0
  const showServerFilter = serverOptions.length > 0
    && (sourceFilter === 'all' || sourceFilter === 'downstream')

  useEffect(() => {
    if (!autoScroll || filteredLogs.length === 0) {
      return
    }
    topAnchorRef.current?.scrollIntoView({ behavior: 'smooth', block: 'start' })
  }, [autoScroll, filteredLogs.length])

  useEffect(() => {
    if (sourceFilter !== 'downstream' && sourceFilter !== 'all' && serverFilter !== 'all') {
      setServerFilter('all')
    }
  }, [sourceFilter, serverFilter])

  const renderEmptyState = () => {
    if (isDisconnected) {
      return (
        <>
          <p className="text-sm font-medium">Core is not running</p>
          <p className="text-xs">Start the core to see logs</p>
        </>
      )
    }
    if (isWaiting) {
      return (
        <>
          <p className="text-sm font-medium">Waiting for logs...</p>
          <Button
            variant="ghost"
            size="sm"
            onClick={forceRefresh}
            className="mt-2"
          >
            <RefreshCwIcon className="size-3 mr-1" />
            Restart Log Stream
          </Button>
        </>
      )
    }
    return (
      <>
        <p className="text-sm">No logs yet</p>
        <p className="text-xs">Logs will appear here when the core is running</p>
      </>
    )
  }

  return (
    <div className="h-full">
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle className="flex items-center gap-2">
              <ScrollTextIcon className="size-5" />
              Logs
              <Badge variant="secondary" size="sm">
                {logs.length}
              </Badge>
              {isConnected && (
                <Badge variant="success" size="sm" className="ml-1">
                  Connected
                </Badge>
              )}
              {isDisconnected && (
                <Badge variant="error" size="sm" className="ml-1">
                  Disconnected
                </Badge>
              )}
              {isWaiting && (
                <Badge variant="warning" size="sm" className="ml-1">
                  Waiting...
                </Badge>
              )}
            </CardTitle>
            <div className="flex flex-wrap items-center gap-3">
              <div className="flex items-center gap-2">
                <Checkbox
                  id="auto-scroll"
                  checked={autoScroll}
                  onCheckedChange={checked => setAutoScroll(checked === true)}
                />
                <Label htmlFor="auto-scroll" className="text-sm">
                  Auto-scroll
                </Label>
              </div>
              <Select value={levelFilter} onValueChange={handleLevelChange}>
                <SelectTrigger size="sm" className="w-32">
                  <SelectValue>
                    {value =>
                      value
                        ? levelLabels[String(value)] ?? String(value)
                        : 'Filter level'}
                  </SelectValue>
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All levels</SelectItem>
                  <SelectItem value="debug">Debug</SelectItem>
                  <SelectItem value="info">Info</SelectItem>
                  <SelectItem value="warn">Warning</SelectItem>
                  <SelectItem value="error">Error</SelectItem>
                </SelectContent>
              </Select>
              <Select
                value={sourceFilter}
                onValueChange={value => setSourceFilter(value as LogSource | 'all')}
              >
                <SelectTrigger size="sm" className="w-36">
                  <SelectValue>
                    {value =>
                      value
                        ? sourceLabels[String(value)] ?? String(value)
                        : 'Filter source'}
                  </SelectValue>
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All sources</SelectItem>
                  <SelectItem value="core">Core</SelectItem>
                  <SelectItem value="downstream">Downstream</SelectItem>
                  <SelectItem value="ui">Wails UI</SelectItem>
                  <SelectItem value="unknown">Unknown</SelectItem>
                </SelectContent>
              </Select>
              {showServerFilter && (
                <Select value={serverFilter} onValueChange={setServerFilter}>
                  <SelectTrigger size="sm" className="w-40">
                    <SelectValue>
                      {value => (value ? String(value) : 'Filter server')}
                    </SelectValue>
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">All servers</SelectItem>
                    {serverOptions.map(server => (
                      <SelectItem key={server} value={server}>
                        {server}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              )}
              {isWaiting && (
                <Button
                  variant="ghost"
                  size="icon-sm"
                  onClick={forceRefresh}
                  title="Restart log stream"
                >
                  <RefreshCwIcon className="size-4" />
                </Button>
              )}
              <Button
                variant="ghost"
                size="icon-sm"
                onClick={clearLogs}
              >
                <TrashIcon className="size-4" />
              </Button>
            </div>
          </div>
        </CardHeader>
        <Separator />
        <CardContent className="p-0">
          <ScrollArea className="h-80">
            {filteredLogs.length === 0 ? (
              <div className="flex flex-col items-center justify-center h-full py-12 text-muted-foreground">
                <ScrollTextIcon className="size-8 mb-2 opacity-50" />
                {renderEmptyState()}
              </div>
            ) : (
              <div>
                <div ref={topAnchorRef} className="h-0" />
                <div className="divide-y">
                  {filteredLogs.map(log => (
                    <LogItem key={log.id} log={log} />
                  ))}
                </div>
              </div>
            )}
          </ScrollArea>
        </CardContent>
      </Card>
    </div>
  )
}
