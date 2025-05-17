'use client'

import { useState, useEffect } from 'react'
import { Plus, Edit, Trash2, Clock, AlertCircle } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Form, FormControl, FormDescription, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { Table, TableBody, TableCaption, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Slider } from '@/components/ui/slider'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { useForm } from 'react-hook-form'
import { toast } from 'sonner'
interface Domain {
  id: string
  name: string
}

interface RateLimit {
  id: string
  domainId: string
  name: string
  path: string
  requests: number
  duration: number
  action: 'block' | 'throttle' | 'log'
  createdAt: string
}

export default function RateLimitsPage() {
  const [domains, setDomains] = useState<Domain[]>([])
  const [selectedDomain, setSelectedDomain] = useState<string | null>(null)
  const [rateLimits, setRateLimits] = useState<RateLimit[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [openDialog, setOpenDialog] = useState(false)
  const [currentLimit, setCurrentLimit] = useState<RateLimit | null>(null)

  const form = useForm({
    defaultValues: {
      name: '',
      path: '/*',
      requests: 100,
      duration: 60,
      action: 'block',
    },
  })

  useEffect(() => {
    fetchDomains()
  }, [])

  useEffect(() => {
    if (selectedDomain) {
      fetchRateLimits(selectedDomain)
    }
  }, [selectedDomain])

  const fetchDomains = async () => {
    try {
      const response = await fetch('/api/domains', {
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
      })
      
      if (!response.ok) throw new Error('Failed to fetch domains')
      
      const data = await response.json()
      setDomains(data.domains || [])
      
      // Select the first domain by default if available
      if (data.domains && data.domains.length > 0) {
        setSelectedDomain(data.domains[0].id)
      }
    } catch (error) {
      toast.error('Failed to load domains. Please try again.')
      
    } finally {
      setIsLoading(false)
    }
  }

  const fetchRateLimits = async (domainId: string) => {
    setIsLoading(true)
    try {
      const response = await fetch(`/api/domains/${domainId}/rate-limits`, {
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
      })
      
      if (!response.ok) throw new Error('Failed to fetch rate limits')
      
      const data = await response.json()
      setRateLimits(data.rateLimits || [])
    } catch (error) {
      toast.error('Failed to load rate limits. Please try again.')
        
    } finally {
      setIsLoading(false)
    }
  }

  const onSubmit = async (values: any) => {
    if (!selectedDomain) return

    try {
      const method = currentLimit ? 'PUT' : 'POST'
      const url = currentLimit 
        ? `/api/domains/${selectedDomain}/rate-limits/${currentLimit.id}` 
        : `/api/domains/${selectedDomain}/rate-limits`
      
      const response = await fetch(url, {
        method,
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
        body: JSON.stringify(values),
      })

      if (!response.ok) throw new Error('Failed to save rate limit')
      
      toast.success(currentLimit ? 'Rate limit updated successfully' : 'Rate limit added successfully')
      
      setOpenDialog(false)
      form.reset()
      fetchRateLimits(selectedDomain)
    } catch (error) {
      toast.error('Failed to save rate limit. Please try again.')
    }
  }

  const handleDelete = async (limitId: string) => {
    if (!selectedDomain) return
    if (!confirm('Are you sure you want to delete this rate limit?')) return

    try {
      const response = await fetch(`/api/domains/${selectedDomain}/rate-limits/${limitId}`, {
        method: 'DELETE',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
      })

      if (!response.ok) throw new Error('Failed to delete rate limit')
      
      toast.success('Rate limit deleted successfully')
      
      fetchRateLimits(selectedDomain)
    } catch (error) {
      toast.error('Failed to delete rate limit. Please try again.')
    }
  }

  const handleEdit = (limit: RateLimit) => {
    setCurrentLimit(limit)
    form.reset({
      name: limit.name,
      path: limit.path,
      requests: limit.requests,
      duration: limit.duration,
      action: limit.action,
    })
    setOpenDialog(true)
  }

  const handleAdd = () => {
    setCurrentLimit(null)
    form.reset({
      name: '',
      path: '/*',
      requests: 100,
      duration: 60,
      action: 'block',
    })
    setOpenDialog(true)
  }

  const getActionBadge = (action: string) => {
    switch (action) {
      case 'block':
        return <Badge className="bg-red-500">Block</Badge>
      case 'throttle':
        return <Badge className="bg-yellow-500">Throttle</Badge>
      case 'log':
        return <Badge className="bg-blue-500">Log Only</Badge>
      default:
        return <Badge>{action}</Badge>
    }
  }

  return (
    <div className="container mx-auto p-6">
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-3xl font-bold">Rate Limits</h1>
        <Button onClick={handleAdd} disabled={!selectedDomain}>
          <Plus className="mr-2 h-4 w-4" />
          Add Rate Limit
        </Button>
      </div>

      <div className="mb-6">
        <Select 
          value={selectedDomain || ''} 
          onValueChange={(value) => setSelectedDomain(value)}
        >
          <SelectTrigger className="w-[250px]">
            <SelectValue placeholder="Select a domain" />
          </SelectTrigger>
          <SelectContent>
            {domains.map((domain) => (
              <SelectItem key={domain.id} value={domain.id}>{domain.name}</SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Rate Limits</CardTitle>
          <CardDescription>
            {selectedDomain 
              ? `Manage rate limiting for ${domains.find(d => d.id === selectedDomain)?.name}` 
              : 'Please select a domain to manage its rate limits'}
          </CardDescription>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="flex justify-center p-6">Loading rate limits...</div>
          ) : !selectedDomain ? (
            <Alert>
              <AlertTitle>No domain selected</AlertTitle>
              <AlertDescription>
                Please select a domain from the dropdown above to manage its rate limits.
              </AlertDescription>
            </Alert>
          ) : rateLimits.length === 0 ? (
            <Alert>
              <AlertTitle>No rate limits found</AlertTitle>
              <AlertDescription>
                This domain doesn't have any rate limits yet. Click the "Add Rate Limit" button to get started.
              </AlertDescription>
            </Alert>
          ) : (
            <Tabs defaultValue="all">
              <TabsList className="mb-4">
                <TabsTrigger value="all">All</TabsTrigger>
                <TabsTrigger value="block">Block</TabsTrigger>
                <TabsTrigger value="throttle">Throttle</TabsTrigger>
                <TabsTrigger value="log">Log Only</TabsTrigger>
              </TabsList>
              
              <TabsContent value="all">
                <Table>
                  <TableCaption>All rate limits for this domain.</TableCaption>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Name</TableHead>
                      <TableHead>Path</TableHead>
                      <TableHead>Limit</TableHead>
                      <TableHead>Action</TableHead>
                      <TableHead>Created</TableHead>
                      <TableHead className="text-right">Actions</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {rateLimits.map((limit) => (
                      <TableRow key={limit.id}>
                        <TableCell className="font-medium">{limit.name}</TableCell>
                        <TableCell><code className="bg-muted px-1 py-0.5 rounded">{limit.path}</code></TableCell>
                        <TableCell>
                          <div className="flex items-center">
                            <Clock className="mr-1 h-3 w-3" /> 
                            {limit.requests} req/{limit.duration}s
                          </div>
                        </TableCell>
                        <TableCell>{getActionBadge(limit.action)}</TableCell>
                        <TableCell>{new Date(limit.createdAt).toLocaleDateString()}</TableCell>
                        <TableCell className="text-right">
                          <Button variant="ghost" size="sm" onClick={() => handleEdit(limit)}>
                            <Edit className="h-4 w-4" />
                          </Button>
                          <Button variant="ghost" size="sm" onClick={() => handleDelete(limit.id)}>
                            <Trash2 className="h-4 w-4" />
                          </Button>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </TabsContent>
              
              <TabsContent value="block">
                <Table>
                  <TableBody>
                    {rateLimits
                      .filter(limit => limit.action === 'block')
                      .map((limit) => (
                        <TableRow key={limit.id}>
                          <TableCell className="font-medium">{limit.name}</TableCell>
                          <TableCell><code className="bg-muted px-1 py-0.5 rounded">{limit.path}</code></TableCell>
                          <TableCell>
                            <div className="flex items-center">
                              <Clock className="mr-1 h-3 w-3" /> 
                              {limit.requests} req/{limit.duration}s
                            </div>
                          </TableCell>
                          <TableCell>{getActionBadge(limit.action)}</TableCell>
                          <TableCell>{new Date(limit.createdAt).toLocaleDateString()}</TableCell>
                          <TableCell className="text-right">
                            <Button variant="ghost" size="sm" onClick={() => handleEdit(limit)}>
                              <Edit className="h-4 w-4" />
                            </Button>
                            <Button variant="ghost" size="sm" onClick={() => handleDelete(limit.id)}>
                              <Trash2 className="h-4 w-4" />
                            </Button>
                          </TableCell>
                        </TableRow>
                      ))}
                  </TableBody>
                </Table>
              </TabsContent>
              
              <TabsContent value="throttle">
                <Table>
                  <TableBody>
                    {rateLimits
                      .filter(limit => limit.action === 'throttle')
                      .map((limit) => (
                        <TableRow key={limit.id}>
                          <TableCell className="font-medium">{limit.name}</TableCell>
                          <TableCell><code className="bg-muted px-1 py-0.5 rounded">{limit.path}</code></TableCell>
                          <TableCell>
                            <div className="flex items-center">
                              <Clock className="mr-1 h-3 w-3" /> 
                              {limit.requests} req/{limit.duration}s
                            </div>
                          </TableCell>
                          <TableCell>{getActionBadge(limit.action)}</TableCell>
                          <TableCell>{new Date(limit.createdAt).toLocaleDateString()}</TableCell>
                          <TableCell className="text-right">
                            <Button variant="ghost" size="sm" onClick={() => handleEdit(limit)}>
                              <Edit className="h-4 w-4" />
                            </Button>
                            <Button variant="ghost" size="sm" onClick={() => handleDelete(limit.id)}>
                              <Trash2 className="h-4 w-4" />
                            </Button>
                          </TableCell>
                        </TableRow>
                      ))}
                  </TableBody>
                </Table>
              </TabsContent>
              
              <TabsContent value="log">
                <Table>
                  <TableBody>
                    {rateLimits
                      .filter(limit => limit.action === 'log')
                      .map((limit) => (
                        <TableRow key={limit.id}>
                          <TableCell className="font-medium">{limit.name}</TableCell>
                          <TableCell><code className="bg-muted px-1 py-0.5 rounded">{limit.path}</code></TableCell>
                          <TableCell>
                            <div className="flex items-center">
                              <Clock className="mr-1 h-3 w-3" /> 
                              {limit.requests} req/{limit.duration}s
                            </div>
                          </TableCell>
                          <TableCell>{getActionBadge(limit.action)}</TableCell>
                          <TableCell>{new Date(limit.createdAt).toLocaleDateString()}</TableCell>
                          <TableCell className="text-right">
                            <Button variant="ghost" size="sm" onClick={() => handleEdit(limit)}>
                              <Edit className="h-4 w-4" />
                            </Button>
                            <Button variant="ghost" size="sm" onClick={() => handleDelete(limit.id)}>
                              <Trash2 className="h-4 w-4" />
                            </Button>
                          </TableCell>
                        </TableRow>
                      ))}
                  </TableBody>
                </Table>
              </TabsContent>
            </Tabs>
          )}
        </CardContent>
      </Card>

      <Dialog open={openDialog} onOpenChange={setOpenDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{currentLimit ? 'Edit Rate Limit' : 'Add New Rate Limit'}</DialogTitle>
            <DialogDescription>
              {currentLimit 
                ? 'Update the rate limit details below.' 
                : 'Configure a new rate limit for this domain.'}
            </DialogDescription>
          </DialogHeader>
          <Form {...form}>
            <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
              <FormField
                control={form.control}
                name="name"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Name</FormLabel>
                    <FormControl>
                      <Input placeholder="API Rate Limit" {...field} />
                    </FormControl>
                    <FormDescription>
                      A descriptive name for this rate limit
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />
              
              <FormField
                control={form.control}
                name="path"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Path Pattern</FormLabel>
                    <FormControl>
                      <Input placeholder="/api/*" {...field} />
                    </FormControl>
                    <FormDescription>
                      Path pattern to apply the rate limit to (e.g. /api/*, /login)
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />
              
              <div className="grid grid-cols-2 gap-4">
                <FormField
                  control={form.control}
                  name="requests"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Requests</FormLabel>
                      <FormControl>
                        <Input 
                          type="number" 
                          min={1} 
                          {...field} 
                          onChange={e => field.onChange(Number(e.target.value))}
                        />
                      </FormControl>
                      <FormDescription>
                        Max number of requests
                      </FormDescription>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                
                <FormField
                  control={form.control}
                  name="duration"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Period (seconds)</FormLabel>
                      <FormControl>
                        <Input 
                          type="number" 
                          min={1} 
                          {...field} 
                          onChange={e => field.onChange(Number(e.target.value))} 
                        />
                      </FormControl>
                      <FormDescription>
                        Time period in seconds
                      </FormDescription>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              </div>
              
              <FormField
                control={form.control}
                name="action"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Action</FormLabel>
                    <Select 
                      onValueChange={field.onChange} 
                      defaultValue={field.value}
                    >
                      <FormControl>
                        <SelectTrigger>
                          <SelectValue placeholder="Select an action" />
                        </SelectTrigger>
                      </FormControl>
                      <SelectContent>
                        <SelectItem value="block">Block</SelectItem>
                        <SelectItem value="throttle">Throttle</SelectItem>
                        <SelectItem value="log">Log Only</SelectItem>
                      </SelectContent>
                    </Select>
                    <FormDescription>
                      Action to take when rate limit is exceeded
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />
              
              <DialogFooter className="pt-4">
                <Button type="submit">
                  {currentLimit ? 'Update Rate Limit' : 'Add Rate Limit'}
                </Button>
              </DialogFooter>
            </form>
          </Form>
        </DialogContent>
      </Dialog>
    </div>
  )
} 