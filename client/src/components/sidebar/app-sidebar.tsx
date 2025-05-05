'use client';

import * as React from 'react';
import {
  Activity,
  LayoutDashboard,
  Settings2,
  Shield,
  Globe,
  Server,
  Users,
  FileText,
} from 'lucide-react';

import { NavMain } from '@/components/sidebar/nav-main';
import { NavUser } from '@/components/sidebar/nav-user';
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarMenu,
  SidebarRail,
} from '@/components/ui/sidebar';
import Logo from '../logo';

const data = {
  navMain: [
    {
      title: 'Dashboard',
      url: '/dashboard',
      icon: LayoutDashboard,
      isActive: true,
      items: [
        {
          title: 'Overview',
          url: '/dashboard',
        },
        {
          title: 'Metrics',
          url: '/dashboard/metrics',
        },
        {
          title: 'Logs',
          url: '/dashboard/logs',
        },
      ],
    },
    {
      title: 'Domains',
      url: '/domains',
      icon: Globe,
      items: [
        {
          title: 'All Domains',
          url: '/dashboard/domains',
        },
        {
          title: 'SSL Certificates',
          url: '/dashboard/domains/certificates',
        },
        {
          title: 'Error Pages',
          url: '/dashboard/domains/error-pages',
        },
      ],
    },
    {
      title: 'Backend Servers',
      url: '/dashboard/backends',
      icon: Server,
      items: [
        {
          title: 'Server List',
          url: '/dashboard/backends/list',
        },
        {
          title: 'Health Checks',
          url: '/dashboard/backends/health',
        },
      ],
    },
    {
      title: 'Security',
      url: '/dashboard/security',
      icon: Shield,
      items: [
        {
          title: 'IP Rules',
          url: '/dashboard/security/ip-rules',
        },
        {
          title: 'Rate Limits',
          url: '/dashboard/security/rate-limits',
        },
        {
          title: 'Blacklist',
          url: '/dashboard/security/blacklist',
        },
        {
          title: 'Auth Settings',
          url: '/dashboard/security/auth',
        },
      ],
    },
    {
      title: 'Load Balancing',
      url: '/dashboard/load-balancing',
      icon: Activity,
      items: [
        {
          title: 'Rules',
          url: '/dashboard/load-balancing/rules',
        },
        {
          title: 'Status',
          url: '/dashboard/load-balancing/status',
        },
      ],
    },
    {
      title: 'Users',
      url: '/dashboard/users',
      icon: Users,
      items: [
        {
          title: 'All Users',
          url: '/dashboard/users/all',
        },
        {
          title: 'Roles',
          url: '/dashboard/users/roles',
        },
        {
          title: 'Permissions',
          url: '/dashboard/users/permissions',
        },
      ],
    },
    {
      title: 'Audit',
      url: '/dashboard/audit',
      icon: FileText,
      items: [
        {
          title: 'Audit Logs',
          url: '/dashboard/audit/logs',
        },
        {
          title: 'Entity History',
          url: '/dashboard/audit/history',
        },
      ],
    },
    {
      title: 'Settings',
      url: '/dashboard/settings',
      icon: Settings2,
      items: [
        {
          title: 'General',
          url: '/dashboard/settings/general',
        },
        {
          title: 'Profile',
          url: '/dashboard/settings/profile',
        },
        {
          title: 'Integrations',
          url: '/dashboard/settings/integrations',
        },
      ],
    },
  ],

  navSecondary: [
    {
      title: 'Settings',
      url: '/dashboard/settings',
      icon: Settings2,
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
              <a href="#">
                <div className="flex aspect-square size-8 items-center justify-center rounded-lg bg-primary text-primary-foreground">
                  <Logo className="size-4 text-primary-foreground" />
                </div>
                <div className="grid flex-1 text-left text-sm leading-tight">
                  <span className="truncate font-semibold">ViaCortex</span>
                  <span className="truncate text-xs">Fast, Secure, and Scalable</span>
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
        {/* <NavSecondary items={data.navSecondary} /> */}
        <NavUser />
      </SidebarFooter>
      <SidebarRail />
    </Sidebar>
  );
}
