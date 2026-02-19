# PicoClaw Web UI Setup Guide

This guide will help you set up and use the PicoClaw web interface to interact with your AI agent through a browser.

## Overview

The web UI provides a modern chat interface that connects to your PicoClaw agent via WebSocket. You can use it from your laptop or any device on your network.

## Architecture

```
┌─────────────────┐         WebSocket          ┌──────────────────┐
│   Web Browser   │ ◄──────────────────────────► │  PicoClaw       │
│  (localhost:    │         ws://host:8080      │  Gateway Server │
│   5173)         │                             │  (port 8080)    │
└─────────────────┘                             └──────────────────┘
                                                        │
                                                        ▼
                                                  ┌──────────────┐
                                                  │  Agent Loop  │
                                                  │  & LLM       │
                                                  └──────────────┘
```

## Step-by-Step Setup

### Step 1: Configure PicoClaw

First, enable the web channel in your PicoClaw configuration file.

**Location:** `~/.picoclaw/config.json`

Add or update the web channel configuration:

```json
{
  "agents": {
    "defaults": {
      "workspace": "~/.picoclaw/workspace",
      "provider": "openrouter",
      "model": "anthropic/claude-3.5-sonnet",
      "max_tokens": 8192
    }
  },
  "channels": {
    "web": {
      "enabled": true,
      "host": "0.0.0.0",
      "port": 8080,
      "allow_from": []
    }
  },
  "providers": {
    "openrouter": {
      "api_key": "your-api-key-here"
    }
  }
}
```

**Important settings:**
- `enabled: true` - Activates the web channel
- `host: "0.0.0.0"` - Allows connections from any network interface (use "127.0.0.1" for localhost only)
- `port: 8080` - The port for the WebSocket server
- `allow_from: []` - Empty array means allow all connections (add specific IDs for restrictions)

### Step 2: Start PicoClaw Gateway

Build and run PicoClaw with the gateway command:

```bash
# Build PicoClaw (if needed)
make build

# Run the gateway
./picoclaw gateway
```

Or if installed:

```bash
picoclaw gateway
```

You should see output like:

```
[INFO] web channel starting on 0.0.0.0:8080
[INFO] Web channel enabled successfully
[INFO] All channels started
```

### Step 3: Install Web UI Dependencies

Navigate to the web directory and install dependencies:

```bash
cd web
npm install
```

This will install:
- React 19
- Vite (build tool)
- Other required dependencies

### Step 4: Start the Web UI

Start the development server:

```bash
npm run dev
```

You should see:

```
  VITE v6.x.x  ready in xxx ms

  ➜  Local:   http://localhost:5173/
  ➜  Network: use --host to expose
```

### Step 5: Connect Your Browser

1. Open your browser to `http://localhost:5173`
2. You should see the PicoClaw AI Agent interface
3. Check the connection status indicator in the top-right
4. When it shows "Connected" with a green dot, you're ready!

## Using the Web UI

### Basic Chat

1. Type your message in the text area at the bottom
2. Press Enter or click "Send"
3. Your message appears on the right (blue)
4. PicoClaw's response appears on the left (white)

### Keyboard Shortcuts

- **Enter** - Send message
- **Shift + Enter** - New line in message

### Features

- **Real-time Updates** - Messages appear instantly via WebSocket
- **Connection Status** - Green dot when connected, red when disconnected
- **Auto-scroll** - Chat automatically scrolls to latest message
- **Session Management** - Each browser connection gets a unique chat ID

## Connecting from Your Laptop

If you're running PicoClaw on one device and want to access it from your laptop:

### 1. Find Your Server IP

On the machine running PicoClaw:

```bash
# Linux/Mac
ip addr show | grep inet

# Or
ifconfig | grep inet

# Windows
ipconfig
```

Look for your local network IP (usually starts with 192.168.x.x or 10.x.x.x)

### 2. Update WebSocket URL

Edit `web/src/App.jsx` and change the WebSocket connection:

```javascript
// Change this line:
const websocket = new WebSocket('ws://localhost:8080/ws')

// To use your server's IP:
const websocket = new WebSocket('ws://192.168.1.100:8080/ws')
```

### 3. Rebuild and Access

```bash
cd web
npm run build
```

Then open `http://your-server-ip:5173` in your laptop's browser.

## Production Deployment

For production use, build the static files and serve them:

```bash
cd web
npm run build
```

The `dist/` directory will contain optimized static files. Serve these with:

