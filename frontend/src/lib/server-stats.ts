// Shared server statistics aggregation utilities
import type { ServerInitStatus, ServerRuntimeStatus } from '@bindings/mcpd/internal/ui'
import { getElapsedMs } from './time'

export interface PoolStats {
  total: number
  ready: number
  busy: number
  starting: number
  failed: number
  draining: number
}

export interface MetricsSummary {
  totalCalls: number
  totalErrors: number
  avgResponseMs: number | null
  lastCallAgeMs: number | null
  startCount: number
}

export interface AggregatedStats {
  totalServers: number
  totalInstances: number
  readyInstances: number
  busyInstances: number
  startingInstances: number
  failedInstances: number
  drainingInstances: number
  suspendedServers: number
  totalCalls: number
  totalErrors: number
  avgDurationMs: number
  errorRate: number
  utilization: number
}

/**
 * Aggregates pool statistics from a single server runtime status
 */
export function getPoolStats(runtimeStatus: ServerRuntimeStatus): PoolStats {
  const { stats } = runtimeStatus
  return {
    total:
      stats.ready
      + stats.busy
      + stats.starting
      + stats.initializing
      + stats.handshaking
      + stats.draining
      + stats.failed,
    ready: stats.ready,
    busy: stats.busy,
    starting: stats.starting + stats.initializing + stats.handshaking,
    failed: stats.failed,
    draining: stats.draining,
  }
}

/**
 * Aggregates metrics summary from a single server runtime status
 */
export function getMetricsSummary(runtimeStatus: ServerRuntimeStatus): MetricsSummary {
  const { metrics } = runtimeStatus
  const avgResponseMs = metrics.totalCalls > 0 ? metrics.totalDurationMs / metrics.totalCalls : null
  const lastCallAgeMs = getElapsedMs(metrics.lastCallAt)

  return {
    totalCalls: metrics.totalCalls,
    totalErrors: metrics.totalErrors,
    avgResponseMs,
    lastCallAgeMs,
    startCount: metrics.startCount,
  }
}

/**
 * Aggregates statistics from multiple server runtime statuses
 */
export function aggregateStats(
  statuses: ServerRuntimeStatus[],
  initStatuses?: ServerInitStatus[],
): AggregatedStats {
  const result: AggregatedStats = {
    totalServers: statuses.length,
    totalInstances: 0,
    readyInstances: 0,
    busyInstances: 0,
    startingInstances: 0,
    failedInstances: 0,
    drainingInstances: 0,
    suspendedServers: 0,
    totalCalls: 0,
    totalErrors: 0,
    avgDurationMs: 0,
    errorRate: 0,
    utilization: 0,
  }

  if (initStatuses) {
    result.suspendedServers = initStatuses.filter(s => s.state === 'suspended').length
  }

  for (const status of statuses) {
    const { stats } = status
    result.totalInstances += stats.total
    result.readyInstances += stats.ready
    result.busyInstances += stats.busy
    result.startingInstances += stats.starting + stats.initializing + stats.handshaking
    result.failedInstances += stats.failed
    result.drainingInstances += stats.draining

    const { metrics } = status
    result.totalCalls += metrics.totalCalls
    result.totalErrors += metrics.totalErrors
    result.avgDurationMs += metrics.totalDurationMs
  }

  if (result.totalCalls > 0) {
    result.avgDurationMs /= result.totalCalls
    result.errorRate = (result.totalErrors / result.totalCalls) * 100
  }

  if (result.totalInstances > 0) {
    result.utilization = ((result.readyInstances + result.busyInstances) / result.totalInstances) * 100
  }

  return result
}