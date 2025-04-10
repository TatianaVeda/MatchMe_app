import { useState, useEffect } from 'react'
import axios from 'axios'
import toast from 'react-hot-toast'

export default function Recommendations() {
  const [recommendations, setRecommendations] = useState([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    fetchRecommendations()
  }, [])

  const fetchRecommendations = async () => {
    try {
      const response = await axios.get('/api/recommendations')
      const userIds = response.data.user_ids
      const userDetails = await Promise.all(
        userIds.map(id => axios.get(`/api/users/${id}`))
      )
      setRecommendations(userDetails.map(res => res.data))
      setLoading(false)
    } catch (error) {
      console.error('Failed to fetch recommendations:', error)
      toast.error('Failed to load recommendations')
      setLoading(false)
    }
  }

  const handleConnect = async (userId) => {
    try {
      await axios.post('/api/connections/request', { receiver_id: userId })
      toast.success('Connection request sent!')
      setRecommendations(recommendations.filter(user => user.id !== userId))
    } catch (error) {
      console.error('Failed to send connection request:', error)
      toast.error('Failed to send connection request')
    }
  }

  const handleDismiss = async (userId) => {
    try {
      await axios.post('/api/recommendations/dismiss', { dismissed_user_id: userId })
      setRecommendations(recommendations.filter(user => user.id !== userId))
      toast.success('User dismissed')
    } catch (error) {
      console.error('Failed to dismiss user:', error)
      toast.error('Failed to dismiss user')
    }
  }

  if (loading) {
    return <div>Loading...</div>
  }

  return (
    <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
      {recommendations.map((user) => (
        <div key={user.id} className="bg-white rounded-lg shadow p-6">
          <div className="flex items-center space-x-4">
            <div className="h-12 w-12 rounded-full bg-gray-200 flex items-center justify-center">
              {user.avatar ? (
                <img src={user.avatar} alt={user.username} className="h-12 w-12 rounded-full" />
              ) : (
                <span className="text-2xl">👤</span>
              )}
            </div>
            <div>
              <h3 className="text-lg font-medium text-gray-900">{user.username}</h3>
            </div>
          </div>
          <div className="mt-4 flex space-x-3">
            <button
              onClick={() => handleConnect(user.id)}
              className="flex-1 bg-indigo-600 text-white px-4 py-2 rounded-md hover:bg-indigo-700"
            >
              Connect
            </button>
            <button
              onClick={() => handleDismiss(user.id)}
              className="flex-1 bg-gray-200 text-gray-700 px-4 py-2 rounded-md hover:bg-gray-300"
            >
              Dismiss
            </button>
          </div>
        </div>
      ))}
      {recommendations.length === 0 && (
        <div className="col-span-full text-center py-12">
          <p className="text-gray-500">No recommendations available at the moment.</p>
        </div>
      )}
    </div>
  )
}