import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { Card, CardHeader, CardContent, CardFooter, Button, Input } from '@/components/ui'
import { ArrowLeft, Sparkles } from 'lucide-react'
import { useCreateJob } from '../api'

export default function CreateJobPage() {
  const navigate = useNavigate()
  const createJob = useCreateJob()

  const [concept, setConcept] = useState('')
  const [model, setModel] = useState('')
  const [errors, setErrors] = useState<{ concept?: string }>({})

  const validateForm = () => {
    const newErrors: { concept?: string } = {}

    if (!concept.trim()) {
      newErrors.concept = 'กรุณากรอก Concept'
    } else if (concept.trim().length < 5) {
      newErrors.concept = 'Concept ต้องมีอย่างน้อย 5 ตัวอักษร'
    }

    setErrors(newErrors)
    return Object.keys(newErrors).length === 0
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    if (!validateForm()) return

    try {
      const job = await createJob.mutateAsync({
        concept: concept.trim(),
        model: model.trim() || undefined,
      })
      navigate(`/jobs/${job.id}`)
    } catch (error) {
      console.error('Failed to create job:', error)
    }
  }

  return (
    <div className="min-h-screen bg-gray-100">
      {/* Header */}
      <div className="bg-white shadow-sm border-b border-gray-200">
        <div className="max-w-3xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
          <div className="flex items-center gap-4">
            <Link to="/jobs" className="text-gray-600 hover:text-gray-900">
              <ArrowLeft className="h-5 w-5" />
            </Link>
            <h1 className="text-2xl font-bold text-gray-900">สร้างงานใหม่</h1>
          </div>
        </div>
      </div>

      {/* Main Content */}
      <main className="max-w-3xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <form onSubmit={handleSubmit}>
          <Card>
            <CardHeader>
              <div className="flex items-center gap-3">
                <div className="bg-zinc-100 p-2 rounded-lg">
                  <Sparkles className="h-5 w-5 text-zinc-600" />
                </div>
                <div>
                  <h2 className="text-lg font-semibold text-gray-900">สร้างวิดีโอ</h2>
                  <p className="text-sm text-gray-500">
                    อธิบาย concept วิดีโอของคุณ แล้วเราจะสร้างเพลง ภาพ และวิดีโอให้คุณ
                  </p>
                </div>
              </div>
            </CardHeader>

            <CardContent className="space-y-6">
              {/* Concept */}
              <div>
                <label htmlFor="concept" className="block text-sm font-medium text-gray-700 mb-1">
                  Concept <span className="text-red-500">*</span>
                </label>
                <textarea
                  id="concept"
                  value={concept}
                  onChange={(e) => {
                    setConcept(e.target.value)
                    if (errors.concept) setErrors({ ...errors, concept: undefined })
                  }}
                  placeholder="อธิบาย concept วิดีโอของคุณ... (เช่น พระอาทิตย์ตกริมทะเลที่สงบผ่อนคลาย)"
                  rows={4}
                  className={`block w-full rounded-lg border px-4 py-3 text-gray-900 placeholder-gray-500
                    focus:border-zinc-500 focus:ring-2 focus:ring-zinc-500 focus:ring-opacity-50 focus:outline-none
                    resize-none ${errors.concept ? 'border-red-500' : 'border-gray-300'}`}
                />
                {errors.concept && <p className="mt-1 text-sm text-red-600">{errors.concept}</p>}
                <p className="mt-1 text-xs text-gray-500">
                  {concept.length} ตัวอักษร (ขั้นต่ำ 5)
                </p>
              </div>

              {/* Model */}
              <div>
                <label htmlFor="model" className="block text-sm font-medium text-gray-700 mb-1">
                  โมเดล (ไม่บังคับ)
                </label>
                <Input
                  id="model"
                  value={model}
                  onChange={(e) => setModel(e.target.value)}
                  placeholder="anthropic/claude-3.5-sonnet"
                />
                <p className="mt-1 text-xs text-gray-500">
                  เว้นว่างเพื่อใช้โมเดลเริ่มต้น
                </p>
              </div>

              {/* Preview section */}
              {concept.trim().length >= 5 && (
                <div className="border border-gray-200 rounded-lg p-4 bg-gray-50">
                  <h3 className="text-sm font-medium text-gray-700 mb-2">ตัวอย่าง</h3>
                  <div className="space-y-2">
                    <div>
                      <span className="text-xs text-gray-500">Concept:</span>
                      <p className="text-sm text-gray-900">{concept}</p>
                    </div>
                    {model && (
                      <div>
                        <span className="text-xs text-gray-500">โมเดล:</span>
                        <p className="text-sm text-gray-900 font-mono">{model}</p>
                      </div>
                    )}
                  </div>
                </div>
              )}

              {/* Error message */}
              {createJob.isError && (
                <div className="bg-red-50 border border-red-200 rounded-lg p-4">
                  <p className="text-sm text-red-600">
                    ไม่สามารถสร้างงานได้ กรุณาลองอีกครั้ง
                  </p>
                </div>
              )}
            </CardContent>

            <CardFooter className="flex justify-end gap-3">
              <Link to="/jobs">
                <Button type="button" variant="outline">
                  ยกเลิก
                </Button>
              </Link>
              <Button type="submit" isLoading={createJob.isPending}>
                <Sparkles className="h-4 w-4 mr-2" />
                สร้างงาน
              </Button>
            </CardFooter>
          </Card>
        </form>

        {/* Info card */}
        <Card className="mt-6">
          <CardContent className="py-4">
            <h3 className="font-medium text-gray-900 mb-2">วิธีการทำงาน</h3>
            <ol className="list-decimal list-inside space-y-1 text-sm text-gray-600">
              <li>กรอกคำอธิบาย concept วิดีโอของคุณ</li>
              <li>AI ของเราวิเคราะห์ concept</li>
              <li>สร้างเพลงตามอารมณ์</li>
              <li>สร้างภาพให้ตรงกับธีม</li>
              <li>รวมทุกอย่างเป็นวิดีโอ</li>
              <li>วิดีโอของคุณพร้อมดาวน์โหลด!</li>
            </ol>
          </CardContent>
        </Card>
      </main>
    </div>
  )
}
