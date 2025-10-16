# Handlers Documentation

## Overview

Handlers layer bertanggung jawab untuk menangani HTTP requests, melakukan validasi input, memanggil service layer, dan mengembalikan HTTP responses dalam format yang konsisten.

## Architecture

```
handlers/
└── handler.go       # HTTP request handlers untuk semua endpoints
```

## Response Format

Semua responses menggunakan format standar:

```json
{
  "status": 1,           // 1 = success, 0 = error
  "status_code": 200,    // HTTP status code
  "message": "Success message",
  "data": {}             // Response data atau error details
}
```

---

## API Endpoints

### Base URL
```
http://localhost:9001/api/v1
```

---

## 1. Register User

**Endpoint:** `POST /api/v1/auth/register`

**Description:** Register new user account

**Request Headers:**
```
Content-Type: application/json
```

**Request Body:**
```json
{
  "full_name": "Reza Ilham",
  "email": "reza@x.com",
  "phone": "+12345678900",
  "password": "Secret123!"
}
```

**Success Response (201):**
```json
{
  "status": 1,
  "status_code": 201,
  "message": "Registration success",
  "data": {
    "id": 1,
    "full_name": "Reza Ilham",
    "email": "reza@x.com",
    "phone": "+12345678900",
    "avatar_url": null,
    "token": {
      "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
    }
  }
}
```

**Error Responses:**

**400 - Validation Error:**
```json
{
  "status": 0,
  "status_code": 400,
  "message": "Validation error",
  "data": {
    "email": ["invalid format"]
  }
}
```

**409 - Email Already Exists:**
```json
{
  "status": 0,
  "status_code": 409,
  "message": "Email already registered",
  "data": {}
}
```

**cURL Example:**
```bash
curl -X POST http://localhost:9001/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "full_name": "Reza Ilham",
    "email": "reza@x.com",
    "phone": "+12345678900",
    "password": "Secret123!"
  }'
```

---

## 2. Login User

**Endpoint:** `POST /api/v1/auth/login`

**Description:** Authenticate user and get access token

**Request Body:**
```json
{
  "email": "reza@x.com",
  "password": "Secret123!"
}
```

**Success Response (200):**
```json
{
  "status": 1,
  "status_code": 200,
  "message": "Login success",
  "data": {
    "id": 1,
    "full_name": "Reza Ilham",
    "email": "reza@x.com",
    "phone": "+12345678900",
    "avatar_url": null,
    "token": {
      "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
    }
  }
}
```

**Error Response (401):**
```json
{
  "status": 0,
  "status_code": 401,
  "message": "Invalid email or password",
  "data": {}
}
```

**cURL Example:**
```bash
curl -X POST http://localhost:9001/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "reza@x.com",
    "password": "Secret123!"
  }'
```

---

## 3. Get User Profile

**Endpoint:** `GET /api/v1/me`

**Description:** Get logged-in user profile

**Request Headers:**
```
Authorization: Bearer <access_token>
```

**Success Response (200):**
```json
{
  "status": 1,
  "status_code": 200,
  "message": "Profile loaded successfully",
  "data": {
    "id": 1,
    "full_name": "Reza Ilham",
    "email": "reza@x.com",
    "phone": "+1 234 567 8900",
    "avatar_url": null
  }
}
```

**Error Response (401):**
```json
{
  "status": 0,
  "status_code": 401,
  "message": "Unauthorized - invalid or expired token",
  "data": {}
}
```

**cURL Example:**
```bash
curl -X GET http://localhost:9001/api/v1/me \
  -H "Authorization: Bearer <access_token>"
```

---

## 4. Update User Profile

**Endpoint:** `PUT /api/v1/me`

**Description:** Update logged-in user profile

**Request Headers:**
```
Authorization: Bearer <access_token>
Content-Type: application/json
```

**Request Body:**
```json
{
  "full_name": "Reza Ilham",
  "phone": "+1 234 567 8900"
}
```

