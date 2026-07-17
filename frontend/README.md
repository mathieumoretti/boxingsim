# Boxing Simulator Frontend

This is the React-based frontend for the Boxing Simulator application.

## Features

- Modern React UI with functional components and hooks
- Responsive design using CSS Grid and Flexbox
- Client-side routing with React Router
- Authentication flows (login, register)
- Boxer management interface
- Fight arena simulation

## Technologies Used

- React 18+
- React Router 6+
- Webpack 5+
- Babel 7+
- CSS Modules

## Development Setup

### Prerequisites

- Node.js and npm
- Make (for development commands)

### Installation

```bash
npm install
```

### Running in Development Mode

```bash
npm start
```

This will start the development server with hot reloading on port 3000.

### Building for Production

```bash
npm run build
```

This creates an optimized production bundle in the `dist/` directory.

## Project Structure

```
src/
├── index.js          # Entry point
├── App.js            # Main application component
├── index.css         # Global styles
└── components/       # Reusable UI components
    ├── Login.js
    ├── Register.js
    ├── Dashboard.js
    ├── CreateBoxer.js
    └── Auth.css
```

## API Integration

The frontend communicates with the backend through HTTP requests to the Go server running on port 8080.

### Authentication Endpoints
- `POST /auth/register` - Register a new user
- `POST /auth/login` - Login existing user

### Boxer Endpoints
- `POST /boxers` - Create a new boxer
- `GET /users/{id}/boxers` - Get user's boxers

## Deployment

The production build should be served by the Go backend server. The contents of the `dist/` directory are what get served when the Go server handles static file requests.