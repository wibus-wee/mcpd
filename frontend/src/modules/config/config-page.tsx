// Input: Config hooks, atoms, UI components
// Output: ConfigPage component - main configuration management view
// Position: Main page in config module

import { useAtom } from 'jotai'
import {
  FileIcon,
  FileSliders,
  FolderIcon,
  RefreshCwIcon,
} from 'lucide-react'
import { m } from 'motion/react'

import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Separator } from '@/components/ui/separator'
import { Skeleton } from '@/components/ui/skeleton'
import { Spring } from '@/lib/spring'

import { selectedProfileNameAtom } from './atoms'
import { CallersPanel } from './components/callers-panel'
import { ProfileDetailView } from './components/profile-detail'
import { ProfileList } from './components/profile-list'
import { useCallers, useConfigMode, useProfile, useProfiles } from './hooks'

function ConfigHeader() {
  const { data: configMode, isLoading, mutate } = useConfigMode()

  if (isLoading) {
    return (
      <div className="flex items-center justify-between">
        <div className="space-y-1">
          <Skeleton className="h-6 w-48" />
          <Skeleton className="h-4 w-64" />
        </div>
      </div>
    )
  }

  const ModeIcon = configMode?.mode === 'directory' ? FolderIcon : FileIcon

  return (
    <m.div
      className="flex items-center justify-between"
      initial={{ opacity: 0, y: -10 }}
      animate={{ opacity: 1, y: 0 }}
      transition={Spring.smooth(0.3)}
    >
      <div className="space-y-1">
        <div className="flex items-center gap-2">
          <FileSliders className="size-5" />
          <h1 className="font-semibold text-xl">Configuration</h1>
          {configMode && (
            <Badge variant="secondary" className="gap-1">
              <ModeIcon className="size-3" />
              {configMode.mode === 'directory' ? 'Directory' : 'Single File'}
            </Badge>
          )}
          {configMode?.isWritable && (
            <Badge variant="outline" className="text-success">
              Writable
            </Badge>
          )}
        </div>
        <p className="text-muted-foreground text-xs">
          Find and manage your configuration profiles here.
        </p>
      </div>
      <Button
        variant="ghost"
        size="icon-sm"
        onClick={() => mutate()}
      >
        <RefreshCwIcon className="size-4" />
      </Button>
    </m.div>
  )
}

function ConfigContent() {
  const [selectedProfileName, setSelectedProfileName] = useAtom(selectedProfileNameAtom)
  const { data: profiles, isLoading: profilesLoading, mutate: mutateProfiles } = useProfiles()
  const { data: profile, isLoading: profileLoading } = useProfile(selectedProfileName)
  const { data: callers, isLoading: callersLoading, mutate: mutateCallers } = useCallers()

  // Auto-select default profile if none selected
  if (!selectedProfileName && profiles && profiles.length > 0) {
    const defaultProfile = profiles.find(p => p.isDefault) || profiles[0]
    setSelectedProfileName(defaultProfile.name)
  }

  return (
    <div className="grid grid-cols-[280px_1fr] gap-6 h-[calc(100vh-12rem)]">
      {/* Left sidebar */}
      <div className="space-y-4 overflow-y-auto">
        <ProfileList
          profiles={profiles || []}
          selectedProfile={selectedProfileName}
          onSelect={setSelectedProfileName}
          isLoading={profilesLoading}
          onRefresh={() => mutateProfiles()}
        />
        <CallersPanel
          callers={callers || {}}
          isLoading={callersLoading}
          onRefresh={() => mutateCallers()}
        />
      </div>

      {/* Main content */}
      <div className="overflow-y-auto">
        {selectedProfileName ? (
          <ProfileDetailView
            profile={profile}
            isLoading={profileLoading}
          />
        ) : (
          <Card>
            <CardContent className="flex items-center justify-center h-64 text-muted-foreground">
              Select a profile to view its configuration
            </CardContent>
          </Card>
        )}
      </div>
    </div>
  )
}

export function ConfigPage() {
  return (
    <div className="space-y-6 p-6">
      <ConfigHeader />
      <Separator />
      <ConfigContent />
    </div>
  )
}
