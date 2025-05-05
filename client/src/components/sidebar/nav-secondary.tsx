'use client';
import * as React from 'react';

import {
  SidebarGroup,
  SidebarGroupContent,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from '@/components/ui/sidebar';
import Link from 'next/link';
import { useTheme } from 'next-themes';

export function NavSecondary({
  items,
  ...props
}: {
  items: {
    title: string;
    url: string;
    icon: React.ElementType;
  }[];
} & React.ComponentPropsWithoutRef<typeof SidebarGroup>) {
  const { theme, setTheme } = useTheme();
  const [mounted, setMounted] = React.useState(false);

  // Only show theme UI after mounting to prevent hydration mismatch
  React.useEffect(() => {
    setMounted(true);
  }, []);

  // Render theme buttons only after component is mounted
  const renderThemeButtons = () => {
    if (!mounted) return null;

    return (
      <div className="flex w-full justify-start gap-2">
        <SidebarMenuItem>
          <SidebarMenuButton
            asChild
            size="sm"
            onClick={() => setTheme('light')}
            variant={theme === 'light' ? 'outline' : 'default'}
          >
            <span>Light</span>
          </SidebarMenuButton>
        </SidebarMenuItem>
        <SidebarMenuItem>
          <SidebarMenuButton
            asChild
            size="sm"
            onClick={() => setTheme('dark')}
            variant={theme === 'dark' ? 'outline' : 'default'}
          >
            <span>Dark</span>
          </SidebarMenuButton>
        </SidebarMenuItem>
        <SidebarMenuItem>
          <SidebarMenuButton
            asChild
            size="sm"
            onClick={() => setTheme('system')}
            variant={theme === 'system' ? 'outline' : 'default'}
          >
            <span>System</span>
          </SidebarMenuButton>
        </SidebarMenuItem>
      </div>
    );
  };

  return (
    <SidebarGroup {...props}>
      <SidebarGroupContent>
        <SidebarMenu>
          {items.map((item) => (
            <SidebarMenuItem key={item.title}>
              <SidebarMenuButton asChild size="sm">
                <Link href={item.url}>
                  <item.icon />
                  <span>{item.title}</span>
                </Link>
              </SidebarMenuButton>
            </SidebarMenuItem>
          ))}
          {renderThemeButtons()}
        </SidebarMenu>
      </SidebarGroupContent>
    </SidebarGroup>
  );
}
