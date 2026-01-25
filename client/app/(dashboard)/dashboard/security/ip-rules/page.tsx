"use client";

import { useState, useEffect } from "react";
import { Plus, Trash2 } from "lucide-react";
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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Switch } from "@/components/ui/switch";
import { toast } from "sonner";
import { useForm } from "react-hook-form";
import { Domain, IPRule } from "@/types/domains";

interface IPRuleFormValues {
  ip_range: string;
  rule_type: "whitelist" | "blacklist";
  description: string;
  enabled: boolean;
}

export default function IPRulesPage() {
  const [domains, setDomains] = useState<Domain[]>([]);
  const [selectedDomainId, setSelectedDomainId] = useState<string>("");
  const [ipRules, setIPRules] = useState<IPRule[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [openDialog, setOpenDialog] = useState(false);

  const form = useForm<IPRuleFormValues>({
    defaultValues: {
      ip_range: "",
      rule_type: "whitelist",
      description: "",
      enabled: true,
    },
  });

  useEffect(() => {
    fetchDomains();
  }, []);

  useEffect(() => {
    if (selectedDomainId) {
      fetchIPRules(selectedDomainId);
    }
  }, [selectedDomainId]);

  const fetchDomains = async () => {
    try {
      const response = await fetch("/api/domains", {
        headers: { "Content-Type": "application/json" },
      });

      if (!response.ok) throw new Error("Failed to fetch domains");

      const data = await response.json();
      const domainList = Array.isArray(data) ? data : [];
      setDomains(domainList);

      if (domainList.length > 0) {
        setSelectedDomainId(String(domainList[0].id));
      }
    } catch (error) {
      console.error(error);
      toast.error("Failed to load domains");
    } finally {
      setIsLoading(false);
    }
  };

  const fetchIPRules = async (domainId: string) => {
    setIsLoading(true);
    try {
      const response = await fetch(`/api/domains/${domainId}/ip-rules`, {
        headers: { "Content-Type": "application/json" },
      });

      if (!response.ok) throw new Error("Failed to fetch IP rules");

      const data = await response.json();
      setIPRules(Array.isArray(data) ? data : []);
    } catch (error) {
      console.error(error);
      toast.error("Failed to load IP rules");
    } finally {
      setIsLoading(false);
    }
  };

  const onSubmit = async (values: IPRuleFormValues) => {
    if (!selectedDomainId) return;

    try {
      const response = await fetch(
        `/api/domains/${selectedDomainId}/ip-rules`,
        {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(values),
        },
      );

      if (!response.ok) throw new Error("Failed to add IP rule");

      toast.success("IP rule added");
      setOpenDialog(false);
      form.reset();
      fetchIPRules(selectedDomainId);
    } catch (error) {
      console.error(error);
      toast.error("Failed to add IP rule");
    }
  };

  const handleDelete = async (ruleId: number) => {
    if (!selectedDomainId) return;
    if (!confirm("Delete this IP rule?")) return;

    try {
      const response = await fetch(
        `/api/domains/${selectedDomainId}/ip-rules/${ruleId}`,
        {
          method: "DELETE",
          headers: { "Content-Type": "application/json" },
        },
      );

      if (!response.ok) throw new Error("Failed to delete IP rule");

      toast.success("IP rule deleted");
      fetchIPRules(selectedDomainId);
    } catch (error) {
      console.error(error);
      toast.error("Failed to delete IP rule");
    }
  };

  const selectedDomain = domains.find((d) => String(d.id) === selectedDomainId);

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-3xl font-bold">IP Rules</h1>
          <p className="text-muted-foreground">
            Whitelist or blacklist IP addresses
          </p>
        </div>
        <Button
          onClick={() => setOpenDialog(true)}
          disabled={!selectedDomainId}
        >
          <Plus className="mr-2 h-4 w-4" />
          Add Rule
        </Button>
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
          <CardTitle>IP Rules</CardTitle>
          <CardDescription>
            {selectedDomain
              ? `Manage IP access rules for ${selectedDomain.name}`
              : "Select a domain to manage IP rules"}
          </CardDescription>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="flex justify-center p-6">Loading...</div>
          ) : !selectedDomainId ? (
            <Alert>
              <AlertTitle>No domain selected</AlertTitle>
              <AlertDescription>
                Select a domain to manage its IP rules.
              </AlertDescription>
            </Alert>
          ) : ipRules.length === 0 ? (
            <Alert>
              <AlertTitle>No IP rules</AlertTitle>
              <AlertDescription>
                Add IP rules to whitelist or blacklist traffic.
              </AlertDescription>
            </Alert>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>IP Range</TableHead>
                  <TableHead>Type</TableHead>
                  <TableHead>Description</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Created</TableHead>
                  <TableHead className="text-right">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {ipRules.map((rule) => (
                  <TableRow key={rule.id}>
                    <TableCell className="font-mono">{rule.ip_range}</TableCell>
                    <TableCell>
                      {rule.rule_type === "whitelist" ? (
                        <Badge className="bg-green-500">Whitelist</Badge>
                      ) : (
                        <Badge variant="destructive">Blacklist</Badge>
                      )}
                    </TableCell>
                    <TableCell>{rule.description}</TableCell>
                    <TableCell>
                      {rule.enabled ? (
                        <Badge variant="default">Active</Badge>
                      ) : (
                        <Badge variant="secondary">Disabled</Badge>
                      )}
                    </TableCell>
                    <TableCell>
                      {new Date(rule.created_at).toLocaleDateString()}
                    </TableCell>
                    <TableCell className="text-right">
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => handleDelete(rule.id)}
                      >
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
            <DialogTitle>Add IP Rule</DialogTitle>
            <DialogDescription>
              Add a new IP whitelist or blacklist rule.
            </DialogDescription>
          </DialogHeader>
          <Form {...form}>
            <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
              <FormField
                control={form.control}
                name="ip_range"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>IP Range (CIDR)</FormLabel>
                    <FormControl>
                      <Input placeholder="192.168.1.0/24" {...field} />
                    </FormControl>
                    <FormDescription>
                      Single IP or CIDR notation
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="rule_type"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Rule Type</FormLabel>
                    <Select
                      onValueChange={field.onChange}
                      defaultValue={field.value}
                    >
                      <FormControl>
                        <SelectTrigger>
                          <SelectValue />
                        </SelectTrigger>
                      </FormControl>
                      <SelectContent>
                        <SelectItem value="whitelist">
                          Whitelist (Allow)
                        </SelectItem>
                        <SelectItem value="blacklist">
                          Blacklist (Block)
                        </SelectItem>
                      </SelectContent>
                    </Select>
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
                      <Input placeholder="Office network" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="enabled"
                render={({ field }) => (
                  <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3">
                    <div className="space-y-0.5">
                      <FormLabel>Enabled</FormLabel>
                      <FormDescription>
                        Activate this rule immediately
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
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => setOpenDialog(false)}
                >
                  Cancel
                </Button>
                <Button type="submit">Add Rule</Button>
              </DialogFooter>
            </form>
          </Form>
        </DialogContent>
      </Dialog>
    </div>
  );
}
