"use client";

import * as React from "react";
import {
  LayoutDashboard,
  Settings2,
  Shield,
  Globe,
  Users,
  FileText,
} from "lucide-react";

import { NavMain } from "@/components/sidebar/nav-main";
import { NavUser } from "@/components/sidebar/nav-user";
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarMenu,
  SidebarRail,
} from "@/components/ui/sidebar";
import Logo from "../logo";

const data = {
  navMain: [
    {
      title: "Dashboard",
      url: "/dashboard",
      icon: LayoutDashboard,
      isActive: true,
      items: [
        {
          title: "Overview",
          url: "/dashboard",
        },
      ],
    },
    {
      title: "Domains",
      url: "/dashboard/domains",
      icon: Globe,
      items: [
        {
          title: "All Domains",
          url: "/dashboard/domains",
        },
        {
          title: "Redirects",
          url: "/dashboard/domains/redirects",
        },
      ],
    },
    {
      title: "Security",
      url: "/dashboard/security",
      icon: Shield,
      items: [
        {
          title: "IP Rules",
          url: "/dashboard/security/ip-rules",
        },
        {
          title: "Rate Limits",
          url: "/dashboard/security/rate-limits",
        },
      ],
    },
    {
      title: "Logs",
      url: "/dashboard/logs",
      icon: FileText,
      items: [
        {
          title: "Access Logs",
          url: "/dashboard/logs/access",
        },
        {
          title: "Audit Logs",
          url: "/dashboard/logs/audit",
        },
      ],
    },
    {
      title: "Users",
      url: "/dashboard/users",
      icon: Users,
      items: [
        {
          title: "All Users",
          url: "/dashboard/users",
        },
      ],
    },
    {
      title: "Settings",
      url: "/dashboard/settings",
      icon: Settings2,
      items: [
        {
          title: "Profile",
          url: "/dashboard/settings/profile",
        },
      ],
    },
  ],
};

export function AppSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {
  return (
    <Sidebar collapsible="icon" {...props}>
      <SidebarHeader>
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton size="lg" asChild>
              <a href="/dashboard">
                <div className="flex aspect-square size-8 items-center justify-center rounded-lg bg-primary text-primary-foreground">
                  <Logo className="size-4 text-primary-foreground" />
                </div>
                <div className="grid flex-1 text-left text-sm leading-tight">
                  <span className="truncate font-semibold">ViaCortex</span>
                  <span className="truncate text-xs">Reverse Proxy</span>
                </div>
              </a>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarHeader>
      <SidebarContent>
        <NavMain items={data.navMain} />
      </SidebarContent>
      <SidebarFooter>
        <NavUser />
      </SidebarFooter>
      <SidebarRail />
    </Sidebar>
  );
}
