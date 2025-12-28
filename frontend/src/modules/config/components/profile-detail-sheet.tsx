// Input: Profile name, open state
// Output: ProfileDetailSheet - side panel showing profile details
// Position: Sheet overlay for profile configuration details

import type { ProfileDetail, ServerSpecDetail } from '@bindings/mcpd/internal/ui'
import {
  ClockIcon,
  CpuIcon,
  NetworkIcon,
  ServerIcon,
  SettingsIcon,
  StarIcon,
  TerminalIcon,
} from 'lucide-react'
import { m } from 'motion/react'

import { SectionHeader } from '@/components/custom'
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from '@/components/ui/accordion'
import { Badge } from '@/components/ui/badge'
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from '@/components/ui/empty'
import { Separator } from '@/components/ui/separator'
import {
  Sheet,
  SheetDescription,
  SheetHeader,
  SheetPanel,
  SheetPopup,
  SheetTitle,
} from '@/components/ui/sheet'
import { Skeleton } from '@/components/ui/skeleton'
import { Spring } from '@/lib/spring'

import { useProfile, useProfiles } from '../hooks'

interface ProfileDetailSheetProps {
  profileName: string | null
  open: boolean
  onOpenChange: (open: boolean) => void
}

function DetailRow({
  label,
  value,
  mono = false,
}: {
  label: string
  value: React.ReactNode
  mono?: boolean
}) {
  return (
    <div className="flex items-center justify-between py-2">
      <span className="text-muted-foreground text-sm">{label}</span>
      <span className={mono ? 'font-mono text-sm' : 'text-sm'}>{value}</span>
    </div>
  )
}

function RuntimeSection({ profile }: { profile: ProfileDetail }) {
  const { runtime } = profile

  return (
    <AccordionItem value="runtime">
      <AccordionTrigger>
        <div className="flex items-center gap-2">
          <SettingsIcon className="size-4 text-muted-foreground" />
          <span>Runtime Configuration</span>
        </div>
      </AccordionTrigger>
      <AccordionContent>
        <div className="space-y-1 divide-y divide-border/50">
          <DetailRow label="Route Timeout" value={`${runtime.routeTimeoutSeconds}s`} mono />
          <DetailRow label="Ping Interval" value={`${runtime.pingIntervalSeconds}s`} mono />
          <DetailRow label="Tool Refresh" value={`${runtime.toolRefreshSeconds}s`} mono />
          <DetailRow label="Caller Check" value={`${runtime.callerCheckSeconds}s`} mono />
          <DetailRow
            label="Expose Tools"
            value={
              <Badge variant={runtime.exposeTools ? 'success' : 'secondary'} size="sm">
                {runtime.exposeTools ? 'Yes' : 'No'}
              </Badge>
            }
          />
          <DetailRow
            label="Namespace Strategy"
            value={
              <Badge variant="outline" size="sm">
                {runtime.toolNamespaceStrategy || 'prefix'}
              </Badge>
            }
          />
        </div>

        <div className="mt-4 pt-4 border-t">
          <div className="flex items-center gap-2 text-xs text-muted-foreground mb-3">
            <NetworkIcon className="size-3" />
            RPC Configuration
          </div>
          <div className="space-y-1 divide-y divide-border/50">
            <DetailRow
              label="Listen Address"
              value={
                <span className="font-mono text-xs truncate max-w-45 block text-right">
                  {runtime.rpc.listenAddress}
                </span>
              }
            />
            <DetailRow label="Socket Mode" value={runtime.rpc.socketMode || '0660'} mono />
          </div>
        </div>
      </AccordionContent>
    </AccordionItem>
  )
}

