import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import axios from 'axios'
import toast from 'react-hot-toast'

export default function Connections() {
  const [connections, setConnections] = useState([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    fetchConnections()
  }, [])

  const fetchConnections = async () => {
    try {
      const response = await axios.get('/api/connections')
      const userIds = response.data.connections
      const userDetails = await Promise.all(
        userIds.map(id => axios.get(`/api/users/${id}`))
      )
      setConnections(userDetails.map(res => res.data))
      setLoading(false)
    } catch (error) {
      console.error('Failed to fetch connections:', error)
      toast.error('Failed to load connections')
      setLoading(false)
    }
  }

  const handleRemoveConnection = async (userId) => {
    try {
      await axios.post('/api/connections/remove', { user_id: userId })
      setConnections(connections.filter(user => user.id !== userId))
      toast.success('Connection removed')
    } catch (error) {
      console.error('Failed to remove connection:', error)
      toast.error('Failed to remove connection')
    }
  }

  if (loading) {
    return <div>Loading...</div>
  }

  return (
    <div className="bg-white shadow rounded-lg divide-y divide-gray-200">
      {connections.map((user) => (
        <div key={user.id} className="p-6">
          <div className="flex items-center justify-between">
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
            <div className="flex space-x-3">
              <Link
                to={`/chat/${user.id}`}
                className="bg-indigo-600 text-white px-4 py-2 rounded-md hover:bg-indigo-700"
              >
                Chat
              </Link>
              <button
                onClick={() => handleRemoveConnection(user.id)}
                className="bg-gray-200 text-gray-700 px-4 py-2 rounded-md hover:bg-gray-300"
              >
                Remove
              </button>
            </div>
          </div>
        </div>
      ))}
      {connections.length === 0 && (
        <div className="text-center py-12">
          <p className="text-gray-500">No connections yet.</p>
        </div>
      )}
    </div>
  )
}