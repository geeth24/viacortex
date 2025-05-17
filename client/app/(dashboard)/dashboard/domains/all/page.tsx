import React from 'react';
import { redirect } from 'next/navigation';

export default function AllDomains() {
  // Redirect to main domains page
  redirect('/dashboard/domains');
} 