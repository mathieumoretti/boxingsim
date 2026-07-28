# 3. Authentication System

## Status
Accepted

## Context
The boxing simulator needed a secure authentication system to:
- Manage user accounts and sessions
- Protect API endpoints from unauthorized access
- Support user-specific boxer management
- Provide token-based stateless authentication

## Decision
We implemented JWT-based authentication with the following approach:

**Authentication Flow**:
1. Users register via POST /auth/register with username, email, and password
2. Passwords are hashed using bcrypt for security
3. Upon login (POST /auth/login), a JWT token is generated with user claims
4. Subsequent requests include the JWT in Authorization header
5. Token verification ensures access to protected endpoints

**Security Features**:
- Password hashing with bcrypt
- JWT token generation and validation
- Secret-based signing for integrity
- Proper error handling to prevent information leakage
- Database connection safety checks

## Consequences
- **Pros**:
  - Stateless authentication allows horizontal scaling
  - JWT tokens provide secure, self-contained user identification
  - Bcrypt hashing ensures password security
  - Clear separation of authentication concerns from business logic
  - Easy integration with existing HTTP handlers

- **Cons**:
  - Tokens need to be refreshed periodically (though refresh token not yet implemented)
  - No built-in session management (tokens must be revoked manually)
  - JWT tokens are stored client-side, so security depends on proper storage practices
  - Requires careful handling of token expiration and renewal

## Alternatives Considered
1. Session-based authentication: Would require server-side state management but provides easier token revocation
2. OAuth2/OpenID Connect: More complex to implement but provides integration with external identity providers
3. API Keys: Simpler but less secure for user-specific access control
4. Single Sign-On (SSO): Would add complexity but enable enterprise integration

The JWT approach was chosen because it provides a good balance of security, scalability, and simplicity while fitting well with the stateless nature of REST APIs.