**Success Response (200):**
```json
{
  "status": 1,
  "status_code": 200,
  "message": "Profile updated successfully",
  "data": {
    "id": 1,
    "full_name": "Reza Ilham",
    "email": "reza@x.com",
    "phone": "+1 234 567 8900",
    "avatar_url": null
  }
}
```

**Error Response (400):**
```json
{
  "status": 0,
  "status_code": 400,
  "message": "Validation error",
  "data": {
    "full_name": ["must not be empty"]
  }
}
```

**cURL Example:**
```bash
curl -X PUT http://localhost:9001/api/v1/me \
  -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "full_name": "Reza Ilham",
    "phone": "+1 234 567 8900"
  }'
```

---

## 5. List Contacts

**Endpoint:** `GET /api/v1/contacts?q=&page=1&limit=20`

**Description:** Get contact list with search and pagination

**Request Headers:**
```
Authorization: Bearer <access_token>
```

**Query Parameters:**
- `q` (optional): Search query for full_name or phone
- `page` (optional): Page number (default: 1)
- `limit` (optional): Items per page (default: 20, max: 100)
- `favorite` (optional): Filter by favorite status (true/false)

**Success Response (200):**
```json
{
  "status": 1,
  "status_code": 200,
  "message": "Contacts loaded successfully",
  "data": {
    "count": 4,
    "page": 1,
    "limit": 20,
    "contacts": [
      {
        "id": 1,
        "full_name": "John Doe",
        "phone": "0898209890",
        "email": null,
        "favorite": false,
        "created_at": "2025-10-16T10:00:00Z",
        "updated_at": "2025-10-16T10:00:00Z"
      }
    ]
  }
}
```

**Error Response (401):**
```json
{
  "status": 0,
  "status_code": 401,
  "message": "Unauthorized",
  "data": {}
}
```

**cURL Examples:**

Search contacts:
```bash
curl -X GET "http://localhost:9001/api/v1/contacts?q=john&page=1&limit=10" \
  -H "Authorization: Bearer <access_token>"
```

Filter favorites:
```bash
curl -X GET "http://localhost:9001/api/v1/contacts?favorite=true&page=1&limit=20" \
  -H "Authorization: Bearer <access_token>"
```

---

## 6. Create Contact

**Endpoint:** `POST /api/v1/contacts`

**Description:** Add new contact

**Request Headers:**
```
Authorization: Bearer <access_token>
Content-Type: application/json
```

**Request Body:**
```json
{
  "full_name": "John Doe",
  "phone": "+1 234 567 8900",
  "email": "john@example.com"
}
```

**Success Response (201):**
```json
{
  "status": 1,
  "status_code": 201,
  "message": "Contact created successfully",
  "data": {
    "id": 5,
    "full_name": "John Doe",
    "phone": "+1 234 567 8900",
    "email": "john@example.com",
    "favorite": false,
    "created_at": "2025-10-16T10:00:00Z",
    "updated_at": "2025-10-16T10:00:00Z"
  }
}
```

**Error Response (409):**
```json
{
  "status": 0,
  "status_code": 409,
  "message": "Contact phone already exists",
  "data": {
    "phone": ["0898209890"]
  }
}
```

**cURL Example:**
```bash
curl -X POST http://localhost:9001/api/v1/contacts \
  -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "full_name": "John Doe",
    "phone": "+1 234 567 8900",
    "email": "john@example.com"
  }'
```

---

## 7. Get Contact Detail

**Endpoint:** `GET /api/v1/contacts/{id}`

**Description:** Get contact detail by ID

**Request Headers:**
```
Authorization: Bearer <access_token>
```

