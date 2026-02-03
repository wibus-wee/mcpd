// Input: Dashboard components, tabs/alerts/buttons, core/app hooks
// Output: DashboardPage component - main dashboard view with insights
// Position: Main dashboard page in dashboard module

import { DebugService } from '@bindings/mcpv/internal/ui'
import {
  AlertCircleIcon,
  FileDownIcon,
  Loader2Icon,
  PlayIcon,
  RefreshCwIcon,
  ServerIcon,
  SquareIcon,
} from 'lucide-react'
import { m } from 'motion/react'
import { useState } from 'react'

import { ConnectIdeSheet } from '@/components/common/connect-ide-sheet'
import { UniversalEmptyState } from '@/components/common/universal-empty-state'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogPanel,
  DialogTitle,
} from '@/components/ui/dialog'
import { Separator } from '@/components/ui/separator'
import { toastManager } from '@/components/ui/toast'
import { useCoreActions, useCoreState } from '@/hooks/use-core-state'
import { Spring } from '@/lib/spring'

import { ActiveClientsPanel } from './components/active-clients-panel'
import { ActivityInsights } from './components/activity-insights'
import { BootstrapProgressPanel } from './components/bootstrap-progress'
import { ServerHealthOverview } from './components/server-health-overview'
import { StatusCards } from './components/status-cards'
import { useAppInfo, useBootstrapProgress } from './hooks'

