import { ModeToggle } from '@/components/mode-toggle';
import Logo from '@/components/logo';
import { Button } from '@/components/ui/button';
import Link from 'next/link';

export default function Home() {
  return (
    <div className="flex h-screen flex-col items-center justify-center text-foreground">
      <div className="absolute right-4 top-4">
        <ModeToggle />
      </div>
      <main className="space-y-6 px-4 text-center">
        <div className="animate-float">
          <Logo className="mx-auto h-24 w-24 text-primary" />
        </div>
        <h1 className="font-josefin text-5xl font-bold tracking-tight">ViaCortex</h1>
        <p className="mx-auto max-w-md text-xl text-muted-foreground">
          Modern Reverse Proxy and Load Balancer for Seamless Traffic Management
        </p>
        <Button asChild size="lg" className="mt-8">
          <Link href="/dashboard">Get Started</Link>
        </Button>
      </main>
      <footer className="absolute bottom-4 text-sm text-muted-foreground">
        © 2024 ViaCortex. All rights reserved.
      </footer>
    </div>
  );
}
