import {
  AlertTriangleIcon,
  ServerIcon,
} from 'lucide-react'
import { useMemo } from 'react'

import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from '@/components/ui/empty'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Skeleton } from '@/components/ui/skeleton'
import { useServer } from '@/modules/servers/hooks'
import { buildCommandSummary, DetailRow } from '@/modules/shared/server-detail'

interface ServerConfigPanelProps {
  serverName: string | null
  onDeleted?: () => void
  onEdit?: () => void
}

// type ErrorState = {
//   title: string
//   description: string
// }

function DetailSkeleton() {
  return (
    <div className="space-y-4 p-6">
      <Skeleton className="h-6 w-48" />
      <Skeleton className="h-4 w-64" />
      <Skeleton className="h-36 w-full" />
      <Skeleton className="h-36 w-full" />
    </div>
  )
}

function formatFieldName(fieldName: string): string {
  return fieldName
    .replaceAll(/([A-Z])/g, ' $1')
    .replace(/^./, str => str.toUpperCase())
    .trim()
}

function formatFieldValue(value: any): string {
  if (value === null || value === undefined) return '--'
  if (typeof value === 'boolean') return value ? 'Yes' : 'No'
  if (typeof value === 'string' && value === '') return '--'
  if (Array.isArray(value)) {
    return value.length > 0 ? value.join(', ') : '--'
  }
  if (typeof value === 'object') {
    if (Object.keys(value).length === 0) return '--'
    return JSON.stringify(value, null, 2)
  }
  return String(value)
}

export function ServerConfigPanel({ serverName }: ServerConfigPanelProps) {
  const { data: server, isLoading } = useServer(serverName)

  const commandSummary = server ? buildCommandSummary(server) : '--'
  const envCount = server ? Object.keys(server.env ?? {}).length : 0

  // 生成完整的规格字段列表
  const specFields = useMemo(() => {
    if (!server) return []

    const excludeFields = new Set([
      'disabled',
      'specKey',
    ])

    return Object.entries(server)
      .filter(([key]) => !excludeFields.has(key))
      .map(([key, value]) => ({
        key,
        label: formatFieldName(key),
        value: formatFieldValue(value),
        isMono: typeof value === 'object' || Array.isArray(value) || key === 'http',
      }))
      .sort((a, b) => a.label.localeCompare(b.label))
  }, [server])

  if (!serverName) {
    return (
      <Empty className="py-16">
        <EmptyHeader>
          <EmptyMedia variant="icon">
            <ServerIcon className="size-4" />
          </EmptyMedia>
          <EmptyTitle className="text-sm">Select a server</EmptyTitle>
          <EmptyDescription className="text-xs">
            Choose a server from the list to view its configuration.
          </EmptyDescription>
        </EmptyHeader>
      </Empty>
    )
  }

  if (isLoading) {
    return <DetailSkeleton />
  }

  if (!server) {
    return (
      <Empty className="py-16">
        <EmptyHeader>
          <EmptyMedia variant="icon">
            <AlertTriangleIcon className="size-4" />
          </EmptyMedia>
          <EmptyTitle className="text-sm">Server not found</EmptyTitle>
          <EmptyDescription className="text-xs">
            The selected server could not be loaded.
          </EmptyDescription>
        </EmptyHeader>
      </Empty>
    )
  }

  return (
    <ScrollArea className="h-full">
      <div
        className="space-y-4"
      >
        <DetailRow label="Command" value={commandSummary} mono />
        <DetailRow label="Working directory" value={server.cwd || '--'} mono />
        <DetailRow label="Environment" value={`${envCount} variables`} />
        {specFields.map(field => (
          <DetailRow
            key={field.key}
            label={field.label}
            value={field.value}
            mono={field.isMono}
          />
        ))}
      </div>
    </ScrollArea>
  )
}