function ServerSection({ server }: { server: ServerSpecDetail }) {
  return (
    <AccordionItem value={`server-${server.name}`}>
      <AccordionTrigger>
        <div className="flex items-center gap-2 min-w-0 flex-1">
          <ServerIcon className="size-4 text-muted-foreground shrink-0" />
          <span className="font-mono truncate">{server.name}</span>
          <div className="flex items-center gap-1.5 ml-auto mr-2">
            {server.persistent && (
              <Badge variant="secondary" size="sm">Persistent</Badge>
            )}
            {server.sticky && (
              <Badge variant="outline" size="sm">Sticky</Badge>
            )}
          </div>
        </div>
      </AccordionTrigger>
      <AccordionContent>
        <div className="space-y-4">
          {/* Command */}
          <div>
            <div className="flex items-center gap-2 text-xs text-muted-foreground mb-2">
              <TerminalIcon className="size-3" />
              Command
            </div>
            <div className="bg-muted/50 rounded-md p-3 font-mono text-xs overflow-x-auto">
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
              <p className="text-muted-foreground text-xs mb-2">Environment Variables</p>
              <div className="space-y-1.5">
                {Object.entries(server.env).map(([key, value]) => (
                  <div key={key} className="flex items-center gap-2 text-xs">
                    <Badge variant="outline" size="sm" className="font-mono shrink-0">
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
          <div className="grid grid-cols-3 gap-4">
            <div className="flex items-center gap-2">
              <ClockIcon className="size-3 text-muted-foreground" />
              <div>
                <p className="text-muted-foreground text-xs">Idle</p>
                <p className="font-mono text-sm">{server.idleSeconds}s</p>
              </div>
            </div>
            <div className="flex items-center gap-2">
              <CpuIcon className="size-3 text-muted-foreground" />
              <div>
                <p className="text-muted-foreground text-xs">Max</p>
                <p className="font-mono text-sm">{server.maxConcurrent}</p>
              </div>
            </div>
            <div>
              <p className="text-muted-foreground text-xs">Min Ready</p>
              <p className="font-mono text-sm">{server.minReady}</p>
            </div>
          </div>

          {/* Exposed Tools */}
          {server.exposeTools && server.exposeTools.length > 0 && (
            <div>
              <p className="text-muted-foreground text-xs mb-2">Exposed Tools</p>
              <div className="flex flex-wrap gap-1">
                {server.exposeTools.map(tool => (
                  <Badge key={tool} variant="secondary" size="sm" className="font-mono">
                    {tool}
                  </Badge>
                ))}
              </div>
            </div>
          )}
        </div>
      </AccordionContent>
    </AccordionItem>
  )
}

function ProfileContent({ profile }: { profile: ProfileDetail }) {
  return (
    <m.div
      className="space-y-6"
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      transition={Spring.smooth(0.3)}
    >
      <Accordion multiple defaultValue={['runtime']}>
        <RuntimeSection profile={profile} />
      </Accordion>

      <div className="space-y-3">
        <SectionHeader
          icon={<ServerIcon className="size-4" />}
          title="Servers"
          badge={
            <Badge variant="secondary" size="sm">
              {profile.servers.length}
            </Badge>
          }
        />

        {profile.servers.length === 0 ? (
          <Empty className="py-8">
            <EmptyHeader>
              <EmptyMedia variant="icon">
                <ServerIcon className="size-4" />
              </EmptyMedia>
              <EmptyTitle className="text-base">No servers</EmptyTitle>
              <EmptyDescription>
                Add servers to this profile in your configuration file.
              </EmptyDescription>
            </EmptyHeader>
          </Empty>
        ) : (
          <Accordion multiple>
            {profile.servers.map(server => (
              <ServerSection key={server.name} server={server} />
            ))}
          </Accordion>
        )}
      </div>
    </m.div>
  )
}

function ProfileSheetSkeleton() {
  return (
    <div className="space-y-4">
      <Skeleton className="h-12 w-full" />
      <Skeleton className="h-32 w-full" />
      <Skeleton className="h-32 w-full" />
    </div>
  )
}

export function ProfileDetailSheet({
  profileName,
  open,
  onOpenChange,
}: ProfileDetailSheetProps) {
  const { data: profile, isLoading } = useProfile(profileName)
  const { data: profiles } = useProfiles()

  // Get isDefault from the profiles list since ProfileDetail doesn't have it
  const profileSummary = profiles?.find(p => p.name === profileName)
  const isDefault = profileSummary?.isDefault ?? false

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetPopup side="right" className="w-120 max-w-[90vw]">
        <SheetHeader>
          <div className="flex items-center gap-2">
            <SheetTitle>{profileName || 'Profile'}</SheetTitle>
            {isDefault && (
              <StarIcon className="size-4 fill-warning text-warning" />
            )}
          </div>
          <SheetDescription>
            {profile?.servers.length ?? 0} server{(profile?.servers.length ?? 0) !== 1 ? 's' : ''} configured
          </SheetDescription>
        </SheetHeader>
        <SheetPanel>
          {isLoading ? (
            <ProfileSheetSkeleton />
          ) : profile ? (
            <ProfileContent profile={profile} />
          ) : (
            <Empty className="py-12">
              <EmptyHeader>
                <EmptyTitle>Profile not found</EmptyTitle>
                <EmptyDescription>
                  The selected profile could not be loaded.
                </EmptyDescription>
              </EmptyHeader>
            </Empty>
          )}
        </SheetPanel>
      </SheetPopup>
    </Sheet>
  )
}
