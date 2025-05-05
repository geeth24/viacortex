'use client';

import { PlusCircle } from 'lucide-react';
import { useState } from 'react';

import { Button } from '@/components/ui/button';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { DomainForm } from './domain-form';
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetTrigger } from '@/components/ui/sheet';

export interface BackendServer {
  scheme: string;
  ip: string;
  port: number;
  weight: number;
  is_active: boolean;
}

export interface DomainObject {
  id?: number | null;
  name: string;
  target_url: string;
  ssl_enabled: boolean;
  health_check_enabled: boolean;
  health_check_interval: number;
  custom_error_pages: unknown;
  created_at: string;
  updated_at: string;
}

export interface Domain {
  domain: DomainObject;
  backend_servers: BackendServer[];
}

interface DomainsTableProps {
  domains: Domain[];
}

export function DomainsTable({ domains }: DomainsTableProps) {
  const [open, setOpen] = useState(false);
  const [selectedDomain, setSelectedDomain] = useState<Domain | null>(null);

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-2xl font-bold">Domains</h2>
        <Sheet open={open} onOpenChange={setOpen}>
          <SheetTrigger asChild>
            <Button onClick={() => setSelectedDomain(null)}>
              <PlusCircle className="mr-2 h-4 w-4" />
              Add Domain
            </Button>
          </SheetTrigger>
          <SheetContent className="h-full w-full overflow-y-auto sm:max-w-[600px]">
            <SheetHeader>
              <SheetTitle>{selectedDomain ? 'Edit Domain' : 'Add Domain'}</SheetTitle>
            </SheetHeader>
            <DomainForm domain={selectedDomain} onSuccess={() => setOpen(false)} />
          </SheetContent>
        </Sheet>
      </div>

      <div className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Name</TableHead>
              <TableHead>Target URL</TableHead>
              <TableHead>SSL</TableHead>
              <TableHead>Health Check</TableHead>
              <TableHead>Backend Servers</TableHead>
              <TableHead>Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {domains.map((domain) => (
              <TableRow key={domain.domain.id}>
                <TableCell>{domain.domain.name}</TableCell>
                <TableCell>{domain.domain.target_url}</TableCell>
                <TableCell>{domain.domain.ssl_enabled ? 'Yes' : 'No'}</TableCell>
                <TableCell>
                  {domain.domain.health_check_enabled
                    ? `Every ${domain.domain.health_check_interval}s`
                    : 'Disabled'}
                </TableCell>
                <TableCell>{domain.backend_servers?.length || 0}</TableCell>
                <TableCell>
                  <Sheet>
                    <SheetTrigger asChild>
                      <Button variant="outline" size="sm" onClick={() => setSelectedDomain(domain)}>
                        Edit
                      </Button>
                    </SheetTrigger>
                    <SheetContent className="h-full w-full overflow-y-auto sm:max-w-[600px]">
                      <SheetHeader>
                        <SheetTitle>Edit Domain</SheetTitle>
                      </SheetHeader>
                      <DomainForm domain={domain} onSuccess={() => setOpen(false)} />
                    </SheetContent>
                  </Sheet>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>
    </div>
  );
}
