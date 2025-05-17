import React from 'react';
import { redirect } from 'next/navigation';

export default function SettingsProfile() {
  // Redirect to main profile page
  redirect('/profile');
} 