function DashboardHeader() {
  const { appInfo } = useAppInfo()
  const { coreStatus } = useCoreState()
  const {
    refreshCoreState,
    restartCore,
    startCore,
    stopCore,
  } = useCoreActions()
  const [isExporting, setIsExporting] = useState(false)
  const [debugData, setDebugData] = useState<string | null>(null)
  const appLabel = appInfo?.name
    ? `${appInfo.name} · ${appInfo.version === 'dev' ? 'dev' : `v${appInfo.version}`} (${appInfo.build})`
    : 'mcpv'

  const handleExportDebug = async () => {
    if (isExporting) {
      return
    }
    setIsExporting(true)
    try {
      const result = await DebugService.ExportDebugSnapshot()
      const payload = JSON.stringify(result.snapshot, null, 2)

      // Try clipboard first
      try {
        await navigator.clipboard.writeText(payload)
        toastManager.add({
          type: 'success',
          title: 'Debug snapshot copied',
          description: 'Copied to clipboard',
        })
      }
      catch {
        // Fallback: show dialog for manual copy
        setDebugData(payload)
      }
    }
    catch (err) {
      toastManager.add({
        type: 'error',
        title: 'Export failed',
        description: err instanceof Error ? err.message : 'Export failed',
      })
    }
    finally {
      setIsExporting(false)
    }
  }

  const handleCopyFromDialog = () => {
    if (!debugData) {
      return
    }

    const textarea = document.createElement('textarea')
    textarea.value = debugData
    textarea.style.position = 'fixed'
    textarea.style.opacity = '0'
    document.body.append(textarea)
    textarea.select()
    try {
      document.execCommand('copy')
      toastManager.add({
        type: 'success',
        title: 'Copied',
        description: 'Debug snapshot copied to clipboard',
      })
      setDebugData(null)
    }
    catch {
      toastManager.add({
        type: 'error',
        title: 'Copy failed',
        description: 'Please copy manually',
      })
    }
    finally {
      textarea.remove()
    }
  }

  return (
    <m.div
      initial={{ opacity: 0, y: 10, filter: 'blur(8px)' }}
      animate={{ opacity: 1, y: 0, filter: 'blur(0px)' }}
      transition={Spring.smooth(0.4)}
      className="flex items-center justify-between"
    >
      <div>
        <h1 className="text-2xl font-bold tracking-tight">Dashboard</h1>
        <p className="text-muted-foreground text-sm">{appLabel}</p>
      </div>
      <div className="flex items-center gap-2">
        {coreStatus === 'stopped' ? (
          <Button onClick={startCore} size="sm">
            <PlayIcon className="size-4" />
            Start Core
          </Button>
        ) : coreStatus === 'starting'
          ? (
            <Button onClick={stopCore} variant="outline" size="sm">
              <SquareIcon className="size-4" />
              Cancel
            </Button>
          )
          : coreStatus === 'stopping'
            ? (
              <Button variant="outline" size="sm" disabled>
                <Loader2Icon className="size-4 animate-spin" />
                Stopping...
              </Button>
            )
            : coreStatus === 'running'
              ? (
                <>
                  <Button onClick={stopCore} variant="outline" size="sm">
                    <SquareIcon className="size-4" />
                    Stop
                  </Button>
                  <Button onClick={restartCore} variant="outline" size="sm">
                    <RefreshCwIcon className="size-4" />
                    Restart
                  </Button>
                </>
              )
              : coreStatus === 'error'
                ? (
                  <>
                    <Button onClick={restartCore} size="sm">
                      <RefreshCwIcon className="size-4" />
                      Retry
                    </Button>
                    <Button onClick={stopCore} variant="outline" size="sm">
                      <SquareIcon className="size-4" />
                      Stop
                    </Button>
                  </>
                )
                : null}
        <Button
          variant="outline"
          size="sm"
          onClick={handleExportDebug}
          disabled={isExporting}
        >
          <FileDownIcon className="size-4" />
          {isExporting ? 'Copying...' : 'Copy Debug'}
        </Button>
        <Button
          variant="ghost"
          size="icon-sm"
          onClick={() => refreshCoreState()}
        >
          <RefreshCwIcon className="size-4" />
        </Button>
        <ConnectIdeSheet />
      </div>

      <Dialog open={!!debugData} onOpenChange={open => !open && setDebugData(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Debug Snapshot</DialogTitle>
            <DialogDescription>
              Copy the debug information below to share with support or for troubleshooting.
            </DialogDescription>
          </DialogHeader>
          <DialogPanel>
            <pre className="rounded-lg bg-muted p-4 text-xs overflow-x-auto">
              <code>{debugData}</code>
            </pre>
          </DialogPanel>
          <DialogFooter>
            <DialogClose render={<Button variant="outline" />}>
              Close
            </DialogClose>
            <Button onClick={handleCopyFromDialog}>
              Copy to Clipboard
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </m.div>
  )
}

function DashboardInsights() {
  return (
    <m.div
      initial={{ opacity: 0, y: 10, filter: 'blur(4px)' }}
      animate={{ opacity: 1, y: 0, filter: 'blur(0px)' }}
      transition={Spring.smooth(0.4, 0.1)}
      className="grid gap-4 lg:grid-cols-3"
    >
      <div className="lg:col-span-2 space-y-4">
        <ServerHealthOverview />
      </div>
      <div className="space-y-4">
        <ActiveClientsPanel />
      </div>
      <div className="lg:col-span-3">
        <ActivityInsights />
      </div>
    </m.div>
  )
}

function DashboardContent() {
  return (
    <m.div
      initial={{ opacity: 0, y: 10, filter: 'blur(8px)' }}
      animate={{ opacity: 1, y: 0, filter: 'blur(0px)' }}
      transition={Spring.smooth(0.4)}
      className="space-y-6"
    >
      <StatusCards />

      <DashboardInsights />
    </m.div>
  )
}

function StartingContent() {
  const { state, total } = useBootstrapProgress()

  if (total > 0 || state === 'running') {
    return (
      <div className="flex flex-1 flex-col items-center justify-center gap-6">
        <BootstrapProgressPanel className="w-full max-w-md" />
      </div>
    )
  }

  return (
    <UniversalEmptyState
      icon={Loader2Icon}
      iconClassName="animate-spin"
      title="Starting Core..."
      description="Please wait while the mcpv core is initializing."
    />
  )
}

export function DashboardPage() {
  const { coreStatus, data: coreState } = useCoreState()
  const { startCore } = useCoreActions()

  if (coreStatus === 'stopped') {
    return (
      <div className="flex flex-1 flex-col p-6 overflow-auto">
        <DashboardHeader />
        <Separator className="my-6" />
        <UniversalEmptyState
          icon={ServerIcon}
          title="Core is not running"
          description="Start the mcpv core to see your dashboard and manage MCP servers."
          action={{
            label: 'Start Core',
            onClick: startCore,
          }}
        />
      </div>
    )
  }

  if (coreStatus === 'starting') {
    return (
      <div className="flex flex-1 flex-col p-6 overflow-auto">
        <DashboardHeader />
        <Separator className="my-6" />
        <StartingContent />
      </div>
    )
  }

  if (coreStatus === 'error') {
    return (
      <div className="flex flex-1 flex-col p-6 overflow-auto">
        <DashboardHeader />
        <Separator className="my-6" />
        <m.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={Spring.smooth(0.4)}
        >
          <Alert variant="error">
            <AlertCircleIcon className="size-4" />
            <AlertTitle>Core Error</AlertTitle>
            <AlertDescription>
              {coreState?.error || 'The mcpv core encountered an error. Check the logs for details.'}
            </AlertDescription>
          </Alert>
        </m.div>
      </div>
    )
  }

  return (
    <div className="flex flex-1 flex-col p-6 overflow-scroll">
      <DashboardHeader />
      <Separator className="my-6" />
      <DashboardContent />
    </div>
  )
}
