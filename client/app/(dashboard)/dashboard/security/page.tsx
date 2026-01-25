import { redirect } from 'next/navigation'

export default function SecurityPage() {
  redirect('/dashboard/security/ip-rules')
}
