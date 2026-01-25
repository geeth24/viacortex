"use client";

import { useState, useEffect } from "react";
import { Globe, Shield, ArrowUpRight, Users } from "lucide-react";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { toast } from "sonner";
import { Domain } from "@/types/domains";
import Link from "next/link";

export default function DashboardPage() {
  const [domains, setDomains] = useState<Domain[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    fetchDomains();
  }, []);

  const fetchDomains = async () => {
    try {
      const response = await fetch("/api/domains", {
        headers: { "Content-Type": "application/json" },
      });

      if (!response.ok) throw new Error("Failed to fetch domains");

      const data = await response.json();
      setDomains(Array.isArray(data) ? data : []);
    } catch (error) {
      console.error(error);
      toast.error("Failed to load dashboard data");
    } finally {
      setIsLoading(false);
    }
  };

  const activeDomains = domains.filter((d) => d.enabled).length;
  const sslDomains = domains.filter((d) => d.ssl_enabled).length;

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold">Dashboard</h1>
        <p className="text-muted-foreground">Overview of your reverse proxy</p>
      </div>

      {isLoading ? (
        <div className="flex justify-center p-12">Loading...</div>
      ) : (
        <>
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">
                  Total Domains
                </CardTitle>
                <Globe className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{domains.length}</div>
                <p className="text-xs text-muted-foreground">
                  {activeDomains} active
                </p>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">
                  SSL Certificates
                </CardTitle>
                <Shield className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{sslDomains}</div>
                <p className="text-xs text-muted-foreground">
                  Auto-managed via Let&apos;s Encrypt
                </p>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">
                  Active Proxies
                </CardTitle>
                <ArrowUpRight className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{activeDomains}</div>
                <p className="text-xs text-muted-foreground">
                  Currently routing traffic
                </p>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">Disabled</CardTitle>
                <Users className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">
                  {domains.length - activeDomains}
                </div>
                <p className="text-xs text-muted-foreground">Domains paused</p>
              </CardContent>
            </Card>
          </div>

          <div className="grid gap-4 md:grid-cols-2">
            <Card>
              <CardHeader>
                <CardTitle>Recent Domains</CardTitle>
                <CardDescription>
                  Your most recently added domains
                </CardDescription>
              </CardHeader>
              <CardContent>
                {domains.length === 0 ? (
                  <p className="text-sm text-muted-foreground">
                    No domains configured yet.
                  </p>
                ) : (
                  <div className="space-y-4">
                    {domains.slice(0, 5).map((domain) => (
                      <div
                        key={domain.id}
                        className="flex items-center justify-between"
                      >
                        <div className="space-y-1">
                          <p className="text-sm font-medium leading-none">
                            {domain.name}
                          </p>
                          <p className="text-sm text-muted-foreground truncate max-w-[200px]">
                            {domain.target_url}
                          </p>
                        </div>
                        <div
                          className={`h-2 w-2 rounded-full ${domain.enabled ? "bg-green-500" : "bg-gray-400"}`}
                        />
                      </div>
                    ))}
                  </div>
                )}
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>Quick Actions</CardTitle>
                <CardDescription>Common tasks</CardDescription>
              </CardHeader>
              <CardContent className="space-y-2">
                <Link
                  href="/dashboard/domains"
                  className="block p-3 rounded-lg border hover:bg-accent transition-colors"
                >
                  <div className="flex items-center gap-3">
                    <Globe className="h-5 w-5" />
                    <div>
                      <p className="font-medium">Manage Domains</p>
                      <p className="text-sm text-muted-foreground">
                        Add or edit proxy configurations
                      </p>
                    </div>
                  </div>
                </Link>
                <Link
                  href="/dashboard/security/ip-rules"
                  className="block p-3 rounded-lg border hover:bg-accent transition-colors"
                >
                  <div className="flex items-center gap-3">
                    <Shield className="h-5 w-5" />
                    <div>
                      <p className="font-medium">Security Rules</p>
                      <p className="text-sm text-muted-foreground">
                        Configure IP rules and rate limits
                      </p>
                    </div>
                  </div>
                </Link>
              </CardContent>
            </Card>
          </div>
        </>
      )}
    </div>
  );
}
