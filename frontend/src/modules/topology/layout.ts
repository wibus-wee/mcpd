// Input: Profile/caller/server data from hooks, binding types
// Output: buildTopology function and layoutConfig for graph positioning
// Position: Layout computation layer for topology visualization

import type {
  ActiveCaller,
  ProfileDetail,
  ProfileSummary,
  ServerRuntimeStatus,
} from '@bindings/mcpd/internal/ui'
import { MarkerType, type Edge } from '@xyflow/react'

import type { FlowNode, LayoutResult } from './types'

export const layoutConfig = {
  columns: {
    caller: 0,
    profile: 280,
    server: 560,
    instance: 800,
  },
  nodeGap: 96,
  serverGap: 84,
  instanceGap: 60,
  clusterGap: 140,
  minClusterHeight: 120,
}

type AggregatedServer = {
  key: string
  name: string
  protocolVersions: Set<string>
  strategies: Set<string>
  sessionTTLSeconds: number
  sessionTTLMixed: boolean
  maxConcurrent: number
  exposeToolsCount: number
  profileNames: Set<string>
}

type ServerTagInput = {
  strategy: string
  strategyMixed: boolean
  sessionTTLSeconds: number
  sessionTTLMixed: boolean
  maxConcurrent: number
  exposeToolsCount: number
  profileCount: number
}

const buildServerTags = ({
  strategy,
  strategyMixed,
  sessionTTLSeconds,
  sessionTTLMixed,
  maxConcurrent,
  exposeToolsCount,
  profileCount,
}: ServerTagInput) => {
  const tags: string[] = []

  const strategyLabel = {
    stateless: 'Stateless',
    stateful: 'Stateful',
    persistent: 'Persistent',
    singleton: 'Singleton',
  }[strategy] ?? strategy

  if (strategyMixed) {
    tags.push('Strategy Mixed')
  } else if (strategy !== 'stateless') {
    tags.push(strategyLabel)
  }

  if (strategy === 'stateful') {
    if (sessionTTLMixed) {
      tags.push('Session TTL Mixed')
    } else if (sessionTTLSeconds > 0) {
      tags.push(`Session TTL ${sessionTTLSeconds}s`)
    } else {
      tags.push('Session TTL Off')
    }
  }

  if (maxConcurrent > 0) {
    tags.push(`Max ${maxConcurrent}`)
  }

  if (exposeToolsCount > 0) {
    tags.push(`Tools ${exposeToolsCount}`)
  }

  if (profileCount > 1) {
    tags.push(`Profiles ${profileCount}`)
  }

  return tags
}

