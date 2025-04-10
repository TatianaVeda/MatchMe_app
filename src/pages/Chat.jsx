import { useState, useEffect, useRef } from 'react'
import { useParams } from 'react-router-dom'
import axios from 'axios'
import toast from 'react-hot-toast'

export default function Chat() {
  const { userId } = useParams()
  const [messages, setMessages] = useState([])
  const [newMessage, setNewMessage] = useState('')
  const [loading, setLoading] = useState(true)
  const messagesEndRef = useRef(null)

  useEffect(() => {
    if (userId) {
      fetchMessages()
    }
  }, [userId])

  const fetchMessages = async () => {
    try {
      const response = await axios.get('/api/chat/messages', {
        params: { receiver_id: userId }
      })
      setMessages(response.data.messages)
      setLoading(false)
    } catch (error) {
      console.error('Failed to fetch messages:', error)
      toast.error('Failed to load messages')
      setLoading(false)
    }
  }

  const handleSendMessage = async (e) => {
    e.preventDefault()
    if (!newMessage.trim()) return

    try {
      const response = await axios.post('/api/chat/send', {
        receiver_id: userId,
        message: newMessage
      })
      setMessages([...messages, response.data.chat])
      setNewMessage('')
      messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
    } catch (error) {
      console.error('Failed to send message:', error)
      toast.error('Failed to send message')
    }
  }

  if (loading) {
    return <div>Loading...</div>
  }

  if (!userId) {
    return (
      <div className="text-center py-12">
        <p className="text-gray-500">Select a connection to start chatting.</p>
      </div>
    )
  }

  return (
    <div className="bg-white shadow rounded-lg flex flex-col h-[calc(100vh-12rem)]">
      <div className="flex-1 overflow-y-auto p-4 space-y-4">
        {messages.map((message) => (
          <div
            key={message.id}
            className={`flex ${message.sender_id === parseInt(userId) ? 'justify-start' : 'justify-end'}`}
          >
            <div
              className={`max-w-xs px-4 py-2 rounded-lg ${
                message.sender_id === parseInt(userId)
                  ? 'bg-gray-100'
                  : 'bg-indigo-600 text-white'
              }`}
            >
              <p>{message.message}</p>
              <p className="text-xs mt-1 opacity-75">
                {new Date(message.timestamp).toLocaleTimeString()}
              </p>
            </div>
          </div>
        ))}
        <div ref={messagesEndRef} />
      </div>
      <form onSubmit={handleSendMessage} className="p-4 border-t">
        <div className="flex space-x-4">
          <input
            type="text"
            value={newMessage}
            onChange={(e) => setNewMessage(e.target.value)}
            placeholder="Type a message..."
            className="flex-1 rounded-lg border-gray-300 focus:border-indigo-500 focus:ring-indigo-500"
          />
          <button
            type="submit"
            className="bg-indigo-600 text-white px-4 py-2 rounded-lg hover:bg-indigo-700"
          >
            Send
          </button>
        </div>
      </form>
    </div>
  )
}