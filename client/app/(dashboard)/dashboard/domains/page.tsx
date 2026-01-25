"use client";

import { useState, useEffect } from "react";
import {
  Plus,
  Edit,
  Trash2,
  MoreHorizontal,
  Power,
  PowerOff,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Switch } from "@/components/ui/switch";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { Domain } from "@/types/domains";

interface DomainFormValues {
  name: string;
  target_url: string;
  ssl_enabled: boolean;
  force_https: boolean;
  http2_enabled: boolean;
  hsts_enabled: boolean;
  websocket_enabled: boolean;
  basic_auth_enabled: boolean;
  basic_auth_username: string;
  basic_auth_password: string;
}

export default function DomainsPage() {
  const [domains, setDomains] = useState<Domain[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [openDialog, setOpenDialog] = useState(false);
  const [currentDomain, setCurrentDomain] = useState<Domain | null>(null);

  const form = useForm<DomainFormValues>({
    defaultValues: {
      name: "",
      target_url: "",
      ssl_enabled: true,
      force_https: true,
      http2_enabled: true,
      hsts_enabled: false,
      websocket_enabled: true,
      basic_auth_enabled: false,
      basic_auth_username: "",
      basic_auth_password: "",
    },
  });

  useEffect(() => {
    fetchDomains();
  }, []);

  const fetchDomains = async () => {
    setIsLoading(true);
    try {
      const response = await fetch("/api/domains", {
        headers: {
          "Content-Type": "application/json",
        },
      });

      if (!response.ok) throw new Error("Failed to fetch domains");

      const data = await response.json();
      setDomains(Array.isArray(data) ? data : []);
    } catch (error) {
      console.error(error);
      toast.error("Failed to load domains");
    } finally {
      setIsLoading(false);
    }
  };

  const onSubmit = async (values: DomainFormValues) => {
    try {
      const method = currentDomain ? "PUT" : "POST";
      const url = currentDomain
        ? `/api/domains/${currentDomain.id}`
        : "/api/domains";

      const response = await fetch(url, {
        method,
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(values),
      });

      if (!response.ok) throw new Error("Failed to save domain");

      toast.success(currentDomain ? "Domain updated" : "Domain created");

      setOpenDialog(false);
      form.reset();
      setCurrentDomain(null);
      fetchDomains();
    } catch (error) {
      console.error(error);
      toast.error("Failed to save domain");
    }
  };

  const handleDelete = async (id: number) => {
    if (!confirm("Are you sure you want to delete this domain?")) return;

    try {
      const response = await fetch(`/api/domains/${id}`, {
        method: "DELETE",
        headers: {
          "Content-Type": "application/json",
        },
      });

      if (!response.ok) throw new Error("Failed to delete domain");

      toast.success("Domain deleted");
      fetchDomains();
    } catch (error) {
      console.error(error);
      toast.error("Failed to delete domain");
    }
  };

  const handleToggle = async (domain: Domain) => {
    try {
      const response = await fetch(`/api/domains/${domain.id}/toggle`, {
        method: "PATCH",
        headers: {
          "Content-Type": "application/json",
        },
      });

      if (!response.ok) throw new Error("Failed to toggle domain");

      toast.success(domain.enabled ? "Domain disabled" : "Domain enabled");
      fetchDomains();
    } catch (error) {
      console.error(error);
      toast.error("Failed to toggle domain");
    }
  };

  const handleEdit = (domain: Domain) => {
    setCurrentDomain(domain);
    form.reset({
      name: domain.name,
      target_url: domain.target_url,
      ssl_enabled: domain.ssl_enabled,
      force_https: domain.force_https,
      http2_enabled: domain.http2_enabled,
      hsts_enabled: domain.hsts_enabled,
      websocket_enabled: domain.websocket_enabled,
      basic_auth_enabled: domain.basic_auth_enabled,
      basic_auth_username: domain.basic_auth_username || "",
      basic_auth_password: "",
    });
    setOpenDialog(true);
  };

  const handleAdd = () => {
    setCurrentDomain(null);
    form.reset({
      name: "",
      target_url: "",
      ssl_enabled: true,
      force_https: true,
      http2_enabled: true,
      hsts_enabled: false,
      websocket_enabled: true,
      basic_auth_enabled: false,
      basic_auth_username: "",
      basic_auth_password: "",
    });
    setOpenDialog(true);
  };

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-3xl font-bold">Domains</h1>
          <p className="text-muted-foreground">Manage your proxy domains</p>
        </div>
        <Button onClick={handleAdd}>
          <Plus className="mr-2 h-4 w-4" />
          Add Domain
        </Button>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>All Domains</CardTitle>
          <CardDescription>
            Configure reverse proxy for your domains
          </CardDescription>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="flex justify-center p-6">Loading...</div>
          ) : domains.length === 0 ? (
            <Alert>
              <AlertTitle>No domains</AlertTitle>
              <AlertDescription>
                Add your first domain to get started with the reverse proxy.
              </AlertDescription>
            </Alert>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Domain</TableHead>
                  <TableHead>Target</TableHead>
                  <TableHead>SSL</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Created</TableHead>
                  <TableHead className="text-right">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {domains.map((domain) => (
                  <TableRow key={domain.id}>
                    <TableCell className="font-medium">{domain.name}</TableCell>
                    <TableCell className="max-w-xs truncate">
                      {domain.target_url}
                    </TableCell>
                    <TableCell>
                      {domain.ssl_enabled ? (
                        <Badge variant="default">SSL</Badge>
                      ) : (
                        <Badge variant="outline">No SSL</Badge>
                      )}
                    </TableCell>
                    <TableCell>
                      {domain.enabled ? (
                        <Badge className="bg-green-500">Active</Badge>
                      ) : (
                        <Badge variant="secondary">Disabled</Badge>
                      )}
                    </TableCell>
                    <TableCell>
                      {new Date(domain.created_at).toLocaleDateString()}
                    </TableCell>
                    <TableCell className="text-right">
                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <Button variant="ghost" size="sm">
                            <MoreHorizontal className="h-4 w-4" />
                          </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end">
                          <DropdownMenuItem onClick={() => handleEdit(domain)}>
                            <Edit className="mr-2 h-4 w-4" />
                            Edit
                          </DropdownMenuItem>
                          <DropdownMenuItem
                            onClick={() => handleToggle(domain)}
                          >
                            {domain.enabled ? (
                              <>
                                <PowerOff className="mr-2 h-4 w-4" />
                                Disable
                              </>
                            ) : (
                              <>
                                <Power className="mr-2 h-4 w-4" />
                                Enable
                              </>
                            )}
                          </DropdownMenuItem>
                          <DropdownMenuItem
                            onClick={() => handleDelete(domain.id)}
                            className="text-destructive"
                          >
                            <Trash2 className="mr-2 h-4 w-4" />
                            Delete
                          </DropdownMenuItem>
                        </DropdownMenuContent>
                      </DropdownMenu>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

      <Dialog open={openDialog} onOpenChange={setOpenDialog}>
        <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>
              {currentDomain ? "Edit Domain" : "Add Domain"}
            </DialogTitle>
            <DialogDescription>
              {currentDomain
                ? "Update the domain configuration."
                : "Configure a new domain for reverse proxy."}
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
                      The domain that will be proxied (without protocol)
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
                      <Input
                        placeholder="https://backend.example.com"
                        {...field}
                      />
                    </FormControl>
                    <FormDescription>
                      The backend server URL to proxy requests to
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <div className="grid grid-cols-2 gap-4">
                <FormField
                  control={form.control}
                  name="ssl_enabled"
                  render={({ field }) => (
                    <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3">
                      <div className="space-y-0.5">
                        <FormLabel>SSL/TLS</FormLabel>
                        <FormDescription>Auto SSL certificate</FormDescription>
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
                  name="force_https"
                  render={({ field }) => (
                    <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3">
                      <div className="space-y-0.5">
                        <FormLabel>Force HTTPS</FormLabel>
                        <FormDescription>
                          Redirect HTTP to HTTPS
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
                  name="http2_enabled"
                  render={({ field }) => (
                    <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3">
                      <div className="space-y-0.5">
                        <FormLabel>HTTP/2</FormLabel>
                        <FormDescription>
                          Enable HTTP/2 protocol
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
                  name="hsts_enabled"
                  render={({ field }) => (
                    <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3">
                      <div className="space-y-0.5">
                        <FormLabel>HSTS</FormLabel>
                        <FormDescription>
                          Strict transport security
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
                  name="websocket_enabled"
                  render={({ field }) => (
                    <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3">
                      <div className="space-y-0.5">
                        <FormLabel>WebSocket</FormLabel>
                        <FormDescription>
                          Support WebSocket connections
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
                  name="basic_auth_enabled"
                  render={({ field }) => (
                    <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3">
                      <div className="space-y-0.5">
                        <FormLabel>Basic Auth</FormLabel>
                        <FormDescription>
                          Require authentication
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
              </div>

              {form.watch("basic_auth_enabled") && (
                <div className="grid grid-cols-2 gap-4">
                  <FormField
                    control={form.control}
                    name="basic_auth_username"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Username</FormLabel>
                        <FormControl>
                          <Input {...field} />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                  <FormField
                    control={form.control}
                    name="basic_auth_password"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Password</FormLabel>
                        <FormControl>
                          <Input type="password" {...field} />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                </div>
              )}

              <DialogFooter>
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => setOpenDialog(false)}
                >
                  Cancel
                </Button>
                <Button type="submit">
                  {currentDomain ? "Update" : "Create"}
                </Button>
              </DialogFooter>
            </form>
          </Form>
        </DialogContent>
      </Dialog>
    </div>
  );
}