export const buildTopology = (
  profiles: ProfileSummary[],
  profileDetails: ProfileDetail[],
  callers: Record<string, string>,
  activeCallers: ActiveCaller[],
  runtimeStatus: ServerRuntimeStatus[],
): LayoutResult => {
  const detailsByName = new Map(
    profileDetails.map(profile => [profile.name, profile]),
  )
  const profileNameSet = new Set(profiles.map(profile => profile.name))
  const activeCallerSet = new Set(activeCallers.map(caller => caller.caller))
  const activeCallerMap = new Map(
    activeCallers.map(caller => [caller.caller, caller.pid]),
  )
  const callersByProfile = new Map<string, string[]>()
  const serversByKey = new Map<string, AggregatedServer>()
  const runtimeStatusBySpecKey = new Map(
    runtimeStatus.map(status => [status.specKey, status]),
  )

  for (const [caller, profileName] of Object.entries(callers)) {
    const bucket = callersByProfile.get(profileName) ?? []
    bucket.push(caller)
    callersByProfile.set(profileName, bucket)
  }

  const missingProfiles = Array.from(callersByProfile.keys()).filter(
    profileName => !profileNameSet.has(profileName),
  )

  const orderedProfiles = [
    ...profiles.map(profile => ({
      name: profile.name,
      isDefault: profile.isDefault,
      isMissing: false,
    })),
    ...missingProfiles.sort().map(profileName => ({
      name: profileName,
      isDefault: false,
      isMissing: true,
    })),
  ]

  const nodes: FlowNode[] = []
  const edges: Edge[] = []
  const profilePositions = new Map<string, number>()
  const serverPositions = new Map<string, number>()

  let cursorY = 0

  for (const profile of orderedProfiles) {
    const profileDetail = detailsByName.get(profile.name)
    const servers = profileDetail?.servers ?? []
    const callerList = (callersByProfile.get(profile.name) ?? []).slice().sort()
    const clusterSize = Math.max(callerList.length, 1)
    const clusterHeight = (clusterSize - 1) * layoutConfig.nodeGap
    const baseline = Math.max(clusterHeight, layoutConfig.minClusterHeight)
    const profileY = cursorY + baseline / 2
    profilePositions.set(profile.name, profileY)

    nodes.push({
      id: `profile:${profile.name}`,
      type: 'profile',
      position: {
        x: layoutConfig.columns.profile,
        y: profileY,
      },
      data: {
        name: profile.name,
        serverCount: servers.length,
        isDefault: profile.isDefault,
        isMissing: profile.isMissing,
      },
    })

    for (const server of servers) {
      const serverKey = server.specKey || server.name
      const protocolLabel = server.protocolVersion || 'default'
      const existing = serversByKey.get(serverKey)

      if (existing) {
        existing.strategies.add(server.strategy)
        if (existing.sessionTTLSeconds !== server.sessionTTLSeconds) {
          existing.sessionTTLMixed = true
        }
        existing.maxConcurrent = Math.max(
          existing.maxConcurrent,
          server.maxConcurrent,
        )
        existing.exposeToolsCount = Math.max(
          existing.exposeToolsCount,
          server.exposeTools.length,
        )
        existing.profileNames.add(profile.name)
        existing.protocolVersions.add(protocolLabel)
      } else {
        serversByKey.set(serverKey, {
          key: serverKey,
          name: server.name,
          protocolVersions: new Set([protocolLabel]),
          strategies: new Set([server.strategy]),
          sessionTTLSeconds: server.sessionTTLSeconds,
          sessionTTLMixed: false,
          maxConcurrent: server.maxConcurrent,
          exposeToolsCount: server.exposeTools.length,
          profileNames: new Set([profile.name]),
        })
      }

      edges.push({
        id: `edge:profile:${profile.name}->server:${serverKey}`,
        source: `profile:${profile.name}`,
        target: `server:${serverKey}`,
        type: 'smoothstep',
        markerEnd: {
          type: MarkerType.ArrowClosed,
          color: 'var(--chart-2)',
        },
        style: {
          stroke: 'var(--chart-2)',
          strokeWidth: 1.5,
          strokeOpacity: 0.6,
        },
      })
    }

    const callerStartY =
      profileY - ((callerList.length - 1) * layoutConfig.nodeGap) / 2

    callerList.forEach((caller, index) => {
      const nodeId = `caller:${caller}`
      const isActive = activeCallerSet.has(caller)
      const pid = activeCallerMap.get(caller)

      nodes.push({
        id: nodeId,
        type: 'caller',
        position: {
          x: layoutConfig.columns.caller,
          y: callerStartY + index * layoutConfig.nodeGap,
        },
        data: {
          name: caller,
          profileName: profile.name,
          pid: isActive ? pid : undefined,
        },
      })

      edges.push({
        id: `edge:${nodeId}->profile:${profile.name}`,
        source: nodeId,
        target: `profile:${profile.name}`,
        type: 'smoothstep',
        animated: isActive,
        markerEnd: {
          type: MarkerType.ArrowClosed,
          color: isActive ? 'var(--chart-4)' : 'var(--info)',
        },
        style: {
          stroke: isActive ? 'var(--chart-4)' : 'var(--info)',
          strokeWidth: isActive ? 2 : 1.4,
          strokeOpacity: isActive ? 0.9 : 0.55,
          strokeDasharray: isActive ? '6 4' : '4 4',
        },
      })
    })

    cursorY += baseline + layoutConfig.clusterGap
  }

  const serverEntries = Array.from(serversByKey.values()).map(entry => {
    const profileYs = Array.from(entry.profileNames)
      .map(name => profilePositions.get(name))
      .filter((value): value is number => value !== undefined)
    const desiredY =
      profileYs.length > 0
        ? profileYs.reduce((sum, value) => sum + value, 0) / profileYs.length
        : 0

    return {
      entry,
      desiredY,
    }
  })

  serverEntries.sort((a, b) => a.desiredY - b.desiredY)

  let lastServerY = -Infinity

  for (const { entry, desiredY } of serverEntries) {
    const resolvedY = Math.max(desiredY, lastServerY + layoutConfig.serverGap)
    lastServerY = resolvedY
    const protocolVersion =
      entry.protocolVersions.size === 1
        ? Array.from(entry.protocolVersions)[0]
        : 'mixed'

    nodes.push({
      id: `server:${entry.key}`,
      type: 'server',
      position: {
        x: layoutConfig.columns.server,
        y: resolvedY,
      },
      data: {
        name: entry.name,
        protocolVersion,
        tags: buildServerTags({
          strategy:
            entry.strategies.size === 1
              ? Array.from(entry.strategies)[0]
              : 'mixed',
          strategyMixed: entry.strategies.size > 1,
          sessionTTLSeconds: entry.sessionTTLSeconds,
          sessionTTLMixed: entry.sessionTTLMixed,
          maxConcurrent: entry.maxConcurrent,
          exposeToolsCount: entry.exposeToolsCount,
          profileCount: entry.profileNames.size,
        }),
      },
    })
    serverPositions.set(entry.key, resolvedY)
  }

  let lastInstanceY = -Infinity
  let instanceCount = 0

  for (const [serverKey, serverStatus] of runtimeStatusBySpecKey.entries()) {
    const serverY = serverPositions.get(serverKey)
    if (serverY === undefined) continue

    const instances = serverStatus.instances
    if (instances.length === 0) continue

    const instanceStartY =
      serverY - ((instances.length - 1) * layoutConfig.instanceGap) / 2

    instances.forEach((instance, index) => {
      instanceCount++
      const nodeId = `instance:${serverKey}:${instance.id}`

      nodes.push({
        id: nodeId,
        type: 'instance',
        position: {
          x: layoutConfig.columns.instance,
          y: instanceStartY + index * layoutConfig.instanceGap,
        },
        data: {
          id: instance.id,
          state: instance.state,
          busyCount: instance.busyCount,
        },
      })

      edges.push({
        id: `edge:server:${serverKey}->instance:${instance.id}`,
        source: `server:${serverKey}`,
        target: nodeId,
        type: 'smoothstep',
        markerEnd: {
          type: MarkerType.ArrowClosed,
          color: 'var(--border)',
        },
        style: {
          stroke: 'var(--border)',
          strokeWidth: 1,
          strokeOpacity: 0.5,
        },
      })
    })

    lastInstanceY = Math.max(
      lastInstanceY,
      instanceStartY + (instances.length - 1) * layoutConfig.instanceGap,
    )
  }

  const serverCount = serversByKey.size

  return {
    nodes,
    edges,
    profileCount: orderedProfiles.length,
    serverCount,
    callerCount: Object.keys(callers).length,
    instanceCount,
  }
}
