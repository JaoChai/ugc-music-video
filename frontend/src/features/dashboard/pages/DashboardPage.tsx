import { Link } from 'react-router-dom'
import { Briefcase, CheckCircle2, Loader2, AlertCircle, Plus, ArrowRight } from 'lucide-react'
import { Card, CardHeader, CardContent, Button } from '@/components/ui'
import { useAuthStore } from '@/stores/auth.store'
import { DashboardLayout, StatsCard, RecentJobsList } from '../components'
import { useJobStatsQuery, useRecentJobsQuery } from '../hooks'

export default function DashboardPage() {
  const { user, isAuthenticated } = useAuthStore()
  const { data: stats, isLoading: isLoadingStats } = useJobStatsQuery()
  const { data: recentJobs, isLoading: isLoadingJobs } = useRecentJobsQuery(5)

  // Get greeting based on time of day
  const getGreeting = () => {
    const hour = new Date().getHours()
    if (hour < 12) return 'สวัสดีตอนเช้า'
    if (hour < 18) return 'สวัสดีตอนบ่าย'
    return 'สวัสดีตอนเย็น'
  }

  return (
    <DashboardLayout>
      {/* Header Section */}
      <div className="mb-8">
        <h1 className="text-2xl sm:text-3xl font-bold text-gray-900">
          {getGreeting()}{isAuthenticated && user?.name ? `, ${user.name}` : ''}!
        </h1>
        <p className="text-gray-600 mt-1">
          ภาพรวมงานสร้างวิดีโอ UGC ของคุณ
        </p>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 sm:gap-6 mb-8">
        <StatsCard
          icon={Briefcase}
          title="งานทั้งหมด"
          value={stats?.total ?? 0}
          variant="gray"
          isLoading={isLoadingStats}
        />
        <StatsCard
          icon={CheckCircle2}
          title="เสร็จสิ้น"
          value={stats?.completed ?? 0}
          variant="green"
          isLoading={isLoadingStats}
        />
        <StatsCard
          icon={Loader2}
          title="กำลังดำเนินการ"
          value={stats?.inProgress ?? 0}
          variant="yellow"
          isLoading={isLoadingStats}
        />
        <StatsCard
          icon={AlertCircle}
          title="ล้มเหลว"
          value={stats?.failed ?? 0}
          variant="red"
          isLoading={isLoadingStats}
        />
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Recent Jobs */}
        <div className="lg:col-span-2">
          <Card>
            <CardHeader>
              <div className="flex items-center justify-between">
                <h2 className="text-lg font-semibold text-gray-900">งานล่าสุด</h2>
                <Link
                  to="/jobs"
                  className="text-sm text-zinc-700 hover:text-zinc-900 font-medium flex items-center gap-1"
                >
                  ดูทั้งหมด
                  <ArrowRight className="h-4 w-4" />
                </Link>
              </div>
            </CardHeader>
            <CardContent className="p-0">
              <RecentJobsList jobs={recentJobs || []} isLoading={isLoadingJobs} />
            </CardContent>
          </Card>
        </div>

        {/* Quick Actions */}
        <div className="lg:col-span-1">
          <Card>
            <CardHeader>
              <h2 className="text-lg font-semibold text-gray-900">ดำเนินการด่วน</h2>
            </CardHeader>
            <CardContent>
              <div className="space-y-3">
                <Link to="/jobs/create" className="block">
                  <Button className="w-full justify-start" size="lg">
                    <Plus className="h-5 w-5 mr-3" />
                    สร้างงานใหม่
                  </Button>
                </Link>
                <Link to="/jobs" className="block">
                  <Button variant="outline" className="w-full justify-start" size="lg">
                    <Briefcase className="h-5 w-5 mr-3" />
                    ดูงานทั้งหมด
                  </Button>
                </Link>
              </div>

              {/* Tips Section */}
              <div className="mt-6 p-4 bg-zinc-100 rounded-lg">
                <h3 className="text-sm font-semibold text-zinc-900 mb-2">เคล็ดลับ</h3>
                <p className="text-sm text-zinc-700">
                  ใส่รายละเอียด concept ให้ชัดเจนเพื่อผลลัพธ์ที่ดีกว่า รวมถึงอารมณ์ สไตล์ และกลุ่มเป้าหมาย
                </p>
              </div>
            </CardContent>
          </Card>

          {/* Stats Summary Card */}
          {!isLoadingStats && stats && stats.total > 0 && (
            <Card className="mt-6">
              <CardContent className="pt-6">
                <h3 className="text-sm font-medium text-gray-500 mb-4">อัตราความสำเร็จ</h3>
                <div className="flex items-end gap-2">
                  <span className="text-3xl font-bold text-gray-900">
                    {stats.total > 0 ? Math.round((stats.completed / stats.total) * 100) : 0}%
                  </span>
                  <span className="text-sm text-gray-500 mb-1">อัตราเสร็จสิ้น</span>
                </div>
                <div className="mt-4 h-2 bg-gray-200 rounded-full overflow-hidden">
                  <div
                    className="h-full bg-green-500 rounded-full transition-all duration-500"
                    style={{ width: `${stats.total > 0 ? (stats.completed / stats.total) * 100 : 0}%` }}
                  />
                </div>
                <div className="mt-3 flex justify-between text-xs text-gray-500">
                  <span>{stats.completed} เสร็จสิ้น</span>
                  <span>{stats.total} ทั้งหมด</span>
                </div>
              </CardContent>
            </Card>
          )}
        </div>
      </div>
    </DashboardLayout>
  )
}
