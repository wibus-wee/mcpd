// Input: ServerRuntimeStatus from bindings, useRuntimeStatus hook
// Output: ServerRuntimeStatus component with color-coded instance state indicators
// Position: Runtime status display component for server instances

import type { ServerInitStatus, ServerRuntimeStatus } from '@bindings/mcpd/internal/ui'
import { WailsService } from '@bindings/mcpd/internal/ui'
import { useState } from 'react'

import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'
import { formatDuration, formatLatency, getElapsedMs } from '@/lib/time'

import { useRuntimeStatus, useServerInitStatus } from '../hooks'

const stateColors: Record<string, string> = {
  ready: 'bg-success',
  busy: 'bg-warning',
  starting: 'bg-info',
  initializing: 'bg-info/80',
  handshaking: 'bg-info/60',
  draining: 'bg-muted-foreground',
  stopped: 'bg-muted-foreground/50',
  failed: 'bg-destructive',
}

const stateLabels: Record<string, string> = {
  ready: 'Ready',
  busy: 'Busy',
  starting: 'Starting',
  initializing: 'Initializing',
  handshaking: 'Handshaking',
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

interface ServerRuntimeSummaryProps {
	specKey: string
	className?: string
}

export function ServerRuntimeSummary({ specKey, className }: ServerRuntimeSummaryProps) {
	const { data: runtimeStatus } = useRuntimeStatus()
	const { data: initStatus } = useServerInitStatus()
	const serverStatus = (runtimeStatus as ServerRuntimeStatus[] | undefined)?.find(
		status => status.specKey === specKey,
	)
	const init = (initStatus as ServerInitStatus[] | undefined)?.find(
		status => status.specKey === specKey,
	)

	if (!serverStatus && !init) {
		return null
	}

	if (!serverStatus && init) {
		return (
			<div className={cn('space-y-2 text-xs text-muted-foreground', className)}>
				<InitStatusLine status={init} />
			</div>
		)
	}

	if (!serverStatus) {
		return null
	}

	return (
		<ServerRuntimeDetails
			status={serverStatus}
			initStatus={init}
			className={className}
		/>
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
	const metrics = status.metrics
	const initLine = initStatus ? <InitStatusLine status={initStatus} /> : null

	if (instances.length === 0) {
		return (
			<div className={cn('space-y-2 text-xs text-muted-foreground', className)}>
				{initLine}
				<div>No active instances</div>
			</div>
		)
	}

	const uptimeMs = getOldestUptimeMs(instances)
	const restartCount = Math.max(0, metrics.startCount - 1)
	const avgResponseMs = metrics.totalCalls > 0
		? metrics.totalDurationMs / metrics.totalCalls
		: null
	const lastCallAgeMs = getElapsedMs(metrics.lastCallAt)

	return (
		<div className={cn('space-y-2', className)}>
			{initLine}
			<div className="flex flex-wrap items-center gap-2 text-xs text-muted-foreground">
				{uptimeMs !== null && (
					<span>Up {formatDuration(uptimeMs)}</span>
				)}
				<span>Restarts {restartCount}</span>
				<span>
					Avg {avgResponseMs === null ? '--' : formatLatency(avgResponseMs)}
				</span>
				{lastCallAgeMs !== null && (
					<span>Last call {formatDuration(lastCallAgeMs)} ago</span>
				)}
			</div>
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
					{stats.initializing > 0 && (
						<span className="flex items-center gap-1">
							<StateDot state="initializing" />
							<span>{stats.initializing}</span>
						</span>
					)}
					{stats.handshaking > 0 && (
						<span className="flex items-center gap-1">
							<StateDot state="handshaking" />
							<span>{stats.handshaking}</span>
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
			<div className="space-y-1 text-xs text-muted-foreground">
				{instances.map((inst) => (
					<div key={inst.id} className="flex flex-wrap items-center gap-2">
						<StateDot state={inst.state} />
						<span
							className="font-mono text-foreground/80"
							title={inst.id}
						>
							{formatInstanceId(inst.id)}
						</span>
						{renderInstanceTimeline(inst)}
					</div>
				))}
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
				<StateDot state="initializing" />
				<span className="text-muted-foreground">Initializing</span>
			</span>
			<span className="flex items-center gap-1">
				<StateDot state="handshaking" />
				<span className="text-muted-foreground">Handshaking</span>
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

function getOldestUptimeMs(instances: ServerRuntimeStatus['instances']): number | null {
	let oldestStartedAt: number | null = null
	for (const inst of instances) {
		const startedAt = getInstanceStartedAt(inst)
		if (startedAt === null) {
			continue
		}
		if (oldestStartedAt === null || startedAt < oldestStartedAt) {
			oldestStartedAt = startedAt
		}
	}
	if (oldestStartedAt === null) {
		return null
	}
	return Math.max(0, Date.now() - oldestStartedAt)
}

function renderInstanceTimeline(inst: ServerRuntimeStatus['instances'][number]) {
	const parts: string[] = []
	const uptimeMs = getElapsedMs(inst.handshakedAt || inst.spawnedAt)
	if (uptimeMs !== null) {
		parts.push(`Up ${formatDuration(uptimeMs)}`)
	}

	const spawnedAt = parseTimestamp(inst.spawnedAt)
	const handshakedAt = parseTimestamp(inst.handshakedAt)
	if (spawnedAt !== null && handshakedAt !== null) {
		const handshakeMs = Math.max(0, handshakedAt - spawnedAt)
		parts.push(`Handshake ${formatLatency(handshakeMs)}`)
	}

	const heartbeatAgeMs = getElapsedMs(inst.lastHeartbeatAt)
	if (heartbeatAgeMs !== null) {
		parts.push(`Heartbeat ${formatDuration(heartbeatAgeMs)} ago`)
	}

	if (parts.length === 0) {
		return null
	}

	return <span>{parts.join(' · ')}</span>
}

function getInstanceStartedAt(inst: ServerRuntimeStatus['instances'][number]) {
	const handshakedAt = parseTimestamp(inst.handshakedAt)
	if (handshakedAt !== null) {
		return handshakedAt
	}
	return parseTimestamp(inst.spawnedAt)
}

function parseTimestamp(value: string) {
	if (!value) {
		return null
	}
	const parsed = Date.parse(value)
	if (Number.isNaN(parsed)) {
		return null
	}
	return parsed
}

function formatInstanceId(id: string) {
	if (id.length <= 12) {
		return id
	}
	return `${id.slice(0, 8)}...${id.slice(-3)}`
}

function InitStatusLine({ status }: { status: ServerInitStatus }) {
	const [isRetrying, setIsRetrying] = useState(false)
	const [retryError, setRetryError] = useState<string | null>(null)
	const retryInfo = formatRetryInfo(status)

	const handleRetry = async () => {
		if (isRetrying) {
			return
		}
		setIsRetrying(true)
		setRetryError(null)
		try {
			await WailsService.RetryServerInit({ specKey: status.specKey })
		} catch (err) {
			setRetryError(err instanceof Error ? err.message : 'Retry failed')
		} finally {
			setIsRetrying(false)
		}
	}

	return (
		<div className="flex flex-wrap items-center gap-2 text-xs">
			<InitBadge status={status} />
			{retryInfo && (
				<span className="text-muted-foreground">{retryInfo}</span>
			)}
			{status.state === 'suspended' && (
				<Button
					variant="outline"
					size="xs"
					onClick={handleRetry}
					disabled={isRetrying}
				>
					{isRetrying ? 'Retrying...' : 'Retry'}
				</Button>
			)}
			{status.lastError && (
				<span
					className="text-destructive truncate max-w-xs"
					title={status.lastError}
				>
					{status.lastError}
				</span>
			)}
			{retryError && (
				<span className="text-destructive">{retryError}</span>
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
	suspended: 'warning',
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
	case 'suspended':
		return 'Suspended'
	default:
		return `Pending (${counts})`
	}
}

function formatRetryInfo(status: ServerInitStatus) {
	const parts: string[] = []
	if (status.retryCount > 0) {
		parts.push(`Retries: ${status.retryCount}`)
	}
	if (status.nextRetryAt) {
		const retryAt = Date.parse(status.nextRetryAt)
		if (!Number.isNaN(retryAt)) {
			const remainingSeconds = Math.max(0, Math.round((retryAt - Date.now()) / 1000))
			parts.push(`Next in ${remainingSeconds}s`)
		} else {
			parts.push('Next retry scheduled')
		}
	}
	if (parts.length === 0) {
		return ''
	}
	return parts.join(' | ')
}
