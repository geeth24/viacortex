'use client'

import { useState, useEffect } from 'react'
import { Plus, Trash2, Edit } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Form, FormControl, FormDescription, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Switch } from '@/components/ui/switch'
import { toast } from 'sonner'
import { useForm } from 'react-hook-form'
import { Domain, Redirect } from '@/types/domains'

interface RedirectFormValues {
  source_path: string
  target_url: string
  redirect_type: number
  preserve_query: boolean
  enabled: boolean
}

export default function RedirectsPage() {
  const [domains, setDomains] = useState<Domain[]>([])
  const [selectedDomainId, setSelectedDomainId] = useState<string>('')
  const [redirects, setRedirects] = useState<Redirect[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [openDialog, setOpenDialog] = useState(false)
  const [currentRedirect, setCurrentRedirect] = useState<Redirect | null>(null)

  const form = useForm<RedirectFormValues>({
    defaultValues: {
      source_path: '',
      target_url: '',
      redirect_type: 301,
      preserve_query: true,
      enabled: true,
    },
  })

  useEffect(() => {
    fetchDomains()
  }, [])

  useEffect(() => {
    if (selectedDomainId) {
      fetchRedirects(selectedDomainId)
    }
  }, [selectedDomainId])

  const fetchDomains = async () => {
    try {
      const response = await fetch('/api/domains', {
        headers: { 'Content-Type': 'application/json' },
      })

      if (!response.ok) throw new Error('Failed to fetch domains')

      const data = await response.json()
      const domainList = Array.isArray(data) ? data : []
      setDomains(domainList)

      if (domainList.length > 0) {
        setSelectedDomainId(String(domainList[0].id))
      }
    } catch (error) {
      console.error(error)
      toast.error('Failed to load domains')
    } finally {
      setIsLoading(false)
    }
  }

  const fetchRedirects = async (domainId: string) => {
    setIsLoading(true)
    try {
      const response = await fetch(`/api/domains/${domainId}/redirects`, {
        headers: { 'Content-Type': 'application/json' },
      })

      if (!response.ok) throw new Error('Failed to fetch redirects')

      const data = await response.json()
      setRedirects(Array.isArray(data) ? data : [])
    } catch (error) {
      console.error(error)
      toast.error('Failed to load redirects')
    } finally {
      setIsLoading(false)
    }
  }

  const onSubmit = async (values: RedirectFormValues) => {
    if (!selectedDomainId) return

    try {
      const method = currentRedirect ? 'PUT' : 'POST'
      const url = currentRedirect
        ? `/api/domains/${selectedDomainId}/redirects/${currentRedirect.id}`
        : `/api/domains/${selectedDomainId}/redirects`

      const response = await fetch(url, {
        method,
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(values),
      })

      if (!response.ok) throw new Error('Failed to save redirect')

      toast.success(currentRedirect ? 'Redirect updated' : 'Redirect created')
      setOpenDialog(false)
      form.reset()
      setCurrentRedirect(null)
      fetchRedirects(selectedDomainId)
    } catch (error) {
      console.error(error)
      toast.error('Failed to save redirect')
    }
  }

  const handleDelete = async (redirectId: number) => {
    if (!selectedDomainId) return
    if (!confirm('Delete this redirect?')) return

    try {
      const response = await fetch(`/api/domains/${selectedDomainId}/redirects/${redirectId}`, {
        method: 'DELETE',
        headers: { 'Content-Type': 'application/json' },
      })

      if (!response.ok) throw new Error('Failed to delete redirect')

      toast.success('Redirect deleted')
      fetchRedirects(selectedDomainId)
    } catch (error) {
      console.error(error)
      toast.error('Failed to delete redirect')
    }
  }

  const handleEdit = (redirect: Redirect) => {
    setCurrentRedirect(redirect)
    form.reset({
      source_path: redirect.source_path,
      target_url: redirect.target_url,
      redirect_type: redirect.redirect_type,
      preserve_query: redirect.preserve_query,
      enabled: redirect.enabled,
    })
    setOpenDialog(true)
  }

  const handleAdd = () => {
    setCurrentRedirect(null)
    form.reset({
      source_path: '',
      target_url: '',
      redirect_type: 301,
      preserve_query: true,
      enabled: true,
    })
    setOpenDialog(true)
  }

  const getRedirectTypeBadge = (type: number) => {
    switch (type) {
      case 301:
        return <Badge variant="default">301 Permanent</Badge>
      case 302:
        return <Badge variant="secondary">302 Found</Badge>
      case 307:
        return <Badge variant="outline">307 Temporary</Badge>
      case 308:
        return <Badge>308 Permanent</Badge>
      default:
        return <Badge variant="outline">{type}</Badge>
    }
  }

  const selectedDomain = domains.find(d => String(d.id) === selectedDomainId)

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-3xl font-bold">Redirects</h1>
          <p className="text-muted-foreground">Configure URL redirects for your domains</p>
        </div>
        <Button onClick={handleAdd} disabled={!selectedDomainId}>
          <Plus className="mr-2 h-4 w-4" />
          Add Redirect
        </Button>
      </div>

      <div className="flex items-center gap-4">
        <span className="text-sm font-medium">Domain:</span>
        <Select value={selectedDomainId} onValueChange={setSelectedDomainId}>
          <SelectTrigger className="w-[250px]">
            <SelectValue placeholder="Select a domain" />
          </SelectTrigger>
          <SelectContent>
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
          <CardTitle>URL Redirects</CardTitle>
          <CardDescription>
            {selectedDomain
              ? `Manage redirects for ${selectedDomain.name}`
              : 'Select a domain to manage redirects'}
          </CardDescription>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="flex justify-center p-6">Loading...</div>
          ) : !selectedDomainId ? (
            <Alert>
              <AlertTitle>No domain selected</AlertTitle>
              <AlertDescription>Select a domain to manage redirects.</AlertDescription>
            </Alert>
          ) : redirects.length === 0 ? (
            <Alert>
              <AlertTitle>No redirects</AlertTitle>
              <AlertDescription>Add redirects to route traffic to different URLs.</AlertDescription>
            </Alert>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Source Path</TableHead>
                  <TableHead>Target URL</TableHead>
                  <TableHead>Type</TableHead>
                  <TableHead>Query</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead className="text-right">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {redirects.map((redirect) => (
                  <TableRow key={redirect.id}>
                    <TableCell className="font-mono">{redirect.source_path}</TableCell>
                    <TableCell className="max-w-xs truncate">{redirect.target_url}</TableCell>
                    <TableCell>{getRedirectTypeBadge(redirect.redirect_type)}</TableCell>
                    <TableCell>
                      {redirect.preserve_query ? (
                        <Badge variant="outline">Preserve</Badge>
                      ) : (
                        <Badge variant="secondary">Strip</Badge>
                      )}
                    </TableCell>
                    <TableCell>
                      {redirect.enabled ? (
                        <Badge className="bg-green-500">Active</Badge>
                      ) : (
                        <Badge variant="secondary">Disabled</Badge>
                      )}
                    </TableCell>
                    <TableCell className="text-right space-x-1">
                      <Button variant="ghost" size="sm" onClick={() => handleEdit(redirect)}>
                        <Edit className="h-4 w-4" />
                      </Button>
                      <Button variant="ghost" size="sm" onClick={() => handleDelete(redirect.id)}>
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

      <Dialog open={openDialog} onOpenChange={setOpenDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{currentRedirect ? 'Edit Redirect' : 'Add Redirect'}</DialogTitle>
            <DialogDescription>Configure the URL redirect rule.</DialogDescription>
          </DialogHeader>
          <Form {...form}>
            <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
              <FormField
                control={form.control}
                name="source_path"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Source Path</FormLabel>
                    <FormControl>
                      <Input placeholder="/old-page" {...field} />
                    </FormControl>
                    <FormDescription>Path to redirect from</FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="target_url"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Target URL</FormLabel>
                    <FormControl>
                      <Input placeholder="https://example.com/new-page" {...field} />
                    </FormControl>
                    <FormDescription>URL to redirect to</FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="redirect_type"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Redirect Type</FormLabel>
                    <Select
                      onValueChange={(v) => field.onChange(parseInt(v))}
                      defaultValue={String(field.value)}
                    >
                      <FormControl>
                        <SelectTrigger>
                          <SelectValue />
                        </SelectTrigger>
                      </FormControl>
                      <SelectContent>
                        <SelectItem value="301">301 - Permanent</SelectItem>
                        <SelectItem value="302">302 - Found</SelectItem>
                        <SelectItem value="307">307 - Temporary</SelectItem>
                        <SelectItem value="308">308 - Permanent (preserves method)</SelectItem>
                      </SelectContent>
                    </Select>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="preserve_query"
                render={({ field }) => (
                  <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3">
                    <div className="space-y-0.5">
                      <FormLabel>Preserve Query String</FormLabel>
                      <FormDescription>Keep query parameters in redirect</FormDescription>
                    </div>
                    <FormControl>
                      <Switch checked={field.value} onCheckedChange={field.onChange} />
                    </FormControl>
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="enabled"
                render={({ field }) => (
                  <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3">
                    <div className="space-y-0.5">
                      <FormLabel>Enabled</FormLabel>
                      <FormDescription>Activate this redirect</FormDescription>
                    </div>
                    <FormControl>
                      <Switch checked={field.value} onCheckedChange={field.onChange} />
                    </FormControl>
                  </FormItem>
                )}
              />

              <DialogFooter>
                <Button type="button" variant="outline" onClick={() => setOpenDialog(false)}>
                  Cancel
                </Button>
                <Button type="submit">{currentRedirect ? 'Update' : 'Create'}</Button>
              </DialogFooter>
            </form>
          </Form>
        </DialogContent>
      </Dialog>
    </div>
  )
}