**Success Response (200):**
```json
{
  "status": 1,
  "status_code": 200,
  "message": "Contact detail loaded",
  "data": {
    "id": 5,
    "full_name": "John Doe",
    "phone": "+1 234 567 8900",
    "email": "john@example.com",
    "favorite": false,
    "created_at": "2025-10-16T10:00:00Z",
    "updated_at": "2025-10-16T10:00:00Z"
  }
}
```

**Error Response (404):**
```json
{
  "status": 0,
  "status_code": 404,
  "message": "Contact not found",
  "data": {}
}
```

**cURL Example:**
```bash
curl -X GET http://localhost:9001/api/v1/contacts/5 \
  -H "Authorization: Bearer <access_token>"
```

---

## 8. Update Contact

**Endpoint:** `PUT /api/v1/contacts/{id}`

**Description:** Update existing contact

**Request Headers:**
```
Authorization: Bearer <access_token>
Content-Type: application/json
```

**Request Body:**
```json
{
  "full_name": "Johnathan Doe",
  "phone": "0898209890",
  "email": null,
  "favorite": true
}
```

**Success Response (200):**
```json
{
  "status": 1,
  "status_code": 200,
  "message": "Contact updated successfully",
  "data": {
    "id": 5,
    "full_name": "Johnathan Doe",
    "phone": "0898209890",
    "email": null,
    "favorite": true,
    "created_at": "2025-10-16T10:00:00Z",
    "updated_at": "2025-10-16T10:30:00Z"
  }
}
```

**Error Response (400):**
```json
{
  "status": 0,
  "status_code": 400,
  "message": "Validation error",
  "data": {
    "phone": ["invalid format"]
  }
}
```

**cURL Example:**
```bash
curl -X PUT http://localhost:9001/api/v1/contacts/5 \
  -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "full_name": "Johnathan Doe",
    "phone": "0898209890",
    "email": null,
    "favorite": true
  }'
```

---

## 9. Delete Contact

**Endpoint:** `DELETE /api/v1/contacts/{id}`

**Description:** Delete contact by ID

**Request Headers:**
```
Authorization: Bearer <access_token>
```

**Success Response (200):**
```json
{
  "status": 1,
  "status_code": 200,
  "message": "Contact deleted successfully",
  "data": {}
}
```

**Error Response (404):**
```json
{
  "status": 0,
  "status_code": 404,
  "message": "Contact not found",
  "data": {}
}
```

**cURL Example:**
```bash
curl -X DELETE http://localhost:9001/api/v1/contacts/5 \
  -H "Authorization: Bearer <access_token>"
```

---

## Authentication Flow

### 1. Register/Login
```
Client -> POST /api/v1/auth/register
       <- 201 {token: "..."}
```

### 2. Store Token
```javascript
localStorage.setItem('access_token', response.data.token.access_token);
```

### 3. Use Token for Protected Endpoints
```
Client -> GET /api/v1/me
          Authorization: Bearer <token>
       <- 200 {user data}
```

---

## Middleware

### AuthMiddleware

**Purpose:** Validates JWT token and sets userID in context

**Flow:**
1. Extract Authorization header
2. Check "Bearer " prefix
3. Parse and validate JWT token
4. Set userID in Gin context
5. Allow request to proceed

**Usage in Routes:**
```go
api.GET("/me", authMiddleware, handler.GetProfile)
```

**Error Responses:**
- Missing token: 401 "Unauthorized - missing token"
- Invalid format: 401 "Unauthorized - invalid token format"
- Expired/Invalid: 401 "Unauthorized - invalid or expired token"

---

## Error Handling

### Error Types

**400 Bad Request:**
- Invalid request body
- Validation errors
- Invalid parameters

**401 Unauthorized:**
- Missing token
- Invalid token
- Expired token

**403 Forbidden:**
- Accessing other user's resources

**404 Not Found:**
- User not found
- Contact not found

**409 Conflict:**
- Email already exists
- Phone already exists

**500 Internal Server Error:**
- Database errors
- Unexpected errors

### Error Response Format

