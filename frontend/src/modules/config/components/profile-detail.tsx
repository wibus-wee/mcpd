// Input: ProfileDetail data
// Output: ProfileDetailView component showing full profile configuration
// Position: Main content area in config page

import type { ProfileDetail, ServerSpecDetail } from '@bindings/mcpd/internal/ui'
import {
  ChevronDownIcon,
  ClockIcon,
  CpuIcon,
  NetworkIcon,
  ServerIcon,
  SettingsIcon,
  TerminalIcon,
} from 'lucide-react'
import { m } from 'motion/react'
import { useState } from 'react'

import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from '@/components/ui/collapsible'
import { Separator } from '@/components/ui/separator'
import { Skeleton } from '@/components/ui/skeleton'
import { Spring } from '@/lib/spring'
import { cn } from '@/lib/utils'

interface ProfileDetailViewProps {
  profile: ProfileDetail | null | undefined
  isLoading: boolean
}

function RuntimeConfigSection({ profile }: { profile: ProfileDetail }) {
  const [isOpen, setIsOpen] = useState(false)
  const { runtime } = profile

  return (
    <Collapsible open={isOpen} onOpenChange={setIsOpen}>
      <Card>
        <CollapsibleTrigger className="w-full">
          <CardHeader className="pb-3">
            <div className="flex items-center justify-between">
              <CardTitle className="flex items-center gap-2 text-sm">
                <SettingsIcon className="size-4" />
                Runtime Configuration
              </CardTitle>
              <ChevronDownIcon
                className={cn(
                  'size-4 text-muted-foreground transition-transform',
                  isOpen && 'rotate-180',
                )}
              />
            </div>
          </CardHeader>
        </CollapsibleTrigger>
        <CollapsibleContent>
          <CardContent className="pt-0">
            <div className="grid grid-cols-2 gap-4 text-sm">
              <div className="space-y-3">
                <div>
                  <p className="text-muted-foreground text-xs">Route Timeout</p>
                  <p className="font-mono">{runtime.routeTimeoutSeconds}s</p>
                </div>
                <div>
                  <p className="text-muted-foreground text-xs">Ping Interval</p>
                  <p className="font-mono">{runtime.pingIntervalSeconds}s</p>
                </div>
                <div>
                  <p className="text-muted-foreground text-xs">Tool Refresh</p>
                  <p className="font-mono">{runtime.toolRefreshSeconds}s</p>
                </div>
              </div>
              <div className="space-y-3">
                <div>
                  <p className="text-muted-foreground text-xs">Caller Check</p>
                  <p className="font-mono">{runtime.callerCheckSeconds}s</p>
                </div>
                <div>
                  <p className="text-muted-foreground text-xs">Expose Tools</p>
                  <Badge variant={runtime.exposeTools ? 'success' : 'secondary'} size="sm">
                    {runtime.exposeTools ? 'Yes' : 'No'}
                  </Badge>
                </div>
                <div>
                  <p className="text-muted-foreground text-xs">Namespace Strategy</p>
                  <Badge variant="outline" size="sm">
                    {runtime.toolNamespaceStrategy || 'prefix'}
                  </Badge>
                </div>
              </div>
            </div>

            <Separator className="my-4" />

            <div className="space-y-3">
              <div className="flex items-center gap-2 text-xs text-muted-foreground">
                <NetworkIcon className="size-3" />
                RPC Configuration
              </div>
              <div className="grid grid-cols-2 gap-4 text-sm">
                <div>
                  <p className="text-muted-foreground text-xs">Listen Address</p>
                  <p className="font-mono text-xs truncate">{runtime.rpc.listenAddress}</p>
                </div>
                <div>
                  <p className="text-muted-foreground text-xs">Socket Mode</p>
                  <p className="font-mono">{runtime.rpc.socketMode || '0660'}</p>
                </div>
              </div>
            </div>
          </CardContent>
        </CollapsibleContent>
      </Card>
    </Collapsible>
  )
}

