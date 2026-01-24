// Input: Config hooks for profiles/callers, React Flow, topology submodules
// Output: ConfigFlow component - topology graph of profiles, callers, and servers
// Position: Visualization panel inside config module tabs

import {
  Background,
  BackgroundVariant,
  ReactFlow,
  ReactFlowProvider,
  useReactFlow,
  type Edge,
} from '@xyflow/react'
import { useCallback } from 'react'
import '@xyflow/react/dist/style.css'
import { Share2Icon } from 'lucide-react'

import { Skeleton } from '@/components/ui/skeleton'
import { useActiveCallers } from '@/hooks/use-active-callers'
import { useCallers, useProfileDetails, useProfiles, useRuntimeStatus } from '../config/hooks'

import type {
  CallerFlowNode,
  FlowNode,
  ProfileFlowNode,
  ServerFlowNode,
} from './types'
import { nodeTypes } from './nodes'
import { buildTopology } from './layout'
import { FlowEmpty, FlowSkeleton } from './components'

const ConfigFlowInner = ({
  nodes,
  edges,
}: {
  nodes: FlowNode[]
  edges: Edge[]
}) => {
  const { fitView, getNodes } = useReactFlow()

  const onNodeClick = useCallback(
    (_: React.MouseEvent, node: FlowNode) => {
      const allNodes = getNodes() as FlowNode[]

      let relatedNodes: FlowNode[] = []

      if (node.type === 'caller') {
        const callerNode = node as CallerFlowNode
        const profileName = callerNode.data.profileName
        relatedNodes = [node]
        if (profileName) {
          const profileNode = allNodes.find(
            n => n.type === 'profile' && n.id === `profile:${profileName}`,
          )
          if (profileNode) relatedNodes.push(profileNode)
        }
      } else if (node.type === 'profile') {
        const profileNode = node as ProfileFlowNode
        const callersForProfile = allNodes.filter(
          n =>
            n.type === 'caller' &&
            (n as CallerFlowNode).data.profileName === profileNode.data.name,
        )
        relatedNodes = [node, ...callersForProfile]
      } else if (node.type === 'server') {
        const serverNode = node as ServerFlowNode
        const instancesForServer = allNodes.filter(
          n => n.type === 'instance' && n.id.startsWith(`instance:${serverNode.id.replace('server:', '')}:`),
        )
        relatedNodes = [node, ...instancesForServer]
      } else if (node.type === 'instance') {
        const instanceNodeId = node.id
        const serverKey = instanceNodeId.split(':')[1]
        const serverNode = allNodes.find(
          n => n.type === 'server' && n.id === `server:${serverKey}`,
        )
        relatedNodes = [node]
        if (serverNode) relatedNodes.push(serverNode)
      }

      if (relatedNodes.length > 0) {
        fitView({
          nodes: relatedNodes,
          padding: 0.6,
          duration: 400,
          minZoom: 0.4,
          maxZoom: 1,
        })
      }
    },
    [fitView, getNodes],
  )

  return (
    <ReactFlow
      nodes={nodes}
      edges={edges}
      nodeTypes={nodeTypes}
      className="h-full w-full"
      fitView
      fitViewOptions={{ padding: 0.2 }}
      nodesDraggable={false}
      nodesConnectable={false}
      zoomOnScroll
      panOnScroll
      minZoom={0.4}
      maxZoom={1.2}
      onNodeClick={onNodeClick}
    >
      <Background
        variant={BackgroundVariant.Dots}
        gap={20}
        size={1.5}
        color="var(--border)"
      />
    </ReactFlow>
  )
}

export const ConfigFlow = () => {
  const { data: profiles, isLoading: profilesLoading } = useProfiles()
  const { data: callers, isLoading: callersLoading } = useCallers()
  const { data: activeCallers, isLoading: activeCallersLoading } =
    useActiveCallers()
  const { data: profileDetails, isLoading: detailsLoading } =
    useProfileDetails(profiles)
  const { data: runtimeStatus, isLoading: runtimeStatusLoading } = useRuntimeStatus()

  const isLoading =
    profilesLoading || callersLoading || detailsLoading || activeCallersLoading || runtimeStatusLoading
  const { nodes, edges, profileCount, serverCount, callerCount, instanceCount } =
    buildTopology(
      profiles ?? [],
      profileDetails ?? [],
      callers ?? {},
      activeCallers ?? [],
      runtimeStatus ?? [],
    )

  if (isLoading) {
    return (
      <div className="flex h-full flex-col">
        <div className="flex items-center justify-between border-b px-4 py-3">
          <div className="flex items-center gap-2">
            <Share2Icon className="size-4 text-muted-foreground" />
            <span className="text-sm font-medium">Topology</span>
          </div>
          <div className="flex items-center gap-2 text-xs text-muted-foreground">
            <Skeleton className="h-4 w-16" />
            <Skeleton className="h-4 w-16" />
            <Skeleton className="h-4 w-16" />
          </div>
        </div>
        <div className="flex-1 p-4">
          <FlowSkeleton />
        </div>
      </div>
    )
  }

  const hasData = profileCount > 0 || serverCount > 0 || callerCount > 0 || instanceCount > 0

  if (!hasData) {
    return (
      <div className="flex h-full flex-col">
        <div className="flex items-center justify-between border-b px-4 py-3">
          <div className="flex items-center gap-2">
            <Share2Icon className="size-4 text-muted-foreground" />
            <span className="text-sm font-medium">Topology</span>
          </div>
        </div>
        <div className="flex-1 p-4">
          <FlowEmpty />
        </div>
      </div>
    )
  }

  return (
    <div className="flex h-full flex-col">
      <div className="flex-1">
        <div className="h-full rounded-xl">
          <ReactFlowProvider>
            <ConfigFlowInner nodes={nodes} edges={edges} />
          </ReactFlowProvider>
        </div>
      </div>
    </div>
  )
}
