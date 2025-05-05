import { cookies } from 'next/headers';
import { DomainsTable } from '@/components/domains/domains-table';

async function getDomains() {
  const cookieStore = cookies();
  const token = (await cookieStore).get('token');

  const response = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/api/domains`, {
    headers: {
      Authorization: `Bearer ${token?.value}`,
    },
  });
  if (!response.ok) {
    throw new Error('Failed to fetch domains');
  }
  return response.json();
}

export default async function Page() {
  const domains = await getDomains();

  return (
    <div className="container mx-auto py-10">
      <DomainsTable domains={domains} />
    </div>
  );
}
