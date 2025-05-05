import React from 'react'
import Logo from './logo'
import Link from 'next/link'

function Navbar() {
  return (
    <nav
      className='fixed top-0 left-0 right-0 z-50 flex justify-between items-center p-4 bg-background border-b border-primary/50'
    >
      <div className='flex items-center gap-2'>
        <Logo className='text-primary' />
        <p className='text-2xl font-bold font-josefin'>ViaCortex</p>
      </div>

      <div className='flex items-center gap-2'>
        <Link href='/'>Home</Link>
        <Link href='/docs'>Docs</Link>
        <Link href='/pricing'>Pricing</Link>
        <Link href='/login'>Login</Link>
        <Link href='/register'>Register</Link>
        <Link href='/status'>Status</Link>
      </div>
    </nav>
  )
}

export default Navbar