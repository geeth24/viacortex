'use client';

import { zodResolver } from '@hookform/resolvers/zod';
import { useForm } from 'react-hook-form';
import * as z from 'zod';
import { useState } from 'react';

import { Button } from '@/components/ui/button';
import { Checkbox } from '@/components/ui/checkbox';
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import type { Domain } from './domains-table';
import { createDomain, updateDomain } from '../actions';
const backendServerSchema = z.object({
  scheme: z.string().min(1),
  ip: z.string().min(1),
  port: z.number().min(1),
  weight: z.number().min(1),
  is_active: z.boolean(),
});

const domainObjectSchema = z.object({
  name: z.string().min(1),
  target_url: z
    .string()
    .min(1)
    
    .pipe(z.string().url()),
  ssl_enabled: z.boolean(),
  health_check_enabled: z.boolean(),
  health_check_interval: z.number().min(1),
  custom_error_pages: z.record(z.string(), z.string()),
  created_at: z.string(),
  updated_at: z.string(),
});
const formSchema = z.object({
  domain: domainObjectSchema,
  backend_servers: z.array(backendServerSchema),
});

interface DomainFormProps {
  domain?: Domain | null;
  onSuccess?: () => void;
}

export function DomainForm({ domain, onSuccess }: DomainFormProps) {
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const form = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      domain: {
        name: domain?.domain?.name || '',
        target_url: domain?.domain?.target_url || '',
        ssl_enabled: domain?.domain?.ssl_enabled || false,
        health_check_enabled: domain?.domain?.health_check_enabled || false,
        health_check_interval: domain?.domain?.health_check_interval || 10,
        custom_error_pages: domain?.domain?.custom_error_pages || {},
        created_at: domain?.domain?.created_at || new Date().toISOString(),
        updated_at: domain?.domain?.updated_at || new Date().toISOString(),
      },
      backend_servers: domain?.backend_servers || [
        {
          scheme: 'https',
          ip: '',
          port: 443,
          weight: 1,
          is_active: true,
        },
      ],
    },
  });

  async function onSubmit(values: z.infer<typeof formSchema>) {
    setIsSubmitting(true);
    setError(null);

    try {
      if (domain) {
        console.log('Updating domain:', values);
        await updateDomain(domain.domain.id as number, {
          domain: values.domain,
          backend_servers: values.backend_servers,
        });
      } else {
        console.log('Creating domain:', values);
        await createDomain({
          domain: values.domain,
          backend_servers: values.backend_servers,
        });
      }
      onSuccess?.();
    } catch (err) {
      console.error('Failed to save domain:', err);
      setError(err instanceof Error ? err.message : 'Failed to save domain');
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-8">
        <FormField
          control={form.control}
          name="domain.name"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Domain Name</FormLabel>
              <FormControl>
                <Input placeholder="example.com" {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="domain.target_url"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Target URL</FormLabel>
              <FormControl>
                <Input
                  placeholder="maxbrowser.win"
                  {...field}
                  onChange={(e) => {
                    const value = e.target.value.trim();
                    field.onChange(value);
                  }}
                />
              </FormControl>
              <FormDescription>
                Protocol (https://) will be added automatically if not provided
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="domain.ssl_enabled"
          render={({ field }) => (
            <FormItem className="flex flex-row items-start space-x-3 space-y-0">
              <FormControl>
                <Checkbox checked={field.value} onCheckedChange={field.onChange} />
              </FormControl>
              <div className="space-y-1 leading-none">
                <FormLabel>SSL Enabled</FormLabel>
                <FormDescription>Enable SSL/TLS for this domain</FormDescription>
              </div>
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="domain.health_check_enabled"
          render={({ field }) => (
            <FormItem className="flex flex-row items-start space-x-3 space-y-0">
              <FormControl>
                <Checkbox checked={field.value} onCheckedChange={field.onChange} />
              </FormControl>
              <div className="space-y-1 leading-none">
                <FormLabel>Health Check Enabled</FormLabel>
                <FormDescription>Enable health checks for backend servers</FormDescription>
              </div>
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="domain.health_check_interval"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Health Check Interval (seconds)</FormLabel>
              <FormControl>
                <Input
                  type="number"
                  {...field}
                  onChange={(e) => {
                    const value = e.target.value ? Number(e.target.value) : '';
                    field.onChange(value);
                  }}
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        {form.watch('backend_servers').map((_, index) => (
          <div key={index} className="space-y-4 rounded-lg border p-4">
            <h3 className="font-medium">Backend Server {index + 1}</h3>

            <FormField
              control={form.control}
              name={`backend_servers.${index}.scheme`}
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Scheme</FormLabel>
                  <FormControl>
                    <Input {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name={`backend_servers.${index}.ip`}
              render={({ field }) => (
                <FormItem>
                  <FormLabel>IP Address</FormLabel>
                  <FormControl>
                    <Input {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name={`backend_servers.${index}.port`}
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Port</FormLabel>
                  <FormControl>
                    <Input
                      type="number"
                      {...field}
                      onChange={(e) => {
                        const value = e.target.value ? Number(e.target.value) : '';
                        field.onChange(value);
                      }}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name={`backend_servers.${index}.weight`}
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Weight</FormLabel>
                  <FormControl>
                    <Input
                      type="number"
                      {...field}
                      onChange={(e) => {
                        const value = e.target.value ? Number(e.target.value) : '';
                        field.onChange(value);
                      }}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name={`backend_servers.${index}.is_active`}
              render={({ field }) => (
                <FormItem className="flex flex-row items-start space-x-3 space-y-0">
                  <FormControl>
                    <Checkbox checked={field.value} onCheckedChange={field.onChange} />
                  </FormControl>
                  <div className="space-y-1 leading-none">
                    <FormLabel>Active</FormLabel>
                    <FormDescription>Enable this backend server</FormDescription>
                  </div>
                </FormItem>
              )}
            />
          </div>
        ))}

        {error && <div className="text-sm text-red-500">{error}</div>}

        <div className="flex justify-end gap-4">
          <Button
            type="button"
            variant="outline"
            onClick={() => {
              const currentServers = form.getValues('backend_servers');
              form.setValue('backend_servers', [
                ...currentServers,
                {
                  scheme: 'https',
                  ip: '',
                  port: 443,
                  weight: 1,
                  is_active: true,
                },
              ]);
            }}
          >
            Add Backend Server
          </Button>

          <Button type="submit" disabled={isSubmitting}>
            {isSubmitting ? 'Saving...' : 'Save Domain'}
          </Button>
        </div>
      </form>
    </Form>
  );
}
