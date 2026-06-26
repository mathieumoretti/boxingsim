# Boxing Simulator Web UI

## Overview

This document describes the web user interface for the Boxing Simulator application. The UI provides an interactive way to manage boxers, users, and fights in the boxing simulation game.

## Features Implemented

1. **User Authentication**
   - Login functionality
   - Registration functionality
   - Session management with local storage

2. **Boxer Management**
   - Create new boxers with different classes (heavyweight, middleweight, lightweight, flyweight)
   - View list of user's boxers
   - Boxer statistics display

3. **Fight System**
   - Select two boxers for a fight
   - Start fights between boxers
   - Fight arena visualization

## Architecture

### Frontend Structure
- `index.html` - Main HTML structure with layout and components
- `styles.css` - Styling using modern CSS with responsive design
- `app.js` - JavaScript functionality handling API calls and UI interactions
- `server.go` - Simple static file server for serving the web UI

### API Integration Points
The UI communicates with the backend API at `http://localhost:8080` using standard HTTP requests:

1. **Authentication**
   - POST `/auth/register` - Register new user
   - POST `/auth/login` - Login user

2. **Boxer Management**
   - GET `/boxers` - Get all boxers
   - GET `/boxers/{id}` - Get specific boxer
   - POST `/boxers` - Create new boxer
   - PUT `/boxers/{id}` - Update boxer
   - DELETE `/boxers/{id}` - Delete boxer

3. **User Management**
   - GET `/users/{id}` - Get user details
   - GET `/users/{id}/boxers` - Get boxers for a user

4. **Fight System**
   - POST `/boxers/fight` - Start a fight between two boxers

## How to Run

### Prerequisites
- Go 1.19 or higher installed
- Backend server running on `http://localhost:8080`

### Steps
1. Start the backend server:
   ```bash
   make run
   ```

2. In a separate terminal, start the web UI server:
   ```bash
   make web-dev
   ```

3. Open your browser and navigate to `http://localhost:8081`

## Design Principles

### Responsive Design
The UI is designed to work on different screen sizes using CSS Grid and Flexbox layouts.

### User Experience
- Clear visual hierarchy
- Intuitive navigation between sections
- Immediate feedback for user actions
- Error handling with user-friendly messages

### Security
- Token-based authentication stored in localStorage
- Proper error handling for API calls
- Input validation on the client side

## Future Enhancements

1. **Advanced Fight Simulation**
   - Real-time fight visualization
   - Damage calculation and animations
   - Win/loss tracking

2. **Boxer Progression System**
   - Skill point allocation
   - Training sessions
   - Equipment management

3. **Multiplayer Features**
   - Leaderboards
   - Challenge other players
   - Clan/group systems

4. **Advanced UI Components**
   - Charts for boxer statistics
   - Interactive world map
   - Match history tracking

## Technical Notes

### API Communication
The JavaScript code in `app.js` handles all API communication with proper error handling and user feedback.

### Local Storage
Authentication tokens are stored in localStorage for session persistence between browser sessions.

### Error Handling
All API calls include proper error handling with user-friendly messages displayed via the notification system.

## Contributing

To contribute to this UI:
1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request

The UI follows the same coding standards and patterns as the backend application.