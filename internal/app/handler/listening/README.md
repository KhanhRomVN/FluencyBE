# Listening API Documentation

This document outlines the available API endpoints for the Listening module in the Fluency application.

## Base URL
All endpoints are relative to `/api/v1/listening`

## Question Types
The listening module supports multiple types of questions:
- Fill in the blank
- Choice one (Single choice)
- Choice multi (Multiple choice)
- Map labelling
- Matching

## Endpoints

### Create Listening Question
Creates a new listening question of any supported type.

**Endpoint:** `POST /questions`

**Request Body:**
```json
{
  "type": "string",
  "topic": "string",
  "instruction": "string",
  "audioUrls": ["string"],
  "imageUrls": ["string"],
  "transcript": "string",
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
  "audioUrls": ["string"],
  "imageUrls": ["string"],
  "transcript": "string",
  "maxTime": int,
  "version": int
}
```

### Get Listening Question Detail
Retrieves detailed information about a specific listening question.

**Endpoint:** `GET /questions/:id`

**Response:** `200 OK`
```json
{
  "id": "uuid",
  "type": "string",
  "topic": "string",
  "instruction": "string",
  "audioUrls": ["string"],
  "imageUrls": ["string"],
  "transcript": "string",
  "maxTime": int,
  "version": int
}
```

### Update Listening Question
Updates specific fields of a listening question.

**Endpoint:** `PUT /questions/:id`

**Request Body:**
```json
{
  "topic": "string",
  "instruction": "string",
  "audioUrls": ["string"],
  "imageUrls": ["string"],
  "transcript": "string",
  "maxTime": int
}
```

**Response:** `204 No Content`

### Delete Listening Question
Deletes a specific listening question.

**Endpoint:** `DELETE /questions/:id`

**Response:** `204 No Content`

### Get Updated Listening Questions
Retrieves a list of listening questions that have been updated since a specific version.

**Endpoint:** `POST /questions/updates`

**Request Body:**
```json
{
  "questions": [
    {
      "listeningQuestionId": "uuid",
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
      "audioUrls": ["string"],
      "imageUrls": ["string"],
      "transcript": "string",
      "maxTime": int
    }
  ]
}
```

### Get Multiple Listening Questions by IDs
Retrieves multiple listening questions by their IDs.

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
      "audioUrls": ["string"],
      "imageUrls": ["string"],
      "transcript": "string",
      "maxTime": int
    }
  ]
}
```

### Search Listening Questions
Search and filter listening questions with pagination.

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
        "audioUrls": ["string"],
        "imageUrls": ["string"],
        "transcript": "string",
        "maxTime": int
      }
    ],
    "total": int,
    "page": int,
    "limit": int
  }
}
```

### Delete All Listening Data
Deletes all listening questions and related data.

**Endpoint:** `DELETE /questions/all`

**Response:** `200 OK`
```json
{
  "message": "All listening data deleted successfully"
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