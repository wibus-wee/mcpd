// Input: ServerRuntimeStatus from bindings, useRuntimeStatus hook
// Output: ServerRuntimeStatus component with color-coded instance state indicators
// Position: Runtime status display component for server instances

import type { ServerInitStatus, ServerRuntimeStatus } from '@bindings/mcpd/internal/ui'

import { Badge } from '@/components/ui/badge'
import { cn } from '@/lib/utils'

import { useRuntimeStatus, useServerInitStatus } from '../hooks'

const stateColors: Record<string, string> = {
  ready: 'bg-success',
  busy: 'bg-warning',
  starting: 'bg-info',
  draining: 'bg-muted-foreground',
  stopped: 'bg-muted-foreground/50',
  failed: 'bg-destructive',
}

const stateLabels: Record<string, string> = {
  ready: 'Ready',
  busy: 'Busy',
  starting: 'Starting',
  draining: 'Draining',
  stopped: 'Stopped',
  failed: 'Failed',
}

function StateDot({ state }: { state: string }) {
  const colorClass = stateColors[state] || 'bg-muted-foreground'
  return (
    <span
      className={cn('size-2 rounded-full shrink-0', colorClass)}
      title={stateLabels[state] || state}
    />
  )
}

interface ServerRuntimeIndicatorProps {
	specKey: string
	className?: string
}

export function ServerRuntimeIndicator({
	specKey,
	className,
}: ServerRuntimeIndicatorProps) {
	const { data: runtimeStatus } = useRuntimeStatus()
	const { data: initStatus } = useServerInitStatus()
	const serverStatus = (runtimeStatus as ServerRuntimeStatus[] | undefined)?.find(
		status => status.specKey === specKey,
	)
	const init = (initStatus as ServerInitStatus[] | undefined)?.find(
		status => status.specKey === specKey,
	)

	const hasInstances = serverStatus && serverStatus.instances.length > 0
	if (!init && !hasInstances) {
		return null
	}

	return (
		<div className={cn('flex items-center gap-2', className)}>
			{init && <InitBadge status={init} />}
			{hasInstances && (
				<div className="flex items-center gap-1">
					{serverStatus.instances.slice(0, 5).map(inst => (
						<StateDot key={inst.id} state={inst.state} />
					))}
					{serverStatus.instances.length > 5 && (
						<span className="text-xs text-muted-foreground">
							+{serverStatus.instances.length - 5}
						</span>
					)}
				</div>
			)}
		</div>
	)
}

interface ServerRuntimeDetailsProps {
	status: ServerRuntimeStatus
	className?: string
	initStatus?: ServerInitStatus
}

export function ServerRuntimeDetails({
	status,
	className,
	initStatus,
}: ServerRuntimeDetailsProps) {
	const { instances, stats } = status
	const initLine = initStatus ? <InitStatusLine status={initStatus} /> : null

	if (instances.length === 0) {
		return (
			<div className={cn('space-y-2 text-xs text-muted-foreground', className)}>
				{initLine}
				<div>No active instances</div>
			</div>
		)
	}

	return (
		<div className={cn('space-y-2', className)}>
			{initLine}
			<div className="flex items-center gap-3 text-xs">
				<span className="text-muted-foreground">Instances:</span>
				<div className="flex items-center gap-2">
					{stats.ready > 0 && (
						<span className="flex items-center gap-1">
              <StateDot state="ready" />
              <span>{stats.ready}</span>
            </span>
          )}
          {stats.busy > 0 && (
            <span className="flex items-center gap-1">
              <StateDot state="busy" />
              <span>{stats.busy}</span>
            </span>
          )}
          {stats.starting > 0 && (
            <span className="flex items-center gap-1">
              <StateDot state="starting" />
              <span>{stats.starting}</span>
            </span>
          )}
          {stats.draining > 0 && (
            <span className="flex items-center gap-1">
              <StateDot state="draining" />
              <span>{stats.draining}</span>
            </span>
          )}
          {stats.failed > 0 && (
            <span className="flex items-center gap-1">
              <StateDot state="failed" />
              <span>{stats.failed}</span>
            </span>
          )}
        </div>
      </div>
    </div>
  )
}

export function RuntimeStatusLegend({ className }: { className?: string }) {
	return (
		<div className={cn('flex items-center gap-3 text-xs', className)}>
			<span className="flex items-center gap-1">
				<StateDot state="ready" />
        <span className="text-muted-foreground">Ready</span>
      </span>
      <span className="flex items-center gap-1">
        <StateDot state="busy" />
        <span className="text-muted-foreground">Busy</span>
      </span>
      <span className="flex items-center gap-1">
        <StateDot state="starting" />
        <span className="text-muted-foreground">Starting</span>
      </span>
      <span className="flex items-center gap-1">
        <StateDot state="draining" />
        <span className="text-muted-foreground">Draining</span>
      </span>
      <span className="flex items-center gap-1">
        <StateDot state="failed" />
        <span className="text-muted-foreground">Failed</span>
      </span>
		</div>
	)
}

function InitStatusLine({ status }: { status: ServerInitStatus }) {
	return (
		<div className="flex items-center gap-2 text-xs">
			<InitBadge status={status} />
			{status.lastError && (
				<span
					className="text-destructive truncate max-w-xs"
					title={status.lastError}
				>
					{status.lastError}
				</span>
			)}
		</div>
	)
}

function InitBadge({ status }: { status: ServerInitStatus }) {
	const variant = initVariant[status.state] || 'secondary'
	return (
		<Badge
			variant={variant}
			size="sm"
			className="font-mono"
			title={status.lastError || undefined}
		>
			{formatInitLabel(status)}
		</Badge>
	)
}

const initVariant: Record<string, 'secondary' | 'info' | 'success' | 'warning' | 'destructive'> = {
	pending: 'secondary',
	starting: 'info',
	ready: 'success',
	degraded: 'warning',
	failed: 'destructive',
}

function formatInitLabel(status: ServerInitStatus) {
	const counts = `${status.ready}/${status.minReady}`
	switch (status.state) {
	case 'ready':
		return `Ready (${counts})`
	case 'degraded':
		return `Degraded (${counts})`
	case 'failed':
		return 'Failed'
	case 'starting':
		return `Starting (${counts})`
	default:
		return `Pending (${counts})`
	}
}
