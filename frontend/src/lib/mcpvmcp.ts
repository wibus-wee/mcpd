// Input: None (pure utility functions)
// Output: mcpvmcp command and config builders with support for server/tag modes and build options
// Position: Utility library for generating IDE connection configs and CLI snippets

export type ClientTarget = 'cursor' | 'claude' | 'vscode' | 'codex'
export type SelectorMode = 'server' | 'tag'
export type TransportType = 'stdio' | 'streamable-http'

export const defaultRpcAddress = 'unix:///tmp/mcpv.sock'

export type SelectorConfig = {
  mode: SelectorMode
  value: string
}

export type BuildOptions = {
  // General
  transport?: TransportType
  launchUIOnFail?: boolean
  caller?: string
  urlScheme?: string

  // RPC Settings
  rpcMaxRecvMsgSize?: number
  rpcMaxSendMsgSize?: number
  rpcKeepaliveTime?: number
  rpcKeepaliveTimeout?: number
  rpcTLSEnabled?: boolean
  rpcTLSCertFile?: string
  rpcTLSKeyFile?: string
  rpcTLSCAFile?: string

  // Streamable HTTP settings (for http transport)
  httpUrl?: string
  httpHeaders?: Record<string, string>
}

const resolveServerName = (selector: SelectorConfig) =>
  selector.mode === 'server'
    ? selector.value
    : `mcpv-${selector.value}`

const quoteCliArg = (value: string) => {
  if (!/[\s"]/.test(value)) {
    return value
  }
  return `"${value.replaceAll('"', '\\"')}"`
}

const escapeTomlString = (value: string) =>
  value.replaceAll('\\', '\\\\').replaceAll('"', '\\"')

const formatTomlInlineTable = (values: Record<string, string>) => {
  const entries = Object.entries(values)
    .map(([key, value]) => `"${escapeTomlString(key)}" = "${escapeTomlString(value)}"`)
    .join(', ')
  return `{ ${entries} }`
}

const buildStdioArgs = (selector: SelectorConfig, rpc = defaultRpcAddress, options: BuildOptions = {}) => {
  const args = selector.mode === 'tag'
    ? ['--tag', selector.value]
    : [selector.value]

  // RPC settings
  if (rpc && rpc !== defaultRpcAddress) {
    args.push('--rpc', rpc)
  }
  if (options.rpcMaxRecvMsgSize) {
    args.push('--rpc-max-recv', String(options.rpcMaxRecvMsgSize))
  }
  if (options.rpcMaxSendMsgSize) {
    args.push('--rpc-max-send', String(options.rpcMaxSendMsgSize))
  }
  if (options.rpcKeepaliveTime) {
    args.push('--rpc-keepalive-time', String(options.rpcKeepaliveTime))
  }
  if (options.rpcKeepaliveTimeout) {
    args.push('--rpc-keepalive-timeout', String(options.rpcKeepaliveTimeout))
  }
  if (options.rpcTLSEnabled) {
    args.push('--rpc-tls')
    if (options.rpcTLSCertFile) {
      args.push('--rpc-tls-cert', options.rpcTLSCertFile)
    }
    if (options.rpcTLSKeyFile) {
      args.push('--rpc-tls-key', options.rpcTLSKeyFile)
    }
    if (options.rpcTLSCAFile) {
      args.push('--rpc-tls-ca', options.rpcTLSCAFile)
    }
  }

  // General settings
  if (options.caller) {
    args.push('--caller', options.caller)
  }
  if (options.launchUIOnFail) {
    args.push('--launch-ui-on-fail')
  }
  if (options.urlScheme && options.urlScheme !== 'mcpv') {
    args.push('--url-scheme', options.urlScheme)
  }

  // Transport settings
  if (options.transport && options.transport !== 'stdio') {
    args.push('--transport', options.transport)
  }

  return args
}

const buildHttpServerEntry = (options: BuildOptions) => {
  const url = options.httpUrl ?? ''
  const headers = options.httpHeaders && Object.keys(options.httpHeaders).length > 0
    ? options.httpHeaders
    : undefined
  return headers ? { url, headers } : { url }
}

export function buildMcpCommand(path: string, selector: SelectorConfig, rpc = defaultRpcAddress, options: BuildOptions = {}) {
  const args = buildStdioArgs(selector, rpc, options)
  return [path, ...args].join(' ')
}

export function buildClientConfig(
  _target: ClientTarget,
  path: string,
  selector: SelectorConfig,
  rpc = defaultRpcAddress,
  options: BuildOptions = {},
) {
  const serverName = resolveServerName(selector)

  if (options.transport === 'streamable-http') {
    return JSON.stringify({
      mcpServers: {
        [serverName]: buildHttpServerEntry(options),
      },
    }, null, 2)
  }

  return JSON.stringify({
    mcpServers: {
      [serverName]: {
        command: path,
        args: buildStdioArgs(selector, rpc, options),
      },
    },
  }, null, 2)
}

export function buildCliSnippet(
  path: string,
  selector: SelectorConfig,
  rpc = defaultRpcAddress,
  tool: 'claude' | 'codex',
  options: BuildOptions = {},
) {
  const serverName = resolveServerName(selector)

  if (options.transport === 'streamable-http') {
    const url = options.httpUrl ?? ''
    const headerArgs = options.httpHeaders
      ? Object.entries(options.httpHeaders)
        .map(([key, value]) => `--header ${quoteCliArg(`${key}: ${value}`)}`)
        .join(' ')
      : ''
    const headerSuffix = headerArgs ? ` ${headerArgs}` : ''
    return `${tool} mcp add --transport http ${serverName} ${quoteCliArg(url)}${headerSuffix}`
  }

  const args = buildStdioArgs(selector, rpc, options).map(quoteCliArg).join(' ')
  if (tool === 'claude') {
    return `claude mcp add --transport stdio mcpv -- ${path} ${args}`
  }
  return `codex mcp add mcpv -- ${path} ${args}`
}

export function buildTomlConfig(path: string, selector: SelectorConfig, rpc = defaultRpcAddress, options: BuildOptions = {}) {
  const serverName = resolveServerName(selector)

  if (options.transport === 'streamable-http') {
    const url = options.httpUrl ?? ''
    const lines = [
      `[mcp_servers.${serverName}]`,
      `url = "${escapeTomlString(url)}"`,
    ]
    if (options.httpHeaders && Object.keys(options.httpHeaders).length > 0) {
      lines.push(`http_headers = ${formatTomlInlineTable(options.httpHeaders)}`)
    }
    return lines.join('\n')
  }

  const args = buildStdioArgs(selector, rpc, options)
  const argsArray = `args = ${JSON.stringify(args)}`
  return [
    `[mcp_servers.${serverName}]`,
    `command = "${path}"`,
    argsArray,
    // ``,
    // `[mcp_servers.${serverName}.env]`,
    // `# MY_ENV_VAR = "MY_ENV_VALUE"`,
  ].join('\n')
}
