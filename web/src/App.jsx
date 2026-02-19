import { useState, useEffect, useRef } from 'react'
import './App.css'

function App() {
  const [messages, setMessages] = useState([])
  const [input, setInput] = useState('')
  const [ws, setWs] = useState(null)
  const [connected, setConnected] = useState(false)
  const [chatId, setChatId] = useState(null)
  const messagesEndRef = useRef(null)

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }

  useEffect(() => {
    scrollToBottom()
  }, [messages])

  useEffect(() => {
    const websocket = new WebSocket('ws://localhost:8080/ws')

    websocket.onopen = () => {
      console.log('Connected to PicoClaw')
      setConnected(true)
    }

    websocket.onmessage = (event) => {
      const data = JSON.parse(event.data)

      if (data.type === 'welcome') {
        setChatId(data.chatId)
        setMessages([{ role: 'system', content: 'Connected to PicoClaw AI Agent' }])
      } else if (data.type === 'message') {
        setMessages(prev => [...prev, { role: 'assistant', content: data.content }])
      }
    }

    websocket.onerror = (error) => {
      console.error('WebSocket error:', error)
      setConnected(false)
    }

    websocket.onclose = () => {
      console.log('Disconnected from PicoClaw')
      setConnected(false)
    }

    setWs(websocket)

    return () => {
      websocket.close()
    }
  }, [])

  const sendMessage = () => {
    if (!input.trim() || !ws || !connected) return

    const userMessage = { role: 'user', content: input }
    setMessages(prev => [...prev, userMessage])

    ws.send(JSON.stringify({
      content: input,
      chatId: chatId
    }))

    setInput('')
  }

  const handleKeyPress = (e) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      sendMessage()
    }
  }

  return (
    <div className="app">
      <header className="header">
        <h1>🤖 PicoClaw AI Agent</h1>
        <div className="status">
          <span className={`status-dot ${connected ? 'connected' : 'disconnected'}`}></span>
          <span>{connected ? 'Connected' : 'Disconnected'}</span>
        </div>
      </header>

      <div className="chat-container">
        <div className="messages">
          {messages.map((msg, idx) => (
            <div key={idx} className={`message ${msg.role}`}>
              <div className="message-content">
                <strong>{msg.role === 'user' ? 'You' : msg.role === 'assistant' ? 'PicoClaw' : 'System'}:</strong>
                <p>{msg.content}</p>
              </div>
            </div>
          ))}
          <div ref={messagesEndRef} />
        </div>

        <div className="input-container">
          <textarea
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onKeyPress={handleKeyPress}
            placeholder="Type your message... (Press Enter to send)"
            disabled={!connected}
            rows="3"
          />
          <button onClick={sendMessage} disabled={!connected || !input.trim()}>
            Send
          </button>
        </div>
      </div>

      <footer className="footer">
        <p>PicoClaw - Ultra-lightweight Personal AI Agent</p>
        <p className="info">Chat ID: {chatId || 'Connecting...'}</p>
      </footer>
    </div>
  )
}

export default App
