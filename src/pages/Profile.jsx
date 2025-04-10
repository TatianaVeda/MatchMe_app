import { useState, useEffect } from 'react'
import axios from 'axios'
import toast from 'react-hot-toast'

export default function Profile() {
  const [profile, setProfile] = useState(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const fetchProfile = async () => {
      try {
        const [profileRes, bioRes] = await Promise.all([
          axios.get('/api/me/profile'),
          axios.get('/api/me/bio')
        ])
        setProfile({ ...profileRes.data, ...bioRes.data })
        setLoading(false)
      } catch (error) {
        console.error('Failed to fetch profile:', error)
        toast.error('Failed to load profile')
        setLoading(false)
      }
    }
    fetchProfile()
  }, [])

  if (loading) {
    return <div>Loading...</div>
  }

  return (
    <div className="bg-white shadow rounded-lg">
      <div className="px-4 py-5 sm:p-6">
        <h3 className="text-lg font-medium leading-6 text-gray-900">Profile Information</h3>
        <div className="mt-5 border-t border-gray-200">
          <dl className="divide-y divide-gray-200">
            {profile && Object.entries(profile).map(([key, value]) => (
              <div key={key} className="py-4 sm:grid sm:grid-cols-3 sm:gap-4">
                <dt className="text-sm font-medium text-gray-500 capitalize">
                  {key.replace(/([A-Z])/g, ' $1').trim()}
                </dt>
                <dd className="mt-1 text-sm text-gray-900 sm:col-span-2 sm:mt-0">
                  {value || 'Not set'}
                </dd>
              </div>
            ))}
          </dl>
        </div>
      </div>
    </div>
  )
}