**Validation Error:**
```json
{
  "status": 0,
  "status_code": 400,
  "message": "Validation error",
  "data": {
    "field_name": ["error message 1", "error message 2"]
  }
}
```

**Generic Error:**
```json
{
  "status": 0,
  "status_code": 500,
  "message": "Error message",
  "data": {}
}
```

---

## Testing Endpoints

### Using cURL

**1. Register:**
```bash
curl -X POST http://localhost:9001/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"full_name":"Test User","email":"test@example.com","phone":"081234567890","password":"password123"}'
```

**2. Login and Save Token:**
```bash
TOKEN=$(curl -X POST http://localhost:9001/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}' \
  | jq -r '.data.token.access_token')
```

**3. Get Profile:**
```bash
curl -X GET http://localhost:9001/api/v1/me \
  -H "Authorization: Bearer $TOKEN"
```

**4. Create Contact:**
```bash
curl -X POST http://localhost:9001/api/v1/contacts \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"full_name":"John Doe","phone":"081234567890"}'
```

**5. List Contacts:**
```bash
curl -X GET "http://localhost:9001/api/v1/contacts?page=1&limit=10" \
  -H "Authorization: Bearer $TOKEN"
```

### Using Postman

1. **Create Collection:** Contact Management API
2. **Set Collection Variable:** `base_url = http://localhost:9001/api/v1`
3. **Register/Login:** Save token from response
4. **Set Token:** In Authorization tab, select "Bearer Token"
5. **Test All Endpoints**

---

## Handler Implementation Details

### Helper Functions

**successResponse:**
```go
func (h *Handler) successResponse(c *gin.Context, statusCode int, message string, data interface{})
```

**errorResponse:**
```go
func (h *Handler) errorResponse(c *gin.Context, statusCode int, message string, data interface{})
```

**validationErrorResponse:**
```go
func (h *Handler) validationErrorResponse(c *gin.Context, field string, messages []string)
```

### Request Binding

Using Gin's `ShouldBindJSON` for automatic validation:
```go
var req models.RegisterRequest
if err := c.ShouldBindJSON(&req); err != nil {
    h.errorResponse(c, http.StatusBadRequest, "Invalid request body", gin.H{})
    return
}
```

### Context Usage

**Get User ID from Context:**
```go
userID, exists := c.Get("userID")
if !exists {
    h.errorResponse(c, http.StatusUnauthorized, "Unauthorized", gin.H{})
    return
}
```

**Use in Service Call:**
```go
profile, err := h.service.GetProfile(c.Request.Context(), userID.(uint))
```

---

## Best Practices

### 1. Always Use Context
```go
// ✅ Good
h.service.GetProfile(c.Request.Context(), userID)

// ❌ Bad
h.service.GetProfile(context.Background(), userID)
```

### 2. Consistent Error Handling
```go
if errors.Is(err, service.ErrContactNotFound) {
    h.errorResponse(c, http.StatusNotFound, "Contact not found", gin.H{})
    return
}
```

### 3. Validate User Authorization
```go
userID, exists := c.Get("userID")
if !exists {
    h.errorResponse(c, http.StatusUnauthorized, "Unauthorized", gin.H{})
    return
}
```

### 4. Use Helper Functions
```go
// Instead of repeating JSON responses
h.successResponse(c, http.StatusOK, "Success", data)
```

---

## Next Steps

After handlers are complete:

1. **Integration Testing** - Test all endpoints with real database
2. **API Documentation** - Generate Swagger/OpenAPI docs
3. **Rate Limiting** - Add rate limiting middleware
4. **Logging** - Enhance logging with structured logs
5. **Monitoring** - Add metrics and health checks

---

## References

- [Gin Web Framework](https://gin-gonic.com/)
- [HTTP Status Codes](https://developer.mozilla.org/en-US/docs/Web/HTTP/Status)
- [REST API Best Practices](https://restfulapi.net/)
