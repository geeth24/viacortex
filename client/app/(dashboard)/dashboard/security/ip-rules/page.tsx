'use client'

import { useState, useEffect } from 'react'
import { Plus, Trash2, Shield, Globe } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Form, FormControl, FormDescription, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { Table, TableBody, TableCaption, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { toast } from 'sonner'
import { useForm } from 'react-hook-form'

interface Domain {
  id: string
  name: string
}

interface IPRule {
  id: string
  domainId: string
  ipRange: string
  type: 'allow' | 'block'
  description: string
  createdAt: string
}

export default function IPRulesPage() {
  const [domains, setDomains] = useState<Domain[]>([])
  const [selectedDomain, setSelectedDomain] = useState<string | null>(null)
  const [ipRules, setIPRules] = useState<IPRule[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [openDialog, setOpenDialog] = useState(false)
  const { toast } = useToast()

  const form = useForm({
    defaultValues: {
      ipRange: '',
      type: 'allow',
      description: '',
    },
  })

  useEffect(() => {
    fetchDomains()
  }, [])

  useEffect(() => {
    if (selectedDomain) {
      fetchIPRules(selectedDomain)
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
      toast({
        title: 'Error',
        description: 'Failed to load domains. Please try again.',
        variant: 'destructive',
      })
    } finally {
      setIsLoading(false)
    }
  }

  const fetchIPRules = async (domainId: string) => {
    setIsLoading(true)
    try {
      const response = await fetch(`/api/domains/${domainId}/ip-rules`, {
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
      })
      
      if (!response.ok) throw new Error('Failed to fetch IP rules')
      
      const data = await response.json()
      setIPRules(data.ipRules || [])
    } catch (error) {
      toast({
        title: 'Error',
        description: 'Failed to load IP rules. Please try again.',
        variant: 'destructive',
      })
    } finally {
      setIsLoading(false)
    }
  }

  const onSubmit = async (values: any) => {
    if (!selectedDomain) return

    try {
      const response = await fetch(`/api/domains/${selectedDomain}/ip-rules`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
        body: JSON.stringify(values),
      })

      if (!response.ok) throw new Error('Failed to add IP rule')
      
      toast({
        title: 'Success',
        description: 'IP rule added successfully',
      })
      
      setOpenDialog(false)
      form.reset()
      fetchIPRules(selectedDomain)
    } catch (error) {
      toast({
        title: 'Error',
        description: 'Failed to add IP rule. Please try again.',
        variant: 'destructive',
      })
    }
  }

  const handleDelete = async (ruleId: string) => {
    if (!selectedDomain) return
    if (!confirm('Are you sure you want to delete this IP rule?')) return

    try {
      const response = await fetch(`/api/domains/${selectedDomain}/ip-rules/${ruleId}`, {
        method: 'DELETE',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
      })

      if (!response.ok) throw new Error('Failed to delete IP rule')
      
      toast({
        title: 'Success',
        description: 'IP rule deleted successfully',
      })
      
      fetchIPRules(selectedDomain)
    } catch (error) {
      toast({
        title: 'Error',
        description: 'Failed to delete IP rule. Please try again.',
        variant: 'destructive',
      })
    }
  }

  const handleAdd = () => {
    form.reset({
      ipRange: '',
      type: 'allow',
      description: '',
    })
    setOpenDialog(true)
  }

  const getTypeLabel = (type: string) => {
    switch (type) {
      case 'allow':
        return <Badge className="bg-green-500">Allow</Badge>
      case 'block':
        return <Badge className="bg-red-500">Block</Badge>
      default:
        return <Badge>{type}</Badge>
    }
  }

  return (
    <div className="container mx-auto p-6">
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-3xl font-bold">IP Rules</h1>
        <Button onClick={handleAdd} disabled={!selectedDomain}>
          <Plus className="mr-2 h-4 w-4" />
          Add IP Rule
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
          <CardTitle>IP Rules</CardTitle>
          <CardDescription>
            {selectedDomain 
              ? `Manage IP access rules for ${domains.find(d => d.id === selectedDomain)?.name}` 
              : 'Please select a domain to manage its IP rules'}
          </CardDescription>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="flex justify-center p-6">Loading IP rules...</div>
          ) : !selectedDomain ? (
            <Alert>
              <AlertTitle>No domain selected</AlertTitle>
              <AlertDescription>
                Please select a domain from the dropdown above to manage its IP rules.
              </AlertDescription>
            </Alert>
          ) : ipRules.length === 0 ? (
            <Alert>
              <AlertTitle>No IP rules found</AlertTitle>
              <AlertDescription>
                This domain doesn't have any IP rules yet. Click the "Add IP Rule" button to get started.
              </AlertDescription>
            </Alert>
          ) : (
            <Table>
              <TableCaption>A list of IP rules for this domain.</TableCaption>
              <TableHeader>
                <TableRow>
                  <TableHead>IP Range</TableHead>
                  <TableHead>Type</TableHead>
                  <TableHead>Description</TableHead>
                  <TableHead>Created</TableHead>
                  <TableHead className="text-right">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {ipRules.map((rule) => (
                  <TableRow key={rule.id}>
                    <TableCell className="font-medium font-mono">{rule.ipRange}</TableCell>
                    <TableCell>{getTypeLabel(rule.type)}</TableCell>
                    <TableCell>{rule.description}</TableCell>
                    <TableCell>{new Date(rule.createdAt).toLocaleDateString()}</TableCell>
                    <TableCell className="text-right">
                      <Button variant="ghost" size="sm" onClick={() => handleDelete(rule.id)}>
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
            <DialogTitle>Add New IP Rule</DialogTitle>
            <DialogDescription>
              Enter the details for the new IP rule.
            </DialogDescription>
          </DialogHeader>
          <Form {...form}>
            <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
              <FormField
                control={form.control}
                name="ipRange"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>IP Range</FormLabel>
                    <FormControl>
                      <Input placeholder="192.168.1.0/24" {...field} />
                    </FormControl>
                    <FormDescription>
                      Enter an IP address (192.168.1.1) or CIDR range (192.168.1.0/24)
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />
              
              <FormField
                control={form.control}
                name="type"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Rule Type</FormLabel>
                    <Select 
                      onValueChange={field.onChange} 
                      defaultValue={field.value}
                    >
                      <FormControl>
                        <SelectTrigger>
                          <SelectValue placeholder="Select a rule type" />
                        </SelectTrigger>
                      </FormControl>
                      <SelectContent>
                        <SelectItem value="allow">Allow</SelectItem>
                        <SelectItem value="block">Block</SelectItem>
                      </SelectContent>
                    </Select>
                    <FormDescription>
                      Determine whether to allow or block traffic from this IP range
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />
              
              <FormField
                control={form.control}
                name="description"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Description</FormLabel>
                    <FormControl>
                      <Input placeholder="Office IPs" {...field} />
                    </FormControl>
                    <FormDescription>
                      A brief description of this rule
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />
              
              <DialogFooter>
                <Button type="submit">
                  Add IP Rule
                </Button>
              </DialogFooter>
            </form>
          </Form>
        </DialogContent>
      </Dialog>
    </div>
  )
} 