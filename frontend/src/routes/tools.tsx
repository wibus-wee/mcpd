// Input: TanStack Router
// Output: Tools route component
// Position: /tools route

import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/tools')({
  component: ToolsPage,
})

function ToolsPage() {
  return (
    <div className="container py-6">
      <h1 className="mb-4 font-bold text-2xl">Tools</h1>
      <p className="text-muted-foreground">Tools management page</p>
    </div>
  )
}
