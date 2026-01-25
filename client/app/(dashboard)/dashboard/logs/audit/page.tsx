'use client'

import { useState, useEffect } from 'react'
import { RefreshCw } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { toast } from 'sonner'
import { AuditLog } from '@/types/domains'

export default function AuditLogsPage() {
  const [logs, setLogs] = useState<AuditLog[]>([])
  const [isLoading, setIsLoading] = useState(true)

  useEffect(() => {
    fetchLogs()
  }, [])

  const fetchLogs = async () => {
    setIsLoading(true)
    try {
      const response = await fetch('/api/audit?limit=100', {
        headers: { 'Content-Type': 'application/json' },
      })

      if (!response.ok) throw new Error('Failed to fetch audit logs')

      const data = await response.json()
      setLogs(Array.isArray(data) ? data : [])
    } catch (error) {
      console.error(error)
      toast.error('Failed to load audit logs')
    } finally {
      setIsLoading(false)
    }
  }

  const getActionBadge = (action: string) => {
    const colors: Record<string, string> = {
      create: 'bg-green-500',
      update: 'bg-blue-500',
      delete: 'bg-red-500',
      login: 'bg-purple-500',
      register: 'bg-yellow-500',
    }
    return <Badge className={colors[action] || 'bg-gray-500'}>{action}</Badge>
  }

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-3xl font-bold">Audit Logs</h1>
          <p className="text-muted-foreground">Track changes and user activity</p>
        </div>
        <Button onClick={fetchLogs} variant="outline">
          <RefreshCw className="mr-2 h-4 w-4" />
          Refresh
        </Button>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Activity Log</CardTitle>
          <CardDescription>Recent system and user activities</CardDescription>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="flex justify-center p-6">Loading...</div>
          ) : logs.length === 0 ? (
            <Alert>
              <AlertTitle>No audit logs</AlertTitle>
              <AlertDescription>No activity recorded yet.</AlertDescription>
            </Alert>
          ) : (
            <div className="overflow-x-auto">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Time</TableHead>
                    <TableHead>User</TableHead>
                    <TableHead>Action</TableHead>
                    <TableHead>Entity</TableHead>
                    <TableHead>IP Address</TableHead>
                    <TableHead>Changes</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {logs.map((log) => (
                    <TableRow key={log.id}>
                      <TableCell className="whitespace-nowrap">
                        {new Date(log.created_at).toLocaleString()}
                      </TableCell>
                      <TableCell>User #{log.user_id}</TableCell>
                      <TableCell>{getActionBadge(log.action)}</TableCell>
                      <TableCell>
                        <span className="text-muted-foreground">{log.entity_type}</span>
                        <span className="ml-1">#{log.entity_id}</span>
                      </TableCell>
                      <TableCell className="font-mono text-sm">{log.ip_address}</TableCell>
                      <TableCell className="max-w-xs">
                        <pre className="text-xs text-muted-foreground overflow-hidden">
                          {JSON.stringify(log.changes, null, 0).substring(0, 50)}
                          {JSON.stringify(log.changes).length > 50 ? '...' : ''}
                        </pre>
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
