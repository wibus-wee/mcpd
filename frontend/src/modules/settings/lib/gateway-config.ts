// Input: None
// Output: Gateway settings form types, defaults, and mappers
// Position: Settings module config helpers for gateway settings UI

export type GatewayVisibilityMode = 'all' | 'tags' | 'server'
export type GatewayAccessMode = 'local' | 'network'

export type GatewayFormState = {
  enabled: boolean
  visibilityMode: GatewayVisibilityMode
  tagsInput: string
  serverName: string
  accessMode: GatewayAccessMode
  httpAddr: string
  httpPath: string
  httpToken: string
  caller: string
  binaryPath: string
  customArgs: string
  rpc: string
  healthUrl: string
}

export const GATEWAY_SECTION_KEY = 'gateway'

export const GATEWAY_VISIBILITY_OPTIONS = [
  { value: 'all', label: 'All servers' },
  { value: 'tags', label: 'By tags' },
  { value: 'server', label: 'Single server' },
] as const

export const GATEWAY_ACCESS_OPTIONS = [
  { value: 'local', label: 'Local only' },
  { value: 'network', label: 'Network (requires token)' },
] as const

const defaultHTTPAddr = '127.0.0.1:8090'
const defaultHTTPPath = '/mcp'
const defaultCaller = 'mcpvmcp-ui'

export const DEFAULT_GATEWAY_FORM: GatewayFormState = {
  enabled: true,
  visibilityMode: 'all',
  tagsInput: '',
  serverName: '',
  accessMode: 'local',
  httpAddr: defaultHTTPAddr,
  httpPath: defaultHTTPPath,
  httpToken: '',
  caller: defaultCaller,
  binaryPath: '',
  customArgs: '',
  rpc: '',
  healthUrl: '',
}

type GatewaySectionPayload = {
  enabled?: boolean
  binaryPath?: string
  args?: string[]
  httpAddr?: string
  httpPath?: string
  httpToken?: string
  caller?: string
  rpc?: string
  server?: string
  tags?: string[]
  allowAll?: boolean
  healthUrl?: string
}

export function toGatewayFormState(section: unknown): GatewayFormState {
  const payload = resolveSectionObject(section)
  if (!payload) return DEFAULT_GATEWAY_FORM

  const httpAddr = toStringOrDefault(payload.httpAddr, defaultHTTPAddr)
  const httpPath = normalizeHTTPPath(toStringOrDefault(payload.httpPath, defaultHTTPPath))
  const tags = Array.isArray(payload.tags) ? payload.tags.filter(Boolean) : []
  const serverName = toStringOrDefault(payload.server, '')
  // const allowAll = typeof payload.allowAll === 'boolean' ? payload.allowAll : true

  const visibilityMode: GatewayVisibilityMode = serverName
    ? 'server'
    : tags.length > 0
      ? 'tags'
      : 'all'

  const accessMode: GatewayAccessMode = isLocalHost(httpAddr) ? 'local' : 'network'

  return {
    enabled: typeof payload.enabled === 'boolean' ? payload.enabled : true,
    visibilityMode,
    tagsInput: tags.join(', '),
    serverName,
    accessMode,
    httpAddr,
    httpPath,
    httpToken: toStringOrDefault(payload.httpToken, ''),
    caller: toStringOrDefault(payload.caller, defaultCaller),
    binaryPath: toStringOrDefault(payload.binaryPath, ''),
    customArgs: Array.isArray(payload.args) ? payload.args.join('\n') : '',
    rpc: toStringOrDefault(payload.rpc, ''),
    healthUrl: toStringOrDefault(payload.healthUrl, ''),
  }
}

export function buildGatewayPayload(values: GatewayFormState): GatewaySectionPayload {
  const tags = parseTags(values.tagsInput)

  const payload: GatewaySectionPayload = {
    enabled: Boolean(values.enabled),
    binaryPath: values.binaryPath.trim(),
    args: parseArgs(values.customArgs),
    httpAddr: values.httpAddr.trim() || defaultHTTPAddr,
    httpPath: normalizeHTTPPath(values.httpPath),
    httpToken: values.httpToken.trim(),
    caller: values.caller.trim() || defaultCaller,
    rpc: values.rpc.trim(),
    healthUrl: values.healthUrl.trim(),
    allowAll: values.visibilityMode === 'all',
    server: values.visibilityMode === 'server' ? values.serverName.trim() : '',
    tags: values.visibilityMode === 'tags' ? tags : [],
  }

  if (values.visibilityMode === 'all') {
    payload.server = ''
    payload.tags = []
  }

  return payload
}

export function buildEndpointPreview(httpAddr: string, httpPath: string) {
  const addr = httpAddr.trim() || defaultHTTPAddr
  const path = normalizeHTTPPath(httpPath)
  const base = addr.includes('://') ? addr : `http://${addr}`
  return `${base}${path}`
}

export function isLocalHost(addr: string) {
  const host = extractHost(addr)
  if (!host) return true
  return host === 'localhost' || host === '127.0.0.1' || host === '::1'
}

export function normalizeHTTPPath(path: string) {
  const trimmed = path.trim()
  if (!trimmed) return defaultHTTPPath
  if (trimmed.startsWith('/')) return trimmed
  return `/${trimmed}`
}

function resolveSectionObject(section: unknown): GatewaySectionPayload | null {
  if (!section) return null
  if (typeof section === 'string') {
    try {
      const parsed = JSON.parse(section) as unknown
      if (parsed && typeof parsed === 'object') {
        return parsed as GatewaySectionPayload
      }
    }
    catch {
      return null
    }
  }
  if (typeof section === 'object') {
    return section as GatewaySectionPayload
  }
  return null
}

function toStringOrDefault(value: unknown, fallback: string) {
  if (typeof value === 'string' && value.trim() !== '') return value
  return fallback
}

function parseArgs(input: string) {
  const trimmed = input.trim()
  if (!trimmed) return []
  return trimmed.split('\n').map(line => line.trim()).filter(Boolean)
}

function parseTags(input: string) {
  const trimmed = input.trim()
  if (!trimmed) return []
  return trimmed
    .split(',')
    .map(tag => tag.trim())
    .filter(Boolean)
}

function extractHost(addr: string) {
  const trimmed = addr.trim()
  if (!trimmed) return ''
  if (trimmed.includes('://')) {
    try {
      return new URL(trimmed).hostname.toLowerCase()
    }
    catch {
      return ''
    }
  }
  const lastColon = trimmed.lastIndexOf(':')
  const host = lastColon !== -1 ? trimmed.slice(0, lastColon) : trimmed
  return host.toLowerCase()
}