function ServerCard({ server, index }: { server: ServerSpecDetail, index: number }) {
  const [isOpen, setIsOpen] = useState(false)

  return (
    <m.div
      initial={{ opacity: 0, y: 10 }}
      animate={{ opacity: 1, y: 0 }}
      transition={Spring.smooth(0.3, index * 0.05)}
    >
      <Collapsible open={isOpen} onOpenChange={setIsOpen}>
        <Card>
          <CollapsibleTrigger className="w-full text-left">
            <CardHeader className="pb-3">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <ServerIcon className="size-4" />
                  <span className="font-mono font-medium">{server.name}</span>
                </div>
                <div className="flex items-center gap-2">
                  {server.persistent && (
                    <Badge variant="secondary" size="sm">Persistent</Badge>
                  )}
                  {server.sticky && (
                    <Badge variant="outline" size="sm">Sticky</Badge>
                  )}
                  <ChevronDownIcon
                    className={cn(
                      'size-4 text-muted-foreground transition-transform',
                      isOpen && 'rotate-180',
                    )}
                  />
                </div>
              </div>
            </CardHeader>
          </CollapsibleTrigger>
          <CollapsibleContent>
            <CardContent className="pt-0 space-y-4">
              {/* Command */}
              <div>
                <div className="flex items-center gap-2 text-xs text-muted-foreground mb-1">
                  <TerminalIcon className="size-3" />
                  Command
                </div>
                <div className="bg-muted/50 rounded-md p-2 font-mono text-xs overflow-x-auto">
                  {server.cmd.join(' ')}
                </div>
              </div>

              {/* Working Directory */}
              {server.cwd && (
                <div>
                  <p className="text-muted-foreground text-xs mb-1">Working Directory</p>
                  <p className="font-mono text-xs">{server.cwd}</p>
                </div>
              )}

              {/* Environment Variables */}
              {Object.keys(server.env).length > 0 && (
                <div>
                  <p className="text-muted-foreground text-xs mb-1">Environment Variables</p>
                  <div className="space-y-1">
                    {Object.entries(server.env).map(([key, value]) => (
                      <div key={key} className="flex items-center gap-2 text-xs">
                        <Badge variant="outline" size="sm" className="font-mono">
                          {key}
                        </Badge>
                        <span className="text-muted-foreground truncate">
                          {value.startsWith('${') ? value : '••••••'}
                        </span>
                      </div>
                    ))}
                  </div>
                </div>
              )}

              <Separator />

              {/* Settings Grid */}
              <div className="grid grid-cols-3 gap-4 text-sm">
                <div className="flex items-center gap-2">
                  <ClockIcon className="size-3 text-muted-foreground" />
                  <div>
                    <p className="text-muted-foreground text-xs">Idle Timeout</p>
                    <p className="font-mono">{server.idleSeconds}s</p>
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  <CpuIcon className="size-3 text-muted-foreground" />
                  <div>
                    <p className="text-muted-foreground text-xs">Max Concurrent</p>
                    <p className="font-mono">{server.maxConcurrent}</p>
                  </div>
                </div>
                <div>
                  <p className="text-muted-foreground text-xs">Min Ready</p>
                  <p className="font-mono">{server.minReady}</p>
                </div>
              </div>

              {/* Exposed Tools */}
              {server.exposeTools && server.exposeTools.length > 0 && (
                <div>
                  <p className="text-muted-foreground text-xs mb-1">Exposed Tools</p>
                  <div className="flex flex-wrap gap-1">
                    {server.exposeTools.map(tool => (
                      <Badge key={tool} variant="secondary" size="sm" className="font-mono">
                        {tool}
                      </Badge>
                    ))}
                  </div>
                </div>
              )}
            </CardContent>
          </CollapsibleContent>
        </Card>
      </Collapsible>
    </m.div>
  )
}

export function ProfileDetailView({ profile, isLoading }: ProfileDetailViewProps) {
  if (isLoading) {
    return (
      <div className="space-y-4">
        <Skeleton className="h-16 w-full" />
        <Skeleton className="h-32 w-full" />
        <Skeleton className="h-32 w-full" />
      </div>
    )
  }

  if (!profile) {
    return (
      <Card>
        <CardContent className="flex items-center justify-center h-64 text-muted-foreground">
          Profile not found
        </CardContent>
      </Card>
    )
  }

  return (
    <m.div
      className="space-y-4"
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      transition={Spring.smooth(0.3)}
    >
      {/* Profile Header */}
      <div className="flex items-center justify-between">
        <h2 className="font-semibold text-lg">{profile.name}</h2>
        <Badge variant="secondary">
          {profile.servers.length} server{profile.servers.length !== 1 ? 's' : ''}
        </Badge>
      </div>

      {/* Runtime Config */}
      <RuntimeConfigSection profile={profile} />

      {/* Servers */}
      <div className="space-y-3">
        <h3 className="font-medium text-sm flex items-center gap-2">
          <ServerIcon className="size-4" />
          Servers
        </h3>
        {profile.servers.length === 0 ? (
          <Card>
            <CardContent className="flex items-center justify-center h-32 text-muted-foreground">
              No servers configured
            </CardContent>
          </Card>
        ) : (
          profile.servers.map((server, index) => (
            <ServerCard key={server.name} server={server} index={index} />
          ))
        )}
      </div>
    </m.div>
  )
}
