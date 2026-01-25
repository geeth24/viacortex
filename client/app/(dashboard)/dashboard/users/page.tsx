'use client'

import { useState, useEffect } from 'react'
import { Plus, Edit, Trash2, MoreHorizontal } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Form, FormControl, FormDescription, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Avatar, AvatarFallback } from '@/components/ui/avatar'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from '@/components/ui/dropdown-menu'
import { Switch } from '@/components/ui/switch'
import { toast } from 'sonner'
import { useForm } from 'react-hook-form'
import { User } from '@/types/domains'

interface UserFormValues {
  email: string
  name: string
  password: string
  role: 'admin' | 'user'
  active: boolean
}

export default function UsersPage() {
  const [users, setUsers] = useState<User[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [openDialog, setOpenDialog] = useState(false)
  const [currentUser, setCurrentUser] = useState<User | null>(null)

  const form = useForm<UserFormValues>({
    defaultValues: {
      email: '',
      name: '',
      password: '',
      role: 'user',
      active: true,
    },
  })

  useEffect(() => {
    fetchUsers()
  }, [])

  const fetchUsers = async () => {
    setIsLoading(true)
    try {
      const response = await fetch('/api/users', {
        headers: { 'Content-Type': 'application/json' },
      })

      if (!response.ok) throw new Error('Failed to fetch users')

      const data = await response.json()
      setUsers(Array.isArray(data) ? data : [])
    } catch (error) {
      console.error(error)
      toast.error('Failed to load users')
    } finally {
      setIsLoading(false)
    }
  }

  const onSubmit = async (values: UserFormValues) => {
    try {
      const method = currentUser ? 'PUT' : 'POST'
      const url = currentUser ? `/api/users/${currentUser.id}` : '/api/users'

      const payload: Record<string, unknown> = { ...values }
      if (currentUser && !values.password) {
        delete payload.password
      }

      const response = await fetch(url, {
        method,
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
      })

      if (!response.ok) throw new Error('Failed to save user')

      toast.success(currentUser ? 'User updated' : 'User created')
      setOpenDialog(false)
      form.reset()
      setCurrentUser(null)
      fetchUsers()
    } catch (error) {
      console.error(error)
      toast.error('Failed to save user')
    }
  }

  const handleDelete = async (id: number) => {
    if (!confirm('Delete this user?')) return

    try {
      const response = await fetch(`/api/users/${id}`, {
        method: 'DELETE',
        headers: { 'Content-Type': 'application/json' },
      })

      if (!response.ok) throw new Error('Failed to delete user')

      toast.success('User deleted')
      fetchUsers()
    } catch (error) {
      console.error(error)
      toast.error('Failed to delete user')
    }
  }

  const handleEdit = (user: User) => {
    setCurrentUser(user)
    form.reset({
      email: user.email,
      name: user.name || '',
      password: '',
      role: user.role,
      active: user.active,
    })
    setOpenDialog(true)
  }

  const handleAdd = () => {
    setCurrentUser(null)
    form.reset({
      email: '',
      name: '',
      password: '',
      role: 'user',
      active: true,
    })
    setOpenDialog(true)
  }

  const getInitials = (name: string, email: string) => {
    if (name) {
      return name.split(' ').map(p => p[0]).join('').toUpperCase().substring(0, 2)
    }
    return email.substring(0, 2).toUpperCase()
  }

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-3xl font-bold">Users</h1>
          <p className="text-muted-foreground">Manage user accounts</p>
        </div>
        <Button onClick={handleAdd}>
          <Plus className="mr-2 h-4 w-4" />
          Add User
        </Button>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>All Users</CardTitle>
          <CardDescription>System users and their access levels</CardDescription>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="flex justify-center p-6">Loading...</div>
          ) : users.length === 0 ? (
            <Alert>
              <AlertTitle>No users</AlertTitle>
              <AlertDescription>No users found in the system.</AlertDescription>
            </Alert>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>User</TableHead>
                  <TableHead>Email</TableHead>
                  <TableHead>Role</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Last Login</TableHead>
                  <TableHead className="text-right">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {users.map((user) => (
                  <TableRow key={user.id}>
                    <TableCell>
                      <div className="flex items-center gap-3">
                        <Avatar>
                          <AvatarFallback>{getInitials(user.name, user.email)}</AvatarFallback>
                        </Avatar>
                        <span className="font-medium">{user.name || user.email}</span>
                      </div>
                    </TableCell>
                    <TableCell>{user.email}</TableCell>
                    <TableCell>
                      {user.role === 'admin' ? (
                        <Badge className="bg-purple-500">Admin</Badge>
                      ) : (
                        <Badge variant="secondary">User</Badge>
                      )}
                    </TableCell>
                    <TableCell>
                      {user.active ? (
                        <Badge className="bg-green-500">Active</Badge>
                      ) : (
                        <Badge variant="secondary">Inactive</Badge>
                      )}
                    </TableCell>
                    <TableCell>
                      {user.last_login
                        ? new Date(user.last_login).toLocaleString()
                        : 'Never'}
                    </TableCell>
                    <TableCell className="text-right">
                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <Button variant="ghost" size="sm">
                            <MoreHorizontal className="h-4 w-4" />
                          </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end">
                          <DropdownMenuItem onClick={() => handleEdit(user)}>
                            <Edit className="mr-2 h-4 w-4" />
                            Edit
                          </DropdownMenuItem>
                          <DropdownMenuItem
                            onClick={() => handleDelete(user.id)}
                            className="text-destructive"
                          >
                            <Trash2 className="mr-2 h-4 w-4" />
                            Delete
                          </DropdownMenuItem>
                        </DropdownMenuContent>
                      </DropdownMenu>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

      <Dialog open={openDialog} onOpenChange={setOpenDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{currentUser ? 'Edit User' : 'Add User'}</DialogTitle>
            <DialogDescription>
              {currentUser ? 'Update user details.' : 'Create a new user account.'}
            </DialogDescription>
          </DialogHeader>
          <Form {...form}>
            <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
              <FormField
                control={form.control}
                name="email"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Email</FormLabel>
                    <FormControl>
                      <Input type="email" placeholder="user@example.com" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="name"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Name</FormLabel>
                    <FormControl>
                      <Input placeholder="John Doe" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="password"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{currentUser ? 'New Password' : 'Password'}</FormLabel>
                    <FormControl>
                      <Input type="password" {...field} />
                    </FormControl>
                    {currentUser && (
                      <FormDescription>Leave blank to keep current password</FormDescription>
                    )}
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="role"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Role</FormLabel>
                    <Select onValueChange={field.onChange} defaultValue={field.value}>
                      <FormControl>
                        <SelectTrigger>
                          <SelectValue />
                        </SelectTrigger>
                      </FormControl>
                      <SelectContent>
                        <SelectItem value="admin">Admin</SelectItem>
                        <SelectItem value="user">User</SelectItem>
                      </SelectContent>
                    </Select>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="active"
                render={({ field }) => (
                  <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3">
                    <div className="space-y-0.5">
                      <FormLabel>Active</FormLabel>
                      <FormDescription>User can log in</FormDescription>
                    </div>
                    <FormControl>
                      <Switch checked={field.value} onCheckedChange={field.onChange} />
                    </FormControl>
                  </FormItem>
                )}
              />

              <DialogFooter>
                <Button type="button" variant="outline" onClick={() => setOpenDialog(false)}>
                  Cancel
                </Button>
                <Button type="submit">{currentUser ? 'Update' : 'Create'}</Button>
              </DialogFooter>
            </form>
          </Form>
        </DialogContent>
      </Dialog>
    </div>
  )
}
