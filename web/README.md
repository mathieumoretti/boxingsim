# Boxing Simulator Web UI

This directory contains the web user interface for the Boxing Simulator application.

## Features

- User authentication (login/register)
- Boxer management and creation
- Fight arena with boxer selection
- Responsive design that works on desktop and mobile

## Getting Started

### Prerequisites
- Go 1.19 or higher installed
- Backend server running on `http://localhost:8080`

### Running the UI Server

You can run the web UI server in two ways:

1. Using the Makefile (recommended):
   ```bash
   make web-dev
   ```

2. Directly with Go:
   ```bash
   cd web
   go run server.go
   ```

The UI will be available at `http://localhost:8081`

## Architecture

### Files Structure
- `index.html` - Main HTML structure
- `styles.css` - Styling and responsive design
- `app.js` - JavaScript functionality and API integration
- `server.go` - Simple static file server to serve the UI

### API Integration
The UI communicates with the backend API at `http://localhost:8080` using standard HTTP requests for:
- Authentication (login/register)
- Boxer management (create, get, update, delete)
- User management
- Fight system

## Development

To add new features:
1. Modify `index.html` for new UI components
2. Update `styles.css` for styling changes
3. Extend `app.js` with new JavaScript functionality
4. Ensure API endpoints match the backend

## Customization

The UI is designed to be easily customizable:
- Change colors in `styles.css`
- Add new sections in `index.html`
- Extend functionality in `app.js`