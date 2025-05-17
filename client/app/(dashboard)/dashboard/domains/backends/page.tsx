'use client'

import { useState, useEffect } from 'react'
import { Plus, Edit, Trash2, ExternalLink, Server, Heart, AlertTriangle } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Form, FormControl, FormDescription, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Table, TableBody, TableCaption, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Switch } from '@/components/ui/switch'
import { Progress } from '@/components/ui/progress'
import { useForm } from 'react-hook-form'
import { toast } from 'sonner'
import { Domain, BackendServer, DomainWithBackends } from '@/types/domains'

export default function BackendServersPage() {
  const [domains, setDomains] = useState<Domain[]>([])
  const [selectedDomain, setSelectedDomain] = useState<number | null>(null)
  const [backendServers, setBackendServers] = useState<BackendServer[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [openDialog, setOpenDialog] = useState(false)
  const [currentServer, setCurrentServer] = useState<BackendServer | null>(null)

  const form = useForm({
    defaultValues: {
      name: '',
      scheme: 'http',
      ip: '',
      port: 80,
      weight: 1,
      is_active: true,
    },
  })

  useEffect(() => {
    fetchDomains()
  }, [])

  useEffect(() => {
    if (selectedDomain) {
      fetchBackendServers(selectedDomain)
    }
  }, [selectedDomain])

  const fetchDomains = async () => {
    try {
      const response = await fetch('/api/domains', {
        headers: {
          'Content-Type': 'application/json',
        },
      })
      
      if (!response.ok) throw new Error('Failed to fetch domains')
      
      const data = await response.json()
      console.log('Domains response:', data)
      
      // Extract domains from the response structure
      if (Array.isArray(data)) {
        const extractedDomains: Domain[] = data.map(item => item.domain)
        console.log('Extracted domains:', extractedDomains)
        setDomains(extractedDomains)
        
        // Select the first domain by default if available
        if (extractedDomains.length > 0) {
          setSelectedDomain(extractedDomains[0].id)
        }
      } else {
        setDomains([])
      }
    } catch (error) {
      console.error('Domains error:', error)
      toast.error('Failed to load domains. Please try again.')
    } finally {
      setIsLoading(false)
    }
  }

  const fetchBackendServers = async (domainId: number) => {
    setIsLoading(true)
    try {
      console.log(`Fetching backend servers for domain ID: ${domainId}`)
      const response = await fetch(`/api/domains/${domainId}/backends`, {
        headers: {
          'Content-Type': 'application/json',
        },
      })
      
      if (!response.ok) {
        console.error(`Error response: ${response.status} ${response.statusText}`)
        throw new Error('Failed to fetch backend servers')
      }
      
      const data = await response.json()
      console.log('Backend servers response:', data)
      
      // Handle both array and object with backend_servers property
      if (Array.isArray(data)) {
        console.log(`Setting ${data.length} backend servers from array`)
        setBackendServers(data)
      } else {
        console.log(`Setting ${(data.backend_servers || []).length} backend servers from object`)
        setBackendServers(data.backend_servers || [])
      }
    } catch (error) {
      console.error('Backend servers error:', error)
      toast.error('Failed to load backend servers. Please try again.')
    } finally {
      setIsLoading(false)
    }
  }

  const onSubmit = async (values: any) => {
    if (!selectedDomain) return

    try {
      // Make sure domain_id is properly set
      const submissionData = {
        ...values,
        domain_id: selectedDomain
      }
      
      const method = currentServer ? 'PUT' : 'POST'
      const url = currentServer 
        ? `/api/domains/${selectedDomain}/backends/${currentServer.id}` 
        : `/api/domains/${selectedDomain}/backends`
      
      console.log('Submitting server data:', submissionData, 'to', url)
      
      const response = await fetch(url, {
        method,
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(submissionData),
      })

      if (!response.ok) {
        const errorData = await response.json()
        console.error('Server submission error:', errorData)
        throw new Error('Failed to save backend server')
      }
      
      toast.success(currentServer ? 'Backend server updated successfully' : 'Backend server added successfully')
      
      setOpenDialog(false)
      form.reset()
      fetchBackendServers(selectedDomain)
    } catch (error) {
      console.error('Submit error:', error)
      toast.error('Failed to save backend server. Please try again.')
    }
  }

  const handleDelete = async (serverId: number) => {
    if (!selectedDomain) return
    if (!confirm('Are you sure you want to delete this backend server?')) return

    try {
      const response = await fetch(`/api/domains/${selectedDomain}/backends/${serverId}`, {
        method: 'DELETE',
        headers: {
          'Content-Type': 'application/json',
        },
      })

      if (!response.ok) throw new Error('Failed to delete backend server')
      
      toast.success('Backend server deleted successfully')
      
      fetchBackendServers(selectedDomain)
    } catch (error) {
      toast.error('Failed to delete backend server. Please try again.')
    }
  }

  const handleEdit = (server: BackendServer) => {
    setCurrentServer(server)
    form.reset({
      name: server.name || '',
      scheme: server.scheme,
      ip: server.ip,
      port: server.port,
      weight: server.weight,
      is_active: server.is_active,
    })
    setOpenDialog(true)
  }

  const handleAdd = () => {
    setCurrentServer(null)
    form.reset({
      name: '',
      scheme: 'http',
      ip: '',
      port: 80,
      weight: 1,
      is_active: true,
    })
    setOpenDialog(true)
  }

  const getHealthBadge = (healthStatus: string | undefined) => {
    switch (healthStatus) {
      case 'healthy':
        return <Badge className="bg-green-500"><Heart className="mr-1 h-3 w-3" /> Healthy</Badge>
      case 'unhealthy':
        return <Badge className="bg-red-500"><AlertTriangle className="mr-1 h-3 w-3" /> Unhealthy</Badge>
      default:
        return <Badge className="bg-gray-500">Unknown</Badge>
    }
  }

  const getStatusBadge = (isActive: boolean) => {
    return isActive 
      ? <Badge className="bg-green-500">Active</Badge>
      : <Badge className="bg-gray-500">Inactive</Badge>
  }

  const getSchemeBadge = (scheme: string) => {
    switch (scheme) {
      case 'https':
        return <Badge variant="outline" className="bg-green-100">HTTPS</Badge>
      case 'tcp':
        return <Badge variant="outline" className="bg-blue-100">TCP</Badge>
      default:
        return <Badge variant="outline" className="bg-gray-100">HTTP</Badge>
    }
  }

  return (
    <div className="container mx-auto p-6">
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-3xl font-bold">Backend Servers</h1>
        <Button onClick={handleAdd} disabled={!selectedDomain}>
          <Plus className="mr-2 h-4 w-4" />
          Add Server
        </Button>
      </div>

      <div className="mb-6">
        <Select 
          value={selectedDomain?.toString() || ''} 
          onValueChange={(value) => setSelectedDomain(parseInt(value))}
        >
          <SelectTrigger className="w-[250px]">
            <SelectValue placeholder="Select a domain" />
          </SelectTrigger>
          <SelectContent>
            {domains.map((domain) => (
              <SelectItem key={domain.id} value={domain.id.toString()}>{domain.name}</SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Backend Servers</CardTitle>
          <CardDescription>
            {selectedDomain 
              ? `Manage backend servers for ${domains.find(d => d.id === selectedDomain)?.name}` 
              : 'Please select a domain to manage its backend servers'}
          </CardDescription>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="flex justify-center p-6">Loading backends...</div>
          ) : !selectedDomain ? (
            <Alert>
              <AlertTitle>No domain selected</AlertTitle>
              <AlertDescription>
                Please select a domain from the dropdown above to manage its backend servers.
              </AlertDescription>
            </Alert>
          ) : backendServers.length === 0 ? (
            <Alert>
              <AlertTitle>No backend servers found</AlertTitle>
              <AlertDescription>
                This domain doesn't have any backend servers yet. Click the "Add Server" button to get started.
              </AlertDescription>
            </Alert>
          ) : (
            <Table>
              <TableCaption>A list of backend servers for this domain.</TableCaption>
              <TableHeader>
                <TableRow>
                  <TableHead>Name</TableHead>
                  <TableHead>Scheme</TableHead>
                  <TableHead>IP Address</TableHead>
                  <TableHead>Port</TableHead>
                  <TableHead>Health</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Weight</TableHead>
                  <TableHead className="text-right">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {backendServers.map((server) => (
                  <TableRow key={server.id}>
                    <TableCell className="font-medium">{server.name || `Server ${server.id}`}</TableCell>
                    <TableCell>{getSchemeBadge(server.scheme)}</TableCell>
                    <TableCell>{server.ip}</TableCell>
                    <TableCell>{server.port}</TableCell>
                    <TableCell>{getHealthBadge(server.health_status)}</TableCell>
                    <TableCell>{getStatusBadge(server.is_active)}</TableCell>
                    <TableCell>{server.weight}</TableCell>
                    <TableCell className="text-right">
                      <Button variant="ghost" size="sm" onClick={() => handleEdit(server)}>
                        <Edit className="h-4 w-4" />
                      </Button>
                      <Button variant="ghost" size="sm" onClick={() => handleDelete(server.id)}>
                        <Trash2 className="h-4 w-4" />
                      </Button>
                      <Button variant="ghost" size="sm" asChild>
                        <a href={`${server.scheme}://${server.ip}:${server.port}`} target="_blank" rel="noopener noreferrer">
                          <ExternalLink className="h-4 w-4" />
                        </a>
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
            <DialogTitle>{currentServer ? 'Edit Backend Server' : 'Add New Backend Server'}</DialogTitle>
            <DialogDescription>
              {currentServer 
                ? 'Update the backend server details below.' 
                : 'Enter the details for the new backend server.'}
            </DialogDescription>
          </DialogHeader>
          <Form {...form}>
            <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
              <FormField
                control={form.control}
                name="name"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Server Name</FormLabel>
                    <FormControl>
                      <Input placeholder="API Server 1" {...field} />
                    </FormControl>
                    <FormDescription>
                      A descriptive name for this backend server.
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />
              
              <FormField
                control={form.control}
                name="scheme"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Scheme</FormLabel>
                    <Select 
                      onValueChange={field.onChange} 
                      defaultValue={field.value}
                    >
                      <FormControl>
                        <SelectTrigger>
                          <SelectValue placeholder="Select a scheme" />
                        </SelectTrigger>
                      </FormControl>
                      <SelectContent>
                        <SelectItem value="http">HTTP</SelectItem>
                        <SelectItem value="https">HTTPS</SelectItem>
                        <SelectItem value="tcp">TCP</SelectItem>
                      </SelectContent>
                    </Select>
                    <FormDescription>
                      The protocol to use for this backend server
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />
              
              <div className="grid grid-cols-2 gap-4">
                <FormField
                  control={form.control}
                  name="ip"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>IP Address</FormLabel>
                      <FormControl>
                        <Input placeholder="192.168.1.1" {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                
                <FormField
                  control={form.control}
                  name="port"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Port</FormLabel>
                      <FormControl>
                        <Input type="number" {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              </div>
              
              <FormField
                control={form.control}
                name="weight"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Weight</FormLabel>
                    <FormControl>
                      <Input type="number" {...field} />
                    </FormControl>
                    <FormDescription>
                      Relative weight for load balancing (higher values receive more traffic)
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />
              
              <FormField
                control={form.control}
                name="is_active"
                render={({ field }) => (
                  <FormItem className="flex flex-row items-center justify-between rounded-lg border p-4">
                    <div className="space-y-0.5">
                      <FormLabel className="text-base">Active</FormLabel>
                      <FormDescription>
                        Enable this backend server to receive traffic
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
              
              <DialogFooter>
                <Button type="submit">
                  {currentServer ? 'Update Server' : 'Add Server'}
                </Button>
              </DialogFooter>
            </form>
          </Form>
        </DialogContent>
      </Dialog>
    </div>
  )
} 