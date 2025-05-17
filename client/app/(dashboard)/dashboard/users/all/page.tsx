'use client'

import { useState, useEffect } from 'react'
import { Plus, Edit, Trash2, UserCog, Shield, Mail, Check, X } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Form, FormControl, FormDescription, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { Table, TableBody, TableCaption, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuLabel, DropdownMenuSeparator, DropdownMenuTrigger } from '@/components/ui/dropdown-menu'
import { useToast } from '@/components/ui/use-toast'
import { useForm } from 'react-hook-form'

interface User {
  id: string
  name: string
  email: string
  role: 'admin' | 'operator' | 'viewer'
  status: 'active' | 'inactive' | 'pending'
  createdAt: string
  lastLogin: string
}

export default function UsersPage() {
  const [users, setUsers] = useState<User[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [openDialog, setOpenDialog] = useState(false)
  const [openRoleDialog, setOpenRoleDialog] = useState(false)
  const [currentUser, setCurrentUser] = useState<User | null>(null)
  const { toast } = useToast()

  const form = useForm({
    defaultValues: {
      name: '',
      email: '',
      role: 'viewer',
      password: '',
    },
  })

  const roleForm = useForm({
    defaultValues: {
      role: 'viewer',
    },
  })

  useEffect(() => {
    fetchUsers()
  }, [])

  const fetchUsers = async () => {
    setIsLoading(true)
    try {
      const response = await fetch('/api/users', {
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
      })
      
      if (!response.ok) throw new Error('Failed to fetch users')
      
      const data = await response.json()
      setUsers(data.users || [])
    } catch (error) {
      toast({
        title: 'Error',
        description: 'Failed to load users. Please try again.',
        variant: 'destructive',
      })
    } finally {
      setIsLoading(false)
    }
  }

  const onSubmit = async (values: any) => {
    try {
      const method = currentUser ? 'PUT' : 'POST'
      const url = currentUser 
        ? `/api/users/${currentUser.id}` 
        : '/api/users'
      
      const response = await fetch(url, {
        method,
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
        body: JSON.stringify(values),
      })

      if (!response.ok) throw new Error('Failed to save user')
      
      toast({
        title: 'Success',
        description: currentUser ? 'User updated successfully' : 'User created successfully',
      })
      
      setOpenDialog(false)
      form.reset()
      fetchUsers()
    } catch (error) {
      toast({
        title: 'Error',
        description: 'Failed to save user. Please try again.',
        variant: 'destructive',
      })
    }
  }

  const onRoleSubmit = async (values: any) => {
    if (!currentUser) return

    try {
      const response = await fetch(`/api/users/${currentUser.id}/role`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
        body: JSON.stringify(values),
      })

      if (!response.ok) throw new Error('Failed to update user role')
      
      toast({
        title: 'Success',
        description: 'User role updated successfully',
      })
      
      setOpenRoleDialog(false)
      roleForm.reset()
      fetchUsers()
    } catch (error) {
      toast({
        title: 'Error',
        description: 'Failed to update user role. Please try again.',
        variant: 'destructive',
      })
    }
  }

  const handleDelete = async (id: string) => {
    if (!confirm('Are you sure you want to delete this user?')) return

    try {
      const response = await fetch(`/api/users/${id}`, {
        method: 'DELETE',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
      })

      if (!response.ok) throw new Error('Failed to delete user')
      
      toast({
        title: 'Success',
        description: 'User deleted successfully',
      })
      
      fetchUsers()
    } catch (error) {
      toast({
        title: 'Error',
        description: 'Failed to delete user. Please try again.',
        variant: 'destructive',
      })
    }
  }

  const handleEdit = (user: User) => {
    setCurrentUser(user)
    form.reset({
      name: user.name,
      email: user.email,
      role: user.role,
      password: '',
    })
    setOpenDialog(true)
  }

  const handleChangeRole = (user: User) => {
    setCurrentUser(user)
    roleForm.reset({
      role: user.role,
    })
    setOpenRoleDialog(true)
  }

  const handleAdd = () => {
    setCurrentUser(null)
    form.reset({
      name: '',
      email: '',
      role: 'viewer',
      password: '',
    })
    setOpenDialog(true)
  }

  const getRoleBadge = (role: string) => {
    switch (role) {
      case 'admin':
        return <Badge className="bg-purple-500">Admin</Badge>
      case 'operator':
        return <Badge className="bg-blue-500">Operator</Badge>
      case 'viewer':
        return <Badge className="bg-green-500">Viewer</Badge>
      default:
        return <Badge>{role}</Badge>
    }
  }

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'active':
        return <Badge className="bg-green-500">Active</Badge>
      case 'inactive':
        return <Badge className="bg-gray-500">Inactive</Badge>
      case 'pending':
        return <Badge className="bg-yellow-500">Pending</Badge>
      default:
        return <Badge>{status}</Badge>
    }
  }

  const getInitials = (name: string) => {
    return name
      .split(' ')
      .map(part => part[0])
      .join('')
      .toUpperCase()
      .substring(0, 2)
  }

  return (
    <div className="container mx-auto p-6">
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-3xl font-bold">Users</h1>
        <Button onClick={handleAdd}>
          <Plus className="mr-2 h-4 w-4" />
          Add User
        </Button>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>All Users</CardTitle>
          <CardDescription>Manage user accounts and access privileges</CardDescription>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="flex justify-center p-6">Loading users...</div>
          ) : users.length === 0 ? (
            <Alert>
              <AlertTitle>No users found</AlertTitle>
              <AlertDescription>
                There are no users in the system yet. Click the "Add User" button to create your first user.
              </AlertDescription>
            </Alert>
          ) : (
            <Table>
              <TableCaption>A list of all users in the system.</TableCaption>
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
                          <AvatarFallback>{getInitials(user.name)}</AvatarFallback>
                        </Avatar>
                        <div>
                          <p className="font-medium">{user.name}</p>
                          <p className="text-xs text-muted-foreground">
                            Created {new Date(user.createdAt).toLocaleDateString()}
                          </p>
                        </div>
                      </div>
                    </TableCell>
                    <TableCell>{user.email}</TableCell>
                    <TableCell>{getRoleBadge(user.role)}</TableCell>
                    <TableCell>{getStatusBadge(user.status)}</TableCell>
                    <TableCell>
                      {user.lastLogin 
                        ? new Date(user.lastLogin).toLocaleString() 
                        : 'Never'}
                    </TableCell>
                    <TableCell className="text-right">
                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <Button variant="ghost" size="sm">
                            <UserCog className="h-4 w-4" />
                          </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end">
                          <DropdownMenuLabel>Actions</DropdownMenuLabel>
                          <DropdownMenuSeparator />
                          <DropdownMenuItem onClick={() => handleEdit(user)}>
                            <Edit className="mr-2 h-4 w-4" />
                            Edit User
                          </DropdownMenuItem>
                          <DropdownMenuItem onClick={() => handleChangeRole(user)}>
                            <Shield className="mr-2 h-4 w-4" />
                            Change Role
                          </DropdownMenuItem>
                          <DropdownMenuItem onClick={() => handleDelete(user.id)}>
                            <Trash2 className="mr-2 h-4 w-4" />
                            Delete User
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
            <DialogTitle>{currentUser ? 'Edit User' : 'Add New User'}</DialogTitle>
            <DialogDescription>
              {currentUser 
                ? 'Update the user details below.' 
                : 'Enter the details for the new user.'}
            </DialogDescription>
          </DialogHeader>
          <Form {...form}>
            <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
              <FormField
                control={form.control}
                name="name"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Full Name</FormLabel>
                    <FormControl>
                      <Input placeholder="John Doe" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              
              <FormField
                control={form.control}
                name="email"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Email</FormLabel>
                    <FormControl>
                      <Input type="email" placeholder="john@example.com" {...field} />
                    </FormControl>
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
                    <Select 
                      onValueChange={field.onChange} 
                      defaultValue={field.value}
                    >
                      <FormControl>
                        <SelectTrigger>
                          <SelectValue placeholder="Select a role" />
                        </SelectTrigger>
                      </FormControl>
                      <SelectContent>
                        <SelectItem value="admin">Admin</SelectItem>
                        <SelectItem value="operator">Operator</SelectItem>
                        <SelectItem value="viewer">Viewer</SelectItem>
                      </SelectContent>
                    </Select>
                    <FormDescription>
                      {field.value === 'admin' 
                        ? 'Full access to all settings and operations' 
                        : field.value === 'operator' 
                        ? 'Can manage domains and configurations' 
                        : 'View-only access to dashboards and reports'}
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />
              
              <FormField
                control={form.control}
                name="password"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{currentUser ? 'New Password (optional)' : 'Password'}</FormLabel>
                    <FormControl>
                      <Input type="password" {...field} />
                    </FormControl>
                    <FormDescription>
                      {currentUser 
                        ? 'Leave blank to keep current password' 
                        : 'Minimum 8 characters with mixed case, numbers, and symbols'}
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />
              
              <DialogFooter>
                <Button type="submit">
                  {currentUser ? 'Update User' : 'Add User'}
                </Button>
              </DialogFooter>
            </form>
          </Form>
        </DialogContent>
      </Dialog>

      <Dialog open={openRoleDialog} onOpenChange={setOpenRoleDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Change User Role</DialogTitle>
            <DialogDescription>
              Update the role and permissions for {currentUser?.name}
            </DialogDescription>
          </DialogHeader>
          <Form {...roleForm}>
            <form onSubmit={roleForm.handleSubmit(onRoleSubmit)} className="space-y-4">
              <FormField
                control={roleForm.control}
                name="role"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Role</FormLabel>
                    <Select 
                      onValueChange={field.onChange} 
                      defaultValue={field.value}
                    >
                      <FormControl>
                        <SelectTrigger>
                          <SelectValue placeholder="Select a role" />
                        </SelectTrigger>
                      </FormControl>
                      <SelectContent>
                        <SelectItem value="admin">Admin</SelectItem>
                        <SelectItem value="operator">Operator</SelectItem>
                        <SelectItem value="viewer">Viewer</SelectItem>
                      </SelectContent>
                    </Select>
                    <div className="pt-4 space-y-2">
                      <div className="flex items-center justify-between border-b pb-2">
                        <span className="text-sm">View dashboard and metrics</span>
                        <Check className="h-4 w-4 text-green-500" />
                      </div>
                      <div className="flex items-center justify-between border-b pb-2">
                        <span className="text-sm">Manage domains and backends</span>
                        {field.value === 'viewer' 
                          ? <X className="h-4 w-4 text-red-500" /> 
                          : <Check className="h-4 w-4 text-green-500" />}
                      </div>
                      <div className="flex items-center justify-between border-b pb-2">
                        <span className="text-sm">Manage security settings</span>
                        {field.value === 'viewer' 
                          ? <X className="h-4 w-4 text-red-500" /> 
                          : <Check className="h-4 w-4 text-green-500" />}
                      </div>
                      <div className="flex items-center justify-between border-b pb-2">
                        <span className="text-sm">Manage users</span>
                        {field.value === 'admin' 
                          ? <Check className="h-4 w-4 text-green-500" /> 
                          : <X className="h-4 w-4 text-red-500" />}
                      </div>
                      <div className="flex items-center justify-between">
                        <span className="text-sm">System configuration</span>
                        {field.value === 'admin' 
                          ? <Check className="h-4 w-4 text-green-500" /> 
                          : <X className="h-4 w-4 text-red-500" />}
                      </div>
                    </div>
                    <FormMessage />
                  </FormItem>
                )}
              />
              
              <DialogFooter>
                <Button type="submit">
                  Update Role
                </Button>
              </DialogFooter>
            </form>
          </Form>
        </DialogContent>
      </Dialog>
    </div>
  )
} 