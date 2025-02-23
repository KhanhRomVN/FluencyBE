# Grammar API Documentation

This document outlines the available API endpoints for the Grammar module in the Fluency application.

## Base URL
All endpoints are relative to `/api/v1/grammar`

## Question Types
The grammar module supports multiple types of questions:
- Fill in the blank
- Choice one (Multiple choice)
- Error identification
- Sentence transformation

## Endpoints

### Create Grammar Question
Creates a new grammar question of any supported type.

**Endpoint:** `POST /questions`

**Request Body:**
```json
{
  "type": "string",
  "topic": "string",
  "instruction": "string",
  "imageUrls": ["string"],
  "maxTime": int
}
```

**Response:** `201 Created`
```json
{
  "id": "uuid",
  "type": "string",
  "topic": "string",
  "instruction": "string",
  "imageUrls": ["string"],
  "maxTime": int,
  "version": int
}
```

### Get Grammar Question Detail
Retrieves detailed information about a specific grammar question.

**Endpoint:** `GET /questions/:id`

**Response:** `200 OK`
```json
{
  "id": "uuid",
  "type": "string",
  "topic": "string",
  "instruction": "string",
  "imageUrls": ["string"],
  "maxTime": int,
  "version": int
}
```

### Update Grammar Question
Updates specific fields of a grammar question.

**Endpoint:** `PUT /questions/:id`

**Request Body:**
```json
{
  "topic": "string",
  "instruction": "string",
  "imageUrls": ["string"],
  "maxTime": int
}
```

**Response:** `204 No Content`

### Delete Grammar Question
Deletes a specific grammar question.

**Endpoint:** `DELETE /questions/:id`

**Response:** `204 No Content`

### Get Updated Grammar Questions
Retrieves a list of grammar questions that have been updated since a specific version.

**Endpoint:** `POST /questions/updates`

**Request Body:**
```json
{
  "questions": [
    {
      "grammarQuestionId": "uuid",
      "version": int
    }
  ]
}
```

**Response:** `200 OK`
```json
{
  "success": true,
  "data": [
    {
      "id": "uuid",
      "type": "string",
      "topic": "string",
      "instruction": "string",
      "imageUrls": ["string"],
      "maxTime": int
    }
  ]
}
```

### Get Multiple Grammar Questions by IDs
Retrieves multiple grammar questions by their IDs.

**Endpoint:** `POST /questions/list`

**Request Body:**
```json
{
  "question_ids": ["uuid"]
}
```

**Response:** `200 OK`
```json
{
  "success": true,
  "data": [
    {
      "id": "uuid",
      "type": "string",
      "topic": "string",
      "instruction": "string",
      "imageUrls": ["string"],
      "maxTime": int
    }
  ]
}
```

### Search Grammar Questions
Search and filter grammar questions with pagination.

**Endpoint:** `GET /questions/search`

**Query Parameters:**
- `page`: int (optional)
- `limit`: int (optional)
- `type`: string (optional)
- `topic`: string (optional)
- `instruction`: string (optional)

**Response:** `200 OK`
```json
{
  "success": true,
  "data": {
    "questions": [
      {
        "id": "uuid",
        "type": "string",
        "topic": "string",
        "instruction": "string",
        "imageUrls": ["string"],
        "maxTime": int
      }
    ],
    "total": int,
    "page": int,
    "limit": int
  }
}
```

### Delete All Grammar Data
Deletes all grammar questions and related data.

**Endpoint:** `DELETE /questions/all`

**Response:** `200 OK`
```json
{
  "message": "All grammar data deleted successfully"
}
```

## Error Responses

All endpoints may return the following error responses:

- `400 Bad Request`: Invalid request format or parameters
- `404 Not Found`: Requested resource not found
- `500 Internal Server Error`: Server-side error

Error response format:
```json
{
  "error": "Error message"
}