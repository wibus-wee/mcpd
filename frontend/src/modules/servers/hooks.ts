// Input: Config hooks, tools hooks, SWR
// Output: Combined data hooks for servers module
// Position: Data layer for unified servers module

export { useServers, useServer, useConfigMode, useRuntimeStatus, useServerInitStatus, useOpenConfigInEditor } from '@/modules/config/hooks'
export { useToolsByServer, type ServerGroup } from '@/modules/tools/hooks'
export { useActiveClients } from '@/hooks/use-active-clients'
