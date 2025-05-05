'use server';

import { cookies } from 'next/headers';
import { revalidatePath } from 'next/cache';
import type { Domain } from './domains/domains-table';
import type { BackendServer } from './servers/servers-table';

export async function createDomain(data: Partial<Domain>) {
  const cookieStore = await cookies();
  const token = cookieStore.get('token');

  const response = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/api/domains`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${token?.value}`,
    },
    body: JSON.stringify(data),
  });

  if (!response.ok) {
    throw new Error('Failed to create domain');
  }

  revalidatePath('/dashboard/domains/all');
  return response.json();
}

export async function updateDomain(id: number, data: Partial<Domain>) {
  const cookieStore = cookies();
  const token = (await cookieStore).get('token');

  const response = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/api/domains/${id}`, {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${token?.value}`,
    },
    body: JSON.stringify(data),
  });

  if (!response.ok) {
    throw new Error('Failed to update domain');
  }

  revalidatePath('/dashboard/domains/all');
  return response.json();
}

export async function createServer(domainId: number, data: Partial<BackendServer>) {
  const cookieStore = await cookies();
  const token = cookieStore.get('token');

  const response = await fetch(
    `${process.env.NEXT_PUBLIC_API_URL}/api/domains/${domainId}/backends`,
    {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${token?.value}`,
      },
      body: JSON.stringify(data),
    },
  );

  if (!response.ok) {
    throw new Error('Failed to create server');
  }

  revalidatePath('/dashboard/load-balancing/servers/[id]');
  return response.json();
}

export async function updateServer(
  domainId: number,
  serverId: number,
  data: Partial<BackendServer>,
) {
  const cookieStore = await cookies();
  const token = cookieStore.get('token');

  const response = await fetch(
    `${process.env.NEXT_PUBLIC_API_URL}/api/domains/${domainId}/backends/${serverId}`,
    {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${token?.value}`,
      },
      body: JSON.stringify(data),
    },
  );

  if (!response.ok) {
    throw new Error('Failed to update server');
  }

  revalidatePath('/dashboard/load-balancing/servers/[id]');
  return response.json();
}