**Option 1: Simple HTTP Server**
```bash
cd dist
python3 -m http.server 8000
```

**Option 2: Nginx**
```nginx
server {
    listen 80;
    server_name your-domain.com;

    location / {
        root /path/to/picoclaw/web/dist;
        try_files $uri $uri/ /index.html;
    }
}
```

**Option 3: Serve from PicoClaw**
You could extend the web channel to serve static files directly from Go.

## Security Considerations

### For Local Use

The default configuration is suitable for local development:
- Binds to 0.0.0.0 (all interfaces)
- No authentication required
- CORS allows all origins

### For Production

Consider these security enhancements:

1. **Use HTTPS/WSS**
   - Set up TLS certificates
   - Use `wss://` instead of `ws://`

2. **Add Authentication**
   - Implement token-based auth
   - Use the `allow_from` configuration

3. **Restrict CORS**
   - Limit allowed origins in the web channel

4. **Use Reverse Proxy**
   - Put Nginx or Caddy in front
   - Handle SSL termination there

## Troubleshooting

### Connection Failed

**Problem:** Red "Disconnected" indicator

**Solutions:**
1. Check PicoClaw gateway is running: `ps aux | grep picoclaw`
2. Verify port 8080 is open: `netstat -an | grep 8080`
3. Check firewall settings
4. Look at PicoClaw logs for errors

### WebSocket Errors

**Problem:** Console shows WebSocket errors

**Solutions:**
1. Check the WebSocket URL in App.jsx
2. Verify CORS is not blocking (should see Access-Control headers)
3. Check browser console for detailed errors
4. Try a different browser

### Messages Not Sending

**Problem:** Messages appear but no response

**Solutions:**
1. Check PicoClaw has a valid API key configured
2. Look at gateway logs: `picoclaw gateway -v` (verbose mode)
3. Verify the LLM provider is accessible
4. Check workspace files exist

### Port Already in Use

**Problem:** `address already in use` error

**Solutions:**
1. Change port in config.json
2. Kill existing process: `lsof -ti:8080 | xargs kill -9`
3. Use a different port (e.g., 8081)

## Advanced Configuration

### Custom Port

Change in both places:

**config.json:**
```json
{
  "channels": {
    "web": {
      "port": 3000
    }
  }
}
```

**App.jsx:**
```javascript
const websocket = new WebSocket('ws://localhost:3000/ws')
```

### Access Control

Restrict connections to specific chat IDs:

```json
{
  "channels": {
    "web": {
      "allow_from": ["web-12345", "web-67890"]
    }
  }
}
```

### Multiple Channels

Run web alongside other channels:

```json
{
  "channels": {
    "web": {
      "enabled": true,
      "port": 8080
    },
    "telegram": {
      "enabled": true,
      "token": "your-bot-token"
    },
    "discord": {
      "enabled": true,
      "token": "your-bot-token"
    }
  }
}
```

All channels share the same agent brain and memory!

## API Reference

### WebSocket Protocol

The web channel uses JSON messages:

**Client → Server (Send Message):**
```json
{
  "content": "Hello, PicoClaw!",
  "chatId": "web-1234567890"
}
```

**Server → Client (Welcome):**
```json
{
  "type": "welcome",
  "chatId": "web-1234567890"
}
```

**Server → Client (Message):**
```json
{
  "type": "message",
  "content": "Hello! How can I help you?",
  "chatId": "web-1234567890"
}
```

### HTTP Endpoints

**POST /api/chat** - Send message via HTTP
```bash
curl -X POST http://localhost:8080/api/chat \
  -H "Content-Type: application/json" \
  -d '{"content": "Hello", "chatId": "web-123"}'
```

**GET /api/health** - Check server status
```bash
curl http://localhost:8080/api/health
```

## Next Steps

- **Customize the UI** - Edit `web/src/App.jsx` and `web/src/App.css`
- **Add Features** - Implement file upload, voice input, etc.
- **Deploy to Cloud** - Host on Vercel, Netlify, or your own server
- **Mobile App** - Use the same WebSocket protocol in a mobile app

## Getting Help

- Check PicoClaw logs for error messages
- Look at browser console for frontend errors
- Review the main README.md for general PicoClaw setup
- Open an issue on GitHub if you encounter bugs

## Credits

Built with:
- **PicoClaw** - Ultra-lightweight AI agent framework
- **React** - UI library
- **Vite** - Build tool
- **Gorilla WebSocket** - Go WebSocket library
