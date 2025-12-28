// Input: Config hooks, atoms, UI components
// Output: ConfigPage component - main configuration management view
// Position: Main page in config module

import { useAtom } from 'jotai'
import {
  FileIcon,
  FileSliders,
  FolderIcon,
  LayersIcon,
  UsersIcon,
} from 'lucide-react'
import { m } from 'motion/react'
import { useState } from 'react'

import { RefreshButton } from '@/components/custom'
import { Badge } from '@/components/ui/badge'
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from '@/components/ui/empty'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Separator } from '@/components/ui/separator'
import { Skeleton } from '@/components/ui/skeleton'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Spring } from '@/lib/spring'

import { selectedProfileNameAtom } from './atoms'
import { CallersList } from './components/callers-list'
import { ProfileDetailSheet } from './components/profile-detail-sheet'
import { ProfilesList } from './components/profiles-list'
import { useCallers, useConfigMode, useProfiles } from './hooks'

function ConfigHeader() {
  const { data: configMode, isLoading, mutate } = useConfigMode()
  const [isRefreshing, setIsRefreshing] = useState(false)

  const handleRefresh = async () => {
    setIsRefreshing(true)
    await mutate()
    setIsRefreshing(false)
  }

  if (isLoading) {
    return (
      <div className="flex items-center justify-between">
        <div className="space-y-2">
          <Skeleton className="h-7 w-48" />
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
        <div className="flex items-center gap-2.5">
          <FileSliders className="size-5 text-muted-foreground" />
          <h1 className="font-semibold text-xl">Configuration</h1>
          {configMode && (
            <Badge variant="secondary" className="gap-1">
              <ModeIcon className="size-3" />
              {configMode.mode === 'directory' ? 'Directory' : 'Single File'}
            </Badge>
          )}
          {configMode?.isWritable && (
            <Badge variant="outline" className="text-success border-success/30">
              Writable
            </Badge>
          )}
        </div>
        <p className="text-muted-foreground text-sm">
          Manage your profiles and caller mappings.
        </p>
      </div>
      <RefreshButton
        onClick={handleRefresh}
        isLoading={isRefreshing}
        tooltip="Refresh configuration"
      />
    </m.div>
  )
}

function ConfigTabs() {
  const [selectedProfileName, setSelectedProfileName] = useAtom(selectedProfileNameAtom)
  const [sheetOpen, setSheetOpen] = useState(false)
  const {
    data: profiles,
    isLoading: profilesLoading,
    mutate: mutateProfiles,
  } = useProfiles()
  const {
    data: callers,
    isLoading: callersLoading,
    mutate: mutateCallers,
  } = useCallers()

  const handleProfileSelect = (name: string) => {
    setSelectedProfileName(name)
    setSheetOpen(true)
  }

  const profileCount = profiles?.length ?? 0
  const callerCount = callers ? Object.keys(callers).length : 0

  return (
    <>
      <Tabs defaultValue="profiles" className="flex-1 flex flex-col min-h-0">
        <TabsList variant="underline" className="w-full justify-start border-b px-0">
          <TabsTrigger value="profiles" className="gap-2">
            <LayersIcon className="size-4" />
            Profiles
            {profileCount > 0 && (
              <Badge variant="secondary" size="sm">
                {profileCount}
              </Badge>
            )}
          </TabsTrigger>
          <TabsTrigger value="callers" className="gap-2">
            <UsersIcon className="size-4" />
            Callers
            {callerCount > 0 && (
              <Badge variant="secondary" size="sm">
                {callerCount}
              </Badge>
            )}
          </TabsTrigger>
        </TabsList>

        <TabsContent value="profiles" className="flex-1 min-h-0 mt-0 pt-4">
          <ScrollArea className="h-full" scrollFade>
            <ProfilesList
              profiles={profiles ?? []}
              isLoading={profilesLoading}
              selectedProfile={selectedProfileName}
              onSelect={handleProfileSelect}
              onRefresh={() => mutateProfiles()}
            />
          </ScrollArea>
        </TabsContent>

        <TabsContent value="callers" className="flex-1 min-h-0 mt-0 pt-4">
          <ScrollArea className="h-full" scrollFade>
            <CallersList
              callers={callers ?? {}}
              isLoading={callersLoading}
              onRefresh={() => mutateCallers()}
            />
          </ScrollArea>
        </TabsContent>
      </Tabs>

      <ProfileDetailSheet
        profileName={selectedProfileName}
        open={sheetOpen}
        onOpenChange={setSheetOpen}
      />
    </>
  )
}

function ConfigEmpty() {
  return (
    <Empty className="h-full border-0">
      <EmptyHeader>
        <EmptyMedia variant="icon">
          <FileSliders className="size-5" />
        </EmptyMedia>
        <EmptyTitle>No configuration loaded</EmptyTitle>
        <EmptyDescription>
          Start the server with a configuration file to see your profiles here.
        </EmptyDescription>
      </EmptyHeader>
    </Empty>
  )
}

export function ConfigPage() {
  const { data: configMode } = useConfigMode()
  const { data: profiles } = useProfiles()

  const hasConfig = configMode && profiles

  return (
    <div className="flex flex-col h-full p-6 gap-4">
      <ConfigHeader />
      <Separator />
      {hasConfig ? <ConfigTabs /> : <ConfigEmpty />}
    </div>
  )
}
