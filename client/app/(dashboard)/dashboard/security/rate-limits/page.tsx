'use client'

import { useState, useEffect } from 'react'
import { Save, Trash2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Form, FormControl, FormDescription, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Switch } from '@/components/ui/switch'
import { toast } from 'sonner'
import { useForm } from 'react-hook-form'
import { Domain, RateLimit } from '@/types/domains'

interface RateLimitFormValues {
  requests_per_second: number
  burst_size: number
  enabled: boolean
}

export default function RateLimitsPage() {
  const [domains, setDomains] = useState<Domain[]>([])
  const [selectedDomainId, setSelectedDomainId] = useState<string>('')
  const [rateLimit, setRateLimit] = useState<RateLimit | null>(null)
  const [isLoading, setIsLoading] = useState(true)

  const form = useForm<RateLimitFormValues>({
    defaultValues: {
      requests_per_second: 100,
      burst_size: 50,
      enabled: true,
    },
  })

  useEffect(() => {
    fetchDomains()
  }, [])

  useEffect(() => {
    if (selectedDomainId) {
      fetchRateLimit(selectedDomainId)
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

  const fetchRateLimit = async (domainId: string) => {
    setIsLoading(true)
    try {
      const response = await fetch(`/api/domains/${domainId}/rate-limit`, {
        headers: { 'Content-Type': 'application/json' },
      })

      if (response.status === 404) {
        setRateLimit(null)
        form.reset({
          requests_per_second: 100,
          burst_size: 50,
          enabled: true,
        })
        return
      }

      if (!response.ok) throw new Error('Failed to fetch rate limit')

      const data = await response.json()
      setRateLimit(data)
      form.reset({
        requests_per_second: data.requests_per_second,
        burst_size: data.burst_size,
        enabled: data.enabled,
      })
    } catch (error) {
      console.error(error)
      setRateLimit(null)
    } finally {
      setIsLoading(false)
    }
  }

  const onSubmit = async (values: RateLimitFormValues) => {
    if (!selectedDomainId) return

    try {
      const response = await fetch(`/api/domains/${selectedDomainId}/rate-limit`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(values),
      })

      if (!response.ok) throw new Error('Failed to save rate limit')

      toast.success('Rate limit saved')
      fetchRateLimit(selectedDomainId)
    } catch (error) {
      console.error(error)
      toast.error('Failed to save rate limit')
    }
  }

  const handleDelete = async () => {
    if (!selectedDomainId || !rateLimit) return
    if (!confirm('Remove rate limiting for this domain?')) return

    try {
      const response = await fetch(`/api/domains/${selectedDomainId}/rate-limit`, {
        method: 'DELETE',
        headers: { 'Content-Type': 'application/json' },
      })

      if (!response.ok) throw new Error('Failed to delete rate limit')

      toast.success('Rate limit removed')
      setRateLimit(null)
      form.reset({
        requests_per_second: 100,
        burst_size: 50,
        enabled: true,
      })
    } catch (error) {
      console.error(error)
      toast.error('Failed to delete rate limit')
    }
  }

  const selectedDomain = domains.find(d => String(d.id) === selectedDomainId)

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-3xl font-bold">Rate Limits</h1>
          <p className="text-muted-foreground">Configure request rate limiting per domain</p>
        </div>
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
          <CardTitle>Rate Limit Configuration</CardTitle>
          <CardDescription>
            {selectedDomain
              ? `Configure rate limiting for ${selectedDomain.name}`
              : 'Select a domain to configure rate limiting'}
          </CardDescription>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="flex justify-center p-6">Loading...</div>
          ) : !selectedDomainId ? (
            <Alert>
              <AlertTitle>No domain selected</AlertTitle>
              <AlertDescription>Select a domain to configure rate limiting.</AlertDescription>
            </Alert>
          ) : (
            <Form {...form}>
              <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
                <div className="grid grid-cols-2 gap-6">
                  <FormField
                    control={form.control}
                    name="requests_per_second"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Requests per Second</FormLabel>
                        <FormControl>
                          <Input
                            type="number"
                            {...field}
                            onChange={e => field.onChange(parseInt(e.target.value))}
                          />
                        </FormControl>
                        <FormDescription>Max requests allowed per second per IP</FormDescription>
                        <FormMessage />
                      </FormItem>
                    )}
                  />

                  <FormField
                    control={form.control}
                    name="burst_size"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Burst Size</FormLabel>
                        <FormControl>
                          <Input
                            type="number"
                            {...field}
                            onChange={e => field.onChange(parseInt(e.target.value))}
                          />
                        </FormControl>
                        <FormDescription>Max burst of requests allowed</FormDescription>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                </div>

                <FormField
                  control={form.control}
                  name="enabled"
                  render={({ field }) => (
                    <FormItem className="flex flex-row items-center justify-between rounded-lg border p-4">
                      <div className="space-y-0.5">
                        <FormLabel className="text-base">Enable Rate Limiting</FormLabel>
                        <FormDescription>
                          When enabled, requests exceeding the limit will receive 429 errors
                        </FormDescription>
                      </div>
                      <FormControl>
                        <Switch checked={field.value} onCheckedChange={field.onChange} />
                      </FormControl>
                    </FormItem>
                  )}
                />

                <div className="flex justify-between">
                  <Button type="submit">
                    <Save className="mr-2 h-4 w-4" />
                    Save Rate Limit
                  </Button>
                  {rateLimit && (
                    <Button type="button" variant="destructive" onClick={handleDelete}>
                      <Trash2 className="mr-2 h-4 w-4" />
                      Remove Rate Limit
                    </Button>
                  )}
                </div>
              </form>
            </Form>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
