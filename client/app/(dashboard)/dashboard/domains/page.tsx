'use client'

import { useState, useEffect } from 'react'
import { Plus, Edit, Trash2, Server, Globe } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog'
import { Form, FormControl, FormDescription, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { Table, TableBody, TableCaption, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Switch } from '@/components/ui/switch'
import { useForm } from 'react-hook-form'
import { Toaster } from '@/components/ui/toaster'
import { toast } from 'sonner'
import { Domain, BackendServer, DomainWithBackends } from '@/types/domains'

export default function DomainsPage() {
  const [domains, setDomains] = useState<Domain[]>([])
  const [domainWithBackends, setDomainWithBackends] = useState<DomainWithBackends[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [openDialog, setOpenDialog] = useState(false)
  const [currentDomain, setCurrentDomain] = useState<Domain | null>(null)

  const form = useForm({
    defaultValues: {
      name: '',
      target_url: '',
      ssl_enabled: true,
      health_check_enabled: false,
      health_check_interval: 60,
    },
  })

  useEffect(() => {
    fetchDomains()
  }, [])

  const fetchDomains = async () => {
    setIsLoading(true)
    try {
      const response = await fetch('/api/domains', {
        headers: {
          'Content-Type': 'application/json',
        },
      })
      
      if (!response.ok) throw new Error('Failed to fetch domains')
      
      const data = await response.json()
      
      // Extract domains from the response structure
      if (Array.isArray(data)) {
        const extractedDomains: Domain[] = data.map(item => item.domain)
        setDomains(extractedDomains)
        setDomainWithBackends(data)
      } else {
        setDomains([])
        setDomainWithBackends([])
      }
    } catch (error) {
      toast.error('Failed to load domains. Please try again.')
    } finally {
      setIsLoading(false)
    }
  }

  const onSubmit = async (values: any) => {
    try {
      const method = currentDomain ? 'PUT' : 'POST'
      const url = currentDomain 
        ? `/api/domains/${currentDomain.id}` 
        : '/api/domains'
      
      const payload = {
        domain: values,
        backend_servers: [] // We'll handle backend servers on a separate page
      }

      const response = await fetch(url, {
        method,
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(payload),
      })

      if (!response.ok) throw new Error('Failed to save domain')
      
      toast.success(currentDomain ? 'Domain updated successfully' : 'Domain created successfully')
      
      setOpenDialog(false)
      form.reset()
      fetchDomains()
    } catch (error) {
      toast.error('Failed to save domain. Please try again.')
    }
  }

  const handleDelete = async (id: number) => {
    if (!confirm('Are you sure you want to delete this domain?')) return

    try {
      const response = await fetch(`/api/domains/${id}`, {
        method: 'DELETE',
        headers: {
          'Content-Type': 'application/json',
        },
      })

      if (!response.ok) throw new Error('Failed to delete domain')
      
      toast.success('Domain deleted successfully')
      
      fetchDomains()
    } catch (error) {
      toast.error('Failed to delete domain. Please try again.')
    }
  }

  const handleEdit = (domain: Domain) => {
    setCurrentDomain(domain)
    form.reset({
      name: domain.name,
      target_url: domain.target_url,
      ssl_enabled: domain.ssl_enabled,
      health_check_enabled: domain.health_check_enabled,
      health_check_interval: domain.health_check_interval,
    })
    setOpenDialog(true)
  }

  const handleAdd = () => {
    setCurrentDomain(null)
    form.reset({
      name: '',
      target_url: '',
      ssl_enabled: true,
      health_check_enabled: false,
      health_check_interval: 60,
    })
    setOpenDialog(true)
  }

  const getStatusBadge = (domain: Domain) => {
    const domainWithBackend = domainWithBackends.find(item => item.domain.id === domain.id)
    const hasActiveBackends = domainWithBackend?.backend_servers.some(server => server.is_active) || false
    
    if (!hasActiveBackends) {
      return <Badge className="bg-gray-500">Inactive</Badge>
    }
    
    return <Badge className="bg-primary">Active</Badge>
  }

  const countBackends = (domainId: number) => {
    const domainWithBackend = domainWithBackends.find(item => item.domain.id === domainId)
    return domainWithBackend?.backend_servers.length || 0
  }

  const getTargetType = (targetUrl: string) => {
    if (targetUrl.startsWith('tcp://')) {
      return <Badge variant="outline">TCP</Badge>
    } else if (targetUrl.startsWith('https://')) {
      return <Badge variant="outline">HTTPS</Badge>
    } else {
      return <Badge variant="outline">HTTP</Badge>
    }
  }

  return (
    <div className="container mx-auto p-6">
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-3xl font-bold">Domains</h1>
        <Button onClick={handleAdd}>
          <Plus className="mr-2 h-4 w-4" />
          Add Domain
        </Button>
      </div>

      <Tabs defaultValue="all">
        <TabsList className="mb-4">
          <TabsTrigger value="all">All Domains</TabsTrigger>
          <TabsTrigger value="http">HTTP/HTTPS</TabsTrigger>
          <TabsTrigger value="tcp">TCP</TabsTrigger>
        </TabsList>

        <TabsContent value="all">
          <Card>
            <CardHeader>
              <CardTitle>All Domains</CardTitle>
              <CardDescription>Manage all your registered domains</CardDescription>
            </CardHeader>
            <CardContent>
              {isLoading ? (
                <div className="flex justify-center p-6">Loading domains...</div>
              ) : domains.length === 0 ? (
                <Alert>
                  <AlertTitle>No domains found</AlertTitle>
                  <AlertDescription>
                    You have not added any domains yet. Click the "Add Domain" button to get started.
                  </AlertDescription>
                </Alert>
              ) : (
                <Table>
                  <TableCaption>A list of your domains.</TableCaption>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Domain Name</TableHead>
                      <TableHead>Target URL</TableHead>
                      <TableHead>Type</TableHead>
                      <TableHead>Status</TableHead>
                      <TableHead>SSL</TableHead>
                      <TableHead>Health Check</TableHead>
                      <TableHead>Backends</TableHead>
                      <TableHead>Created</TableHead>
                      <TableHead className="text-right">Actions</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {domains.map((domain) => (
                      <TableRow key={domain.id}>
                        <TableCell className="font-medium">{domain.name}</TableCell>
                        <TableCell>{domain.target_url}</TableCell>
                        <TableCell>{getTargetType(domain.target_url)}</TableCell>
                        <TableCell>{getStatusBadge(domain)}</TableCell>
                        <TableCell>{domain.ssl_enabled ? <Badge>Enabled</Badge> : <Badge variant="outline">Disabled</Badge>}</TableCell>
                        <TableCell>{domain.health_check_enabled ? 
                          <Badge>Every {domain.health_check_interval}s</Badge> : 
                          <Badge variant="outline">Disabled</Badge>}</TableCell>
                        <TableCell>{countBackends(domain.id)}</TableCell>
                        <TableCell>{new Date(domain.created_at).toLocaleDateString()}</TableCell>
                        <TableCell className="text-right">
                          <Button variant="ghost" size="sm" onClick={() => handleEdit(domain)}>
                            <Edit className="h-4 w-4" />
                          </Button>
                          <Button variant="ghost" size="sm" onClick={() => handleDelete(domain.id)}>
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
        </TabsContent>

        <TabsContent value="http">
          <Card>
            <CardHeader>
              <CardTitle>HTTP/HTTPS Domains</CardTitle>
              <CardDescription>Domains using HTTP or HTTPS protocol</CardDescription>
            </CardHeader>
            <CardContent>
              <Table>
                <TableBody>
                  {domains
                    .filter(domain => !domain.target_url.startsWith('tcp://'))
                    .map((domain) => (
                      <TableRow key={domain.id}>
                        <TableCell className="font-medium">{domain.name}</TableCell>
                        <TableCell>{domain.target_url}</TableCell>
                        <TableCell>{getTargetType(domain.target_url)}</TableCell>
                        <TableCell>{getStatusBadge(domain)}</TableCell>
                        <TableCell>{domain.ssl_enabled ? <Badge>Enabled</Badge> : <Badge variant="outline">Disabled</Badge>}</TableCell>
                        <TableCell>{domain.health_check_enabled ? 
                          <Badge>Every {domain.health_check_interval}s</Badge> : 
                          <Badge variant="outline">Disabled</Badge>}</TableCell>
                        <TableCell>{countBackends(domain.id)}</TableCell>
                        <TableCell>{new Date(domain.created_at).toLocaleDateString()}</TableCell>
                        <TableCell className="text-right">
                          <Button variant="ghost" size="sm" onClick={() => handleEdit(domain)}>
                            <Edit className="h-4 w-4" />
                          </Button>
                          <Button variant="ghost" size="sm" onClick={() => handleDelete(domain.id)}>
                            <Trash2 className="h-4 w-4" />
                          </Button>
                        </TableCell>
                      </TableRow>
                    ))}
                </TableBody>
              </Table>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="tcp">
          <Card>
            <CardHeader>
              <CardTitle>TCP Domains</CardTitle>
              <CardDescription>Domains using TCP protocol</CardDescription>
            </CardHeader>
            <CardContent>
              <Table>
                <TableBody>
                  {domains
                    .filter(domain => domain.target_url.startsWith('tcp://'))
                    .map((domain) => (
                      <TableRow key={domain.id}>
                        <TableCell className="font-medium">{domain.name}</TableCell>
                        <TableCell>{domain.target_url}</TableCell>
                        <TableCell>{getTargetType(domain.target_url)}</TableCell>
                        <TableCell>{getStatusBadge(domain)}</TableCell>
                        <TableCell>{domain.ssl_enabled ? <Badge>Enabled</Badge> : <Badge variant="outline">Disabled</Badge>}</TableCell>
                        <TableCell>{domain.health_check_enabled ? 
                          <Badge>Every {domain.health_check_interval}s</Badge> : 
                          <Badge variant="outline">Disabled</Badge>}</TableCell>
                        <TableCell>{countBackends(domain.id)}</TableCell>
                        <TableCell>{new Date(domain.created_at).toLocaleDateString()}</TableCell>
                        <TableCell className="text-right">
                          <Button variant="ghost" size="sm" onClick={() => handleEdit(domain)}>
                            <Edit className="h-4 w-4" />
                          </Button>
                          <Button variant="ghost" size="sm" onClick={() => handleDelete(domain.id)}>
                            <Trash2 className="h-4 w-4" />
                          </Button>
                        </TableCell>
                      </TableRow>
                    ))}
                </TableBody>
              </Table>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>

      <Dialog open={openDialog} onOpenChange={setOpenDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{currentDomain ? 'Edit Domain' : 'Add New Domain'}</DialogTitle>
            <DialogDescription>
              {currentDomain 
                ? 'Update the domain details below.' 
                : 'Enter the details for the new domain.'}
            </DialogDescription>
          </DialogHeader>
          <Form {...form}>
            <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
              <FormField
                control={form.control}
                name="name"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Domain Name</FormLabel>
                    <FormControl>
                      <Input placeholder="example.com" {...field} />
                    </FormControl>
                    <FormDescription>
                      Enter the fully qualified domain name.
                    </FormDescription>
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
                      <Input placeholder="https://backend.example.com" {...field} />
                    </FormControl>
                    <FormDescription>
                      Enter the target URL (e.g., https://backend.example.com or tcp://myserver.example.com)
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />
              
              <FormField
                control={form.control}
                name="ssl_enabled"
                render={({ field }) => (
                  <FormItem className="flex flex-row items-center justify-between rounded-lg border p-4">
                    <div className="space-y-0.5">
                      <FormLabel className="text-base">SSL Certificate</FormLabel>
                      <FormDescription>
                        Enable SSL/TLS for this domain
                      </FormDescription>
                    </div>
                    <FormControl>
                      <Switch
                        checked={field.value}
                        onCheckedChange={field.onChange}
                      />
                    </FormControl>
                  </FormItem>
                )}
              />
              
              <FormField
                control={form.control}
                name="health_check_enabled"
                render={({ field }) => (
                  <FormItem className="flex flex-row items-center justify-between rounded-lg border p-4">
                    <div className="space-y-0.5">
                      <FormLabel className="text-base">Health Check</FormLabel>
                      <FormDescription>
                        Enable health checking for backend servers
                      </FormDescription>
                    </div>
                    <FormControl>
                      <Switch
                        checked={field.value}
                        onCheckedChange={field.onChange}
                      />
                    </FormControl>
                  </FormItem>
                )}
              />
              
              {form.watch('health_check_enabled') && (
                <FormField
                  control={form.control}
                  name="health_check_interval"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Health Check Interval (seconds)</FormLabel>
                      <FormControl>
                        <Input type="number" {...field} />
                      </FormControl>
                      <FormDescription>
                        How often to check backend server health (in seconds)
                      </FormDescription>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              )}
              
              <DialogFooter>
                <Button type="submit">
                  {currentDomain ? 'Update Domain' : 'Add Domain'}
                </Button>
              </DialogFooter>
            </form>
          </Form>
        </DialogContent>
      </Dialog>
    </div>
  )
} 