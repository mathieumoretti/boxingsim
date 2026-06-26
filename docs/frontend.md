# Frontend Development for Boxing Simulator

The boxing simulation backend provides a REST API that can be consumed by any frontend application. Here are several approaches to create a user interface:

## Available API Endpoints

### Authentication
- `POST /auth/register` - Register a new user
- `POST /auth/login` - Login and receive JWT token

### Boxer Management
- `GET /boxers` - List all boxers for the authenticated user
- `POST /boxers` - Create a new boxer
- `GET /boxers/{id}` - Get details of a specific boxer

### World State
- `GET /world/time` - Get current game time (to be implemented)

## Frontend Implementation Approaches

### 1. Web Application (React/Vue/Angular)
Create a web-based UI that:
- Handles user authentication
- Displays boxer statistics and status
- Shows training queue and scheduled events
- Provides controls for managing boxers
- Visualizes the boxing world state

### 2. Mobile Application (React Native/Flutter)
Build a mobile app for:
- On-the-go boxer management
- Training scheduling
- Fight tracking
- Social features

### 3. Desktop Application (Electron)
Create a desktop client that:
- Provides full simulation controls
- Displays training progress
- Shows fight history and rankings

## Sample Frontend Flow

1. **User Registration/Login**
   - User creates account via `/auth/register`
   - User logs in via `/auth/login` to get JWT token

2. **Boxer Creation**
   - User creates initial boxer via `/boxers` POST
   - Boxer is stored in database with initial stats

3. **Training Management**
   - View training queue via `/training/queue` (to be implemented)
   - Schedule training sessions
   - See stat improvements over time

4. **World Interaction**
   - View current game time
   - See scheduled events and fights
   - Track boxer progress

## Development Tools

### For Web Frontend:
- React.js or Vue.js with TypeScript
- Axios for API calls
- Redux or Context API for state management
- Tailwind CSS or Material UI for styling

### For Mobile:
- React Native with Expo
- Flutter with Dart
- Native mobile development (iOS/Android)

## Sample Implementation Structure

```
/src
  /components
    /auth
      Login.js
      Register.js
    /boxers
      BoxerList.js
      BoxerDetail.js
      BoxerForm.js
    /training
      TrainingQueue.js
      ScheduleTraining.js
  /services
    api.js (API client)
    auth.js (authentication helper)
  /pages
    Home.js
    Dashboard.js
    Profile.js
  App.js
```

## Next Steps for Development

1. **Implement Missing Endpoints**
   - Add `/world/time` endpoint to get current game time
   - Implement training queue endpoints
   - Add fight scheduling endpoints

2. **Build Simple UI**
   - Create basic HTML/JS interface for testing
   - Use tools like Postman or curl for API testing
   - Gradually build out the frontend

3. **Enhance Backend Functionality**
   - Complete fight simulation system
   - Implement matchmaking
   - Add economy and betting features

## Quick Start with Curl

You can test the backend with simple curl commands:

```bash
# Register a user
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","email":"test@example.com","password":"password123","boxer_name":"Test Fighter"}'

# Login
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"password123"}'

# Create a boxer (after login, use the token in Authorization header)
curl -X POST http://localhost:8080/boxers \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Test Fighter","strength":60.0,"defense":50.0,"agility":55.0}'
```

## Development Roadmap

1. **Phase 1**: Basic web UI with authentication and boxer management
2. **Phase 2**: Training system visualization and scheduling
3. **Phase 3**: Fight simulation with visual representation
4. **Phase 4**: Advanced features like matchmaking, economy, social elements

The backend is fully functional for these purposes and can support any type of frontend client you choose to build.