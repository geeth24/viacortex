'use client';

import { ChevronsUpDown, LogOut, Settings } from 'lucide-react';
import { useEffect, useState } from 'react';
import crypto from 'crypto';
import { useRouter } from 'next/navigation';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import {
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  useSidebar,
} from '@/components/ui/sidebar';
import { useAuth } from '@/contexts/auth-context';

function getEmailHash(email: string) {
  return crypto.createHash('md5').update(email.trim().toLowerCase()).digest('hex');
}

export function NavUser() {
  const router = useRouter();
  const { isMobile } = useSidebar();
  const { user, logout } = useAuth();
  const [useAvatarFallback, setUseAvatarFallback] = useState(false);
  const [avatarUrl, setAvatarUrl] = useState<string>();

  useEffect(() => {
    async function checkGravatarExists() {
      if (!user?.email) return;

      const emailHash = getEmailHash(user.email);
      const gravatarUrl = `https://www.gravatar.com/avatar/${emailHash}?d=404`;

      try {
        const response = await fetch(gravatarUrl);
        if (response.ok) {
          setAvatarUrl(gravatarUrl);
          setUseAvatarFallback(false);
        } else {
          setUseAvatarFallback(true);
        }
      } catch (error) {
        console.error('Error checking Gravatar:', error);
        setUseAvatarFallback(true);
      }
    }

    checkGravatarExists();
  }, [user?.email]);

  return (
    <SidebarMenu>
      <SidebarMenuItem>
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <SidebarMenuButton
              size="lg"
              className="data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground"
            >
              <Avatar className="h-8 w-8 rounded-lg">
                {!useAvatarFallback && <AvatarImage src={avatarUrl} alt={user?.email} />}
                <AvatarFallback className="rounded-lg">
                  {user?.name?.charAt(0).toUpperCase() || user?.email?.charAt(0).toUpperCase()}
                </AvatarFallback>
              </Avatar>
              <div className="grid flex-1 text-left text-sm leading-tight">
                <span className="truncate text-xs">{user?.name || user?.email?.split('@')[0]}</span>
                <span className="truncate text-xs text-muted-foreground">{user?.email}</span>
              </div>
              <ChevronsUpDown className="ml-auto size-4" />
            </SidebarMenuButton>
          </DropdownMenuTrigger>
          <DropdownMenuContent
            className="w-[--radix-dropdown-menu-trigger-width] min-w-56 rounded-lg"
            side={isMobile ? 'bottom' : 'right'}
            align="end"
            sideOffset={4}
          >
            <DropdownMenuLabel className="p-0 font-normal">
              <div className="flex items-center gap-2 px-1 py-1.5 text-left text-sm">
                <Avatar className="h-8 w-8 rounded-lg">
                  {!useAvatarFallback && <AvatarImage src={avatarUrl} alt={user?.email} />}
                  <AvatarFallback className="rounded-lg">
                    {user?.name?.charAt(0).toUpperCase() || user?.email?.charAt(0).toUpperCase()}
                  </AvatarFallback>
                </Avatar>
                <div className="grid flex-1 text-left text-sm leading-tight">
                  <span className="truncate text-xs">
                    {user?.name || user?.email?.split('@')[0]}
                  </span>
                  <span className="truncate text-xs text-muted-foreground">{user?.email}</span>
                </div>
              </div>
            </DropdownMenuLabel>
            <DropdownMenuSeparator />
            <DropdownMenuItem onClick={() => router.push('/dashboard/settings')}>
              <Settings />
              Settings
            </DropdownMenuItem>
            <DropdownMenuItem onClick={logout}>
              <LogOut />
              Log out
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </SidebarMenuItem>
    </SidebarMenu>
  );
}
