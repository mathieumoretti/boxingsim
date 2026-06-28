# Web UI for Boxing Simulator

This directory contains the frontend files for the Boxing Simulator web application.

## Structure

- `index.html` - Main HTML structure with layout and components
- `styles.css` - Styling using modern CSS with responsive design  
- `app.js` - JavaScript functionality handling API calls and UI interactions
- `server.go` - Simple static file server for serving the web UI (deprecated - now integrated into main server)

## How it works

The web UI is served directly by the main application server at port 8080. All routes that don't match API endpoints will serve static files from this directory.

## Development

To run the complete application with the web UI:

1. Start the database services:
   ```bash
   make docker-up
   ```

2. Run the main server (which includes UI serving):
   ```bash
   make dev
   ```

3. Open your browser and navigate to `http://localhost:8080`