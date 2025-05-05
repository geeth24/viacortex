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
import type { BackendServer } from './servers-table';
import { createServer, updateServer } from '../actions';

const serverSchema = z.object({
  scheme: z.string().min(1),
  ip: z.string().min(1),
  port: z.number().min(1),
  weight: z.number().min(1),
  is_active: z.boolean(),
});

interface ServerFormProps {
  server?: BackendServer | null;
  domainId: number;
  onSuccess?: () => void;
}

export function ServerForm({ server, domainId, onSuccess }: ServerFormProps) {
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const form = useForm<z.infer<typeof serverSchema>>({
    resolver: zodResolver(serverSchema),
    defaultValues: {
      scheme: server?.scheme || 'https',
      ip: server?.ip || '',
      port: server?.port || 443,
      weight: server?.weight || 1,
      is_active: server?.is_active ?? true,
    },
  });

  async function onSubmit(values: z.infer<typeof serverSchema>) {
    setIsSubmitting(true);
    setError(null);

    try {
      if (server?.id) {
        await updateServer(domainId, server.id, values);
      } else {
        await createServer(domainId, values);
      }
      onSuccess?.();
    } catch (err) {
      console.error('Failed to save server:', err);
      setError(err instanceof Error ? err.message : 'Failed to save server');
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-8">
        <FormField
          control={form.control}
          name="scheme"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Scheme</FormLabel>
              <FormControl>
                <Input placeholder="https" {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

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
                <Input
                  type="number"
                  {...field}
                  onChange={(e) => field.onChange(Number(e.target.value))}
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="weight"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Weight</FormLabel>
              <FormControl>
                <Input
                  type="number"
                  {...field}
                  onChange={(e) => field.onChange(Number(e.target.value))}
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="is_active"
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

        {error && <div className="text-sm text-red-500">{error}</div>}

        <Button type="submit" disabled={isSubmitting}>
          {isSubmitting ? 'Saving...' : 'Save Server'}
        </Button>
      </form>
    </Form>
  );
}
