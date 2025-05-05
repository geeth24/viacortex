'use client';

import { PlusCircle } from 'lucide-react';
import { useState } from 'react';
import { useParams } from 'next/navigation';

import { Button } from '@/components/ui/button';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { ServerForm } from './server-form';
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetTrigger } from '@/components/ui/sheet';
import { cn } from '@/lib/utils';

export interface NewBackendServer {
  scheme: string;
  ip: string;
  port: number;
  weight: number;
  is_active: boolean;
}

export interface BackendServer {
  id: number;
  domain_id: number;
  scheme: string;
  ip: string;
  port: number;
  weight: number;
  is_active: boolean;
  last_health_check: string;
  health_status: string;
}

interface ServersTableProps {
  servers: BackendServer[];
}

export function ServersTable({ servers }: ServersTableProps) {
  const params = useParams();
  const [open, setOpen] = useState(false);
  const [selectedServer, setSelectedServer] = useState<BackendServer | null>(null);

  const handleServerSelect = (server: BackendServer | null) => {
    setSelectedServer(server);
    setOpen(true);
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-2xl font-bold">Servers</h2>
        <Sheet open={open} onOpenChange={setOpen}>
          <SheetTrigger asChild>
            <Button onClick={() => handleServerSelect(null)}>
              <PlusCircle className="mr-2 h-4 w-4" />
              Add Server
            </Button>
          </SheetTrigger>
          <SheetContent className="w-full sm:max-w-[600px]">
            <SheetHeader>
              <SheetTitle>{selectedServer ? 'Edit Server' : 'Add Server'}</SheetTitle>
            </SheetHeader>
            <ServerForm
              server={selectedServer}
              domainId={Number(params.id)}
              onSuccess={() => setOpen(false)}
            />
          </SheetContent>
        </Sheet>
      </div>

      <div className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>IP Address</TableHead>
              <TableHead>Port</TableHead>
              <TableHead>Weight</TableHead>
              <TableHead>Active</TableHead>
              <TableHead>Health Status</TableHead>
              <TableHead>Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {servers.map((server) => (
              <TableRow
                key={server.id}
                className={cn(
                  server.health_status === 'healthy' ? '' : 'bg-red-300/50 dark:bg-red-900/50',
                )}
              >
                <TableCell>{server.ip}</TableCell>
                <TableCell>{server.port}</TableCell>
                <TableCell>{server.weight}</TableCell>
                <TableCell>{server.is_active ? 'Yes' : 'No'}</TableCell>
                <TableCell>{server.health_status}</TableCell>
                <TableCell>
                  <Button variant="outline" size="sm" onClick={() => handleServerSelect(server)}>
                    Edit
                  </Button>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>
    </div>
  );
}
