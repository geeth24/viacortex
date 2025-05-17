'use client'

import { useState, useEffect } from 'react'
import { BarChart, Bar, LineChart, Line, AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts'
import { BarChart as BarChartIcon, PieChart, Activity, Clock, Users, Download, ArrowUpDown, RefreshCcw } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Chart } from '@/components/ui/chart'
import { useToast } from '@/components/ui/use-toast'

const timeRangeOptions = [
  { value: '1h', label: 'Last Hour' },
  { value: '24h', label: 'Last 24 Hours' },
  { value: '7d', label: 'Last 7 Days' },
  { value: '30d', label: 'Last 30 Days' },
]

interface MetricDataPoint {
  timestamp: string
  requests: number
  errors: number
  bandwidth: number
  avgResponseTime: number
  uniqueVisitors: number
}

interface TotalStats {
  totalRequests: number
  totalErrors: number
  totalBandwidth: string
  avgResponseTime: number
  uniqueVisitors: number
  successRate: number
}

export default function GlobalMetricsPage() {
  const [timeRange, setTimeRange] = useState('24h')
  const [isLoading, setIsLoading] = useState(true)
  const [metricsData, setMetricsData] = useState<MetricDataPoint[]>([])
  const [totalStats, setTotalStats] = useState<TotalStats>({
    totalRequests: 0,
    totalErrors: 0,
    totalBandwidth: '0 MB',
    avgResponseTime: 0,
    uniqueVisitors: 0,
    successRate: 0,
  })
  const { toast } = useToast()

  useEffect(() => {
    fetchMetrics(timeRange)
  }, [timeRange])

  const fetchMetrics = async (range: string) => {
    setIsLoading(true)
    try {
      const response = await fetch(`/api/metrics?timeRange=${range}`, {
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
      })
      
      if (!response.ok) throw new Error('Failed to fetch metrics')
      
      const data = await response.json()
      
      setMetricsData(data.timeSeries || [])
      setTotalStats(data.totals || {
        totalRequests: 0,
        totalErrors: 0,
        totalBandwidth: '0 MB',
        avgResponseTime: 0,
        uniqueVisitors: 0,
        successRate: 0,
      })
    } catch (error) {
      toast({
        title: 'Error',
        description: 'Failed to load metrics. Please try again.',
        variant: 'destructive',
      })
    } finally {
      setIsLoading(false)
    }
  }

  const handleRefresh = () => {
    fetchMetrics(timeRange)
  }

  const formatTimestamp = (timestamp: string) => {
    const date = new Date(timestamp)
    if (timeRange === '1h') {
      return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
    } else if (timeRange === '24h') {
      return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
    } else {
      return date.toLocaleDateString([], { month: 'short', day: 'numeric' })
    }
  }

  return (
    <div className="container mx-auto p-6">
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-3xl font-bold">Global Metrics</h1>
        <div className="flex gap-4">
          <Select 
            value={timeRange} 
            onValueChange={setTimeRange}
          >
            <SelectTrigger className="w-[180px]">
              <SelectValue placeholder="Select time range" />
            </SelectTrigger>
            <SelectContent>
              {timeRangeOptions.map((option) => (
                <SelectItem key={option.value} value={option.value}>{option.label}</SelectItem>
              ))}
            </SelectContent>
          </Select>
          <Button onClick={handleRefresh} variant="outline">
            <RefreshCcw className="mr-2 h-4 w-4" />
            Refresh
          </Button>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 mb-8">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">Total Requests</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex items-center">
              <div className="mr-2 rounded-md bg-primary/10 p-2">
                <Activity className="h-5 w-5 text-primary" />
              </div>
              <div>
                <div className="text-2xl font-bold">{totalStats.totalRequests.toLocaleString()}</div>
                <p className="text-xs text-muted-foreground">
                  Success Rate: {totalStats.successRate.toFixed(2)}%
                </p>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">Average Response Time</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex items-center">
              <div className="mr-2 rounded-md bg-primary/10 p-2">
                <Clock className="h-5 w-5 text-primary" />
              </div>
              <div>
                <div className="text-2xl font-bold">{totalStats.avgResponseTime.toFixed(2)} ms</div>
                <p className="text-xs text-muted-foreground">
                  {timeRange === '24h' ? 'Past 24 hours' : timeRange === '7d' ? 'Past 7 days' : 'Past 30 days'}
                </p>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">Unique Visitors</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex items-center">
              <div className="mr-2 rounded-md bg-primary/10 p-2">
                <Users className="h-5 w-5 text-primary" />
              </div>
              <div>
                <div className="text-2xl font-bold">{totalStats.uniqueVisitors.toLocaleString()}</div>
                <p className="text-xs text-muted-foreground">
                  {timeRange === '24h' ? 'Past 24 hours' : timeRange === '7d' ? 'Past 7 days' : 'Past 30 days'}
                </p>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">Total Bandwidth</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex items-center">
              <div className="mr-2 rounded-md bg-primary/10 p-2">
                <Download className="h-5 w-5 text-primary" />
              </div>
              <div>
                <div className="text-2xl font-bold">{totalStats.totalBandwidth}</div>
                <p className="text-xs text-muted-foreground">
                  {timeRange === '24h' ? 'Past 24 hours' : timeRange === '7d' ? 'Past 7 days' : 'Past 30 days'}
                </p>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">Error Rate</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex items-center">
              <div className="mr-2 rounded-md bg-primary/10 p-2">
                <ArrowUpDown className="h-5 w-5 text-primary" />
              </div>
              <div>
                <div className="text-2xl font-bold">{(100 - totalStats.successRate).toFixed(2)}%</div>
                <p className="text-xs text-muted-foreground">
                  Total Errors: {totalStats.totalErrors.toLocaleString()}
                </p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      <Tabs defaultValue="traffic">
        <TabsList className="mb-4">
          <TabsTrigger value="traffic">Traffic</TabsTrigger>
          <TabsTrigger value="performance">Performance</TabsTrigger>
          <TabsTrigger value="errors">Errors</TabsTrigger>
        </TabsList>

        <TabsContent value="traffic">
          <Card>
            <CardHeader>
              <CardTitle>Traffic Overview</CardTitle>
              <CardDescription>
                Request volume and unique visitors over time
              </CardDescription>
            </CardHeader>
            <CardContent>
              {isLoading ? (
                <div className="flex justify-center items-center h-80">Loading chart data...</div>
              ) : (
                <div className="h-80">
                  <Chart 
                    options={{
                      chart: {
                        type: 'area',
                        toolbar: {
                          show: false,
                        },
                        zoom: {
                          enabled: false,
                        },
                      },
                      dataLabels: {
                        enabled: false,
                      },
                      stroke: {
                        curve: 'smooth',
                        width: 2,
                      },
                      fill: {
                        type: 'gradient',
                        gradient: {
                          shadeIntensity: 1,
                          opacityFrom: 0.7,
                          opacityTo: 0.3,
                          stops: [0, 90, 100],
                        },
                      },
                      xaxis: {
                        categories: metricsData.map(d => formatTimestamp(d.timestamp)),
                      },
                      yaxis: {
                        title: {
                          text: 'Count',
                        },
                      },
                      tooltip: {
                        shared: true,
                      },
                    }}
                    series={[
                      {
                        name: 'Requests',
                        data: metricsData.map(d => d.requests),
                      },
                      {
                        name: 'Unique Visitors',
                        data: metricsData.map(d => d.uniqueVisitors),
                      },
                    ]}
                    type="area"
                    height={350}
                  />
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="performance">
          <Card>
            <CardHeader>
              <CardTitle>Performance Metrics</CardTitle>
              <CardDescription>
                Average response time and bandwidth usage
              </CardDescription>
            </CardHeader>
            <CardContent>
              {isLoading ? (
                <div className="flex justify-center items-center h-80">Loading chart data...</div>
              ) : (
                <div className="h-80">
                  <Chart 
                    options={{
                      chart: {
                        type: 'line',
                        toolbar: {
                          show: false,
                        },
                        zoom: {
                          enabled: false,
                        },
                      },
                      dataLabels: {
                        enabled: false,
                      },
                      stroke: {
                        curve: 'smooth',
                        width: 2,
                      },
                      xaxis: {
                        categories: metricsData.map(d => formatTimestamp(d.timestamp)),
                      },
                      yaxis: [
                        {
                          title: {
                            text: 'Response Time (ms)',
                          },
                        },
                        {
                          opposite: true,
                          title: {
                            text: 'Bandwidth (MB)',
                          },
                        },
                      ],
                      tooltip: {
                        shared: true,
                      },
                    }}
                    series={[
                      {
                        name: 'Avg Response Time (ms)',
                        data: metricsData.map(d => d.avgResponseTime),
                      },
                      {
                        name: 'Bandwidth (MB)',
                        data: metricsData.map(d => d.bandwidth),
                      },
                    ]}
                    type="line"
                    height={350}
                  />
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="errors">
          <Card>
            <CardHeader>
              <CardTitle>Error Rate</CardTitle>
              <CardDescription>
                Successful vs. error requests over time
              </CardDescription>
            </CardHeader>
            <CardContent>
              {isLoading ? (
                <div className="flex justify-center items-center h-80">Loading chart data...</div>
              ) : (
                <div className="h-80">
                  <Chart 
                    options={{
                      chart: {
                        type: 'bar',
                        stacked: true,
                        toolbar: {
                          show: false,
                        },
                        zoom: {
                          enabled: false,
                        },
                      },
                      plotOptions: {
                        bar: {
                          horizontal: false,
                        },
                      },
                      dataLabels: {
                        enabled: false,
                      },
                      xaxis: {
                        categories: metricsData.map(d => formatTimestamp(d.timestamp)),
                      },
                      colors: ['#10b981', '#ef4444'],
                      tooltip: {
                        shared: true,
                      },
                    }}
                    series={[
                      {
                        name: 'Successful Requests',
                        data: metricsData.map(d => d.requests - d.errors),
                      },
                      {
                        name: 'Error Requests',
                        data: metricsData.map(d => d.errors),
                      },
                    ]}
                    type="bar"
                    height={350}
                  />
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  )
} 