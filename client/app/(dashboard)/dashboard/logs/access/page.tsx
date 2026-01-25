'use client'

import { useState, useEffect } from 'react'
import { RefreshCw } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { toast } from 'sonner'
import { Domain, AccessLog } from '@/types/domains'

export default function AccessLogsPage() {
  const [domains, setDomains] = useState<Domain[]>([])
  const [selectedDomainId, setSelectedDomainId] = useState<string>('all')
  const [logs, setLogs] = useState<AccessLog[]>([])
  const [isLoading, setIsLoading] = useState(true)

  useEffect(() => {
    fetchDomains()
    fetchLogs('all')
  }, [])

  useEffect(() => {
    fetchLogs(selectedDomainId)
  }, [selectedDomainId])

  const fetchDomains = async () => {
    try {
      const response = await fetch('/api/domains', {
        headers: { 'Content-Type': 'application/json' },
      })

      if (!response.ok) throw new Error('Failed to fetch domains')

      const data = await response.json()
      setDomains(Array.isArray(data) ? data : [])
    } catch (error) {
      console.error(error)
      toast.error('Failed to load domains')
    }
  }

  const fetchLogs = async (domainId: string) => {
    setIsLoading(true)
    try {
      const url = domainId === 'all'
        ? '/api/logs?limit=100'
        : `/api/domains/${domainId}/logs?limit=100`

      const response = await fetch(url, {
        headers: { 'Content-Type': 'application/json' },
      })

      if (!response.ok) throw new Error('Failed to fetch logs')

      const data = await response.json()
      setLogs(Array.isArray(data) ? data : [])
    } catch (error) {
      console.error(error)
      toast.error('Failed to load access logs')
    } finally {
      setIsLoading(false)
    }
  }

  const getStatusBadge = (status: number) => {
    if (status >= 500) return <Badge variant="destructive">{status}</Badge>
    if (status >= 400) return <Badge className="bg-orange-500">{status}</Badge>
    if (status >= 300) return <Badge variant="secondary">{status}</Badge>
    return <Badge className="bg-green-500">{status}</Badge>
  }

  const getMethodBadge = (method: string) => {
    const colors: Record<string, string> = {
      GET: 'bg-blue-500',
      POST: 'bg-green-500',
      PUT: 'bg-yellow-500',
      DELETE: 'bg-red-500',
      PATCH: 'bg-purple-500',
    }
    return <Badge className={colors[method] || 'bg-gray-500'}>{method}</Badge>
  }

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-3xl font-bold">Access Logs</h1>
          <p className="text-muted-foreground">View request logs for your domains</p>
        </div>
        <Button onClick={() => fetchLogs(selectedDomainId)} variant="outline">
          <RefreshCw className="mr-2 h-4 w-4" />
          Refresh
        </Button>
      </div>

      <div className="flex items-center gap-4">
        <span className="text-sm font-medium">Filter by domain:</span>
        <Select value={selectedDomainId} onValueChange={setSelectedDomainId}>
          <SelectTrigger className="w-[250px]">
            <SelectValue placeholder="All domains" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All Domains</SelectItem>
            {domains.map((domain) => (
              <SelectItem key={domain.id} value={String(domain.id)}>
                {domain.name}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Recent Requests</CardTitle>
          <CardDescription>Last 100 access log entries</CardDescription>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="flex justify-center p-6">Loading...</div>
          ) : logs.length === 0 ? (
            <Alert>
              <AlertTitle>No logs found</AlertTitle>
              <AlertDescription>No access logs recorded yet.</AlertDescription>
            </Alert>
          ) : (
            <div className="overflow-x-auto">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Time</TableHead>
                    <TableHead>Method</TableHead>
                    <TableHead>Path</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead>Response Time</TableHead>
                    <TableHead>Client IP</TableHead>
                    <TableHead>User Agent</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {logs.map((log) => (
                    <TableRow key={log.id}>
                      <TableCell className="whitespace-nowrap">
                        {new Date(log.created_at).toLocaleString()}
                      </TableCell>
                      <TableCell>{getMethodBadge(log.method)}</TableCell>
                      <TableCell className="font-mono text-sm max-w-xs truncate">
                        {log.path}
                      </TableCell>
                      <TableCell>{getStatusBadge(log.status_code)}</TableCell>
                      <TableCell>{log.response_time_ms}ms</TableCell>
                      <TableCell className="font-mono text-sm">{log.client_ip}</TableCell>
                      <TableCell className="max-w-xs truncate text-sm text-muted-foreground">
                        {log.user_agent}
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
