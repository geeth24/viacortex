'use client';
import { usePathname } from 'next/navigation';
import React from 'react'

function NotFoundPage() {
    const pathName = usePathname()
  return (
    <div className="mt-8">
        <h1 className="text-2xl">
            {pathName} not found
        </h1>
    </div>
  )
}

export default NotFoundPage