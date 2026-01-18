# NUB Clubs Connect - Backend API

A comprehensive Go-based REST API for managing university clubs, events, news, and member engagement.

## Features

- **User Management**: Authentication, role-based access control (Student, Club Moderator, System Admin)
- **Club Management**: Create, manage clubs, and handle memberships
- **Event Management**: Create events, handle registrations, capacity management, and waitlists
- **News & Announcements**: Publish club news with multimedia support
- **Event Feedback**: Rating and feedback system for completed events
- **Notifications**: System notifications and announcements
- **Analytics**: Comprehensive dashboard with engagement metrics
- **Activity Logging**: Track user actions for audit purposes

## Tech Stack

- **Language**: Go 1.23+
- **Framework**: Gin Web Framework
- **Database**: PostgreSQL
- **Authentication**: JWT (JSON Web Tokens)
- **Password Hashing**: bcrypt

## Prerequisites

- Go 1.23 or higher
- PostgreSQL 12+
- Git

## Installation

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd nub_admin_api
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Setup environment variables**
   ```bash
   cp .env.example .env
   # Edit .env with your database URL and configuration
   ```

4. **Run the application**
   ```bash
   go run main.go
   ```

The API will be available at `http://localhost:8080`

## Project Structure

```
nub_admin_api/
├── main.go          # Application entry point
├── go.mod           # Go module definition
├── config/          # Configuration management
├── database/        # Database connection and setup
├── models/          # Data structures and models
├── handlers/        # HTTP request handlers (controllers)
├── middleware/      # Gin middleware
├── routes/          # API route definitions
├── utils/           # Utility functions
└── .env.example     # Environment variables template
```

## API Endpoints

### Authentication
- `POST /api/auth/register` - Register new user
- `POST /api/auth/login` - User login
- `POST /api/auth/forgot-password` - Request password reset
- `POST /api/auth/reset-password` - Reset password with token

### Users
- `GET /api/users/:id` - Get user profile
- `PUT /api/users/:id` - Update user profile
- `GET /api/users/:id/clubs` - Get user's clubs
- `GET /api/users/:id/events` - Get user's registered events

### Clubs
- `POST /api/clubs` - Create new club
- `GET /api/clubs` - Get all clubs
- `GET /api/clubs/:id` - Get club details
- `PUT /api/clubs/:id` - Update club
- `GET /api/clubs/:id/members` - Get club members
- `POST /api/clubs/:id/members` - Add member to club
- `DELETE /api/clubs/:id/members/:userId` - Remove member from club

### Events
- `POST /api/events` - Create event
- `GET /api/events` - Get all events
- `GET /api/events/:id` - Get event details
- `PUT /api/events/:id` - Update event
- `POST /api/events/:id/register` - Register for event
- `DELETE /api/events/:id/register/:userId` - Cancel registration
- `GET /api/events/:id/registrations` - Get event registrations
- `POST /api/events/:id/feedback` - Submit event feedback

### News
- `POST /api/news` - Create news post
- `GET /api/news` - Get all news
- `GET /api/news/:id` - Get news details
- `PUT /api/news/:id` - Update news
- `DELETE /api/news/:id` - Delete news

### Admin
- `GET /api/admin/dashboard` - Get dashboard statistics
- `GET /api/admin/analytics` - Get detailed analytics
- `GET /api/admin/events/pending` - Get pending events for approval
- `PUT /api/admin/events/:id/approve` - Approve event
- `PUT /api/admin/events/:id/reject` - Reject event

## Database Setup

The application uses PostgreSQL. Ensure your database is running and properly configured in your `.env` file.

## Environment Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `DATABASE_URL` | PostgreSQL connection string | `postgresql://user:pass@host/db` |
| `PORT` | Server port | `8080` |
| `GIN_MODE` | Gin mode (debug/release) | `debug` |
| `JWT_SECRET` | JWT signing secret | `your-secret-key` |
| `JWT_EXPIRATION` | JWT token expiration | `24h` |

## Development

### Running tests
```bash
go test ./...
```

### Building for production
```bash
go build -o nub_admin_api
```

## API Response Format

All responses follow a consistent JSON format:

**Success Response:**
```json
{
  "success": true,
  "message": "Operation successful",
  "data": {}
}
```

**Error Response:**
```json
{
  "success": false,
  "message": "Error description",
  "error": "error_code"
}
```

## Security Considerations

- All passwords are hashed using bcrypt
- JWT tokens are used for authentication
- Role-based access control is enforced
- Database queries use parameterized statements to prevent SQL injection
- HTTPS should be used in production

