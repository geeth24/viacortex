import React from 'react';
import { redirect } from 'next/navigation';

export default function Dashboard() {
  // Redirect to domains page by default
  redirect('/dashboard/domains');
} 