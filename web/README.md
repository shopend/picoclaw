# PicoClaw Web UI

A modern web interface for interacting with the PicoClaw AI agent through your browser.

## Features

- 🔄 Real-time WebSocket communication
- 💬 Chat-style interface
- 🎨 Modern, responsive design
- 🟢 Connection status indicator
- ⚡ Fast and lightweight

## Setup

### 1. Configure PicoClaw

Make sure your PicoClaw config at `~/.picoclaw/config.json` has the web channel enabled:

```json
{
  "channels": {
    "web": {
      "enabled": true,
      "host": "0.0.0.0",
      "port": 8080,
      "allow_from": []
    }
  }
}
```

### 2. Start PicoClaw Gateway

Run the PicoClaw gateway to start the backend server:

```bash
picoclaw gateway
```

This will start the web channel on `http://localhost:8080`

### 3. Start the Web UI

In this directory, install dependencies and start the dev server:

```bash
npm install
npm run dev
```

The web UI will be available at `http://localhost:5173`

## Usage

1. Open your browser to `http://localhost:5173`
2. Wait for the connection indicator to show "Connected"
3. Start chatting with your PicoClaw AI agent!

## Building for Production

To build the web UI for production:

```bash
npm run build
```

The built files will be in the `dist/` directory. You can serve these with any static file server.

## Configuration

The WebSocket connection URL is hardcoded to `ws://localhost:8080/ws`. To connect to a different host, modify the WebSocket URL in `src/App.jsx`:

```javascript
const websocket = new WebSocket('ws://your-host:8080/ws')
```

## Troubleshooting

**Connection Failed?**
- Make sure PicoClaw gateway is running with `picoclaw gateway`
- Check that the web channel is enabled in your config
- Verify the port (8080) is not being used by another application

**WebSocket Errors?**
- Check browser console for detailed error messages
- Ensure CORS is properly configured (it should be handled automatically)

## Tech Stack

- React 19
- Vite
- WebSocket API
- Modern CSS with animations
