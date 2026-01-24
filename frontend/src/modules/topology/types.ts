// Input: React Flow types, UI binding types
// Output: Type definitions for topology flow nodes and layout
// Position: Shared types for topology module

import type {
  ActiveCaller,
  ProfileDetail,
  ProfileSummary,
  ServerRuntimeStatus,
} from '@bindings/mcpd/internal/ui'
import type { Node, NodeProps } from '@xyflow/react'

export type CallerNodeData = {
  name: string
  profileName?: string
  pid?: number
}

export type ProfileNodeData = {
  name: string
  serverCount: number
  isDefault: boolean
  isMissing: boolean
}

export type ServerNodeData = {
  name: string
  protocolVersion: string
  tags: string[]
}

export type InstanceNodeData = {
  id: string
  state: string
  busyCount: number
}

export type CallerFlowNode = Node<CallerNodeData, 'caller'>
export type ProfileFlowNode = Node<ProfileNodeData, 'profile'>
export type ServerFlowNode = Node<ServerNodeData, 'server'>
export type InstanceFlowNode = Node<InstanceNodeData, 'instance'>
export type FlowNode = CallerFlowNode | ProfileFlowNode | ServerFlowNode | InstanceFlowNode

export interface CallerNodeProps extends NodeProps<CallerFlowNode> {
  isActive?: boolean
}

export type LayoutResult = {
  nodes: FlowNode[]
  edges: import('@xyflow/react').Edge[]
  profileCount: number
  serverCount: number
  callerCount: number
  instanceCount: number
}

export type BuildTopologyInput = {
  profiles: ProfileSummary[]
  profileDetails: ProfileDetail[]
  callers: Record<string, string>
  activeCallers: ActiveCaller[]
  runtimeStatus: ServerRuntimeStatus[]
}
