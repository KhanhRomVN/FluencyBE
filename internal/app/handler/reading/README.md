# Reading API Documentation

This document outlines the available API endpoints for the Reading module in the Fluency application.

## Base URL
All endpoints are relative to `/api/v1/reading`

## Question Types
The reading module supports multiple types of questions:
- Fill in the blank
- Choice one (Single choice)
- Choice multi (Multiple choice)
- True/False
- Matching

## Endpoints

### Create Reading Question
Creates a new reading question of any supported type.

**Endpoint:** `POST /questions`

**Request Body:**
```json
{
  "type": "string",
  "topic": "string",
  "instruction": "string",
  "title": "string",
  "passages": ["string"],
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
  "title": "string",
  "passages": ["string"],
  "imageUrls": ["string"],
  "maxTime": int,
  "version": int
}
```

### Get Reading Question Detail
Retrieves detailed information about a specific reading question, including its associated sub-questions based on type.

**Endpoint:** `GET /questions/:id`

**Response:** `200 OK`
```json
{
  "id": "uuid",
  "type": "string",
  "topic": "string",
  "instruction": "string",
  "title": "string",
  "passages": ["string"],
  "imageUrls": ["string"],
  "maxTime": int,
  "version": int,
  "fillInTheBlankQuestion": {
    "id": "uuid",
    "question": "string",
    "answers": [
      {
        "id": "uuid",
        "answer": "string",
        "explain": "string"
      }
    ]
  },
  "choiceOneQuestion": {
    "id": "uuid",
    "question": "string",
    "explain": "string",
    "options": [
      {
        "id": "uuid",
        "options": "string",
        "isCorrect": boolean
      }
    ]
  },
  "choiceMultiQuestion": {
    "id": "uuid",
    "question": "string",
    "explain": "string",
    "options": [
      {
        "id": "uuid",
        "options": "string",
        "isCorrect": boolean
      }
    ]
  },
  "trueFalse": [
    {
      "id": "uuid",
      "question": "string",
      "answer": boolean,
      "explain": "string"
    }
  ],
  "matching": [
    {
      "id": "uuid",
      "question": "string",
      "answer": "string",
      "explain": "string"
    }
  ]
}
```

### Update Reading Question
Updates specific fields of a reading question.

**Endpoint:** `PUT /questions/:id`

**Request Body:**
```json
{
  "topic": "string",
  "instruction": "string",
  "title": "string",
  "passages": ["string"],
  "imageUrls": ["string"],
  "maxTime": int
}
```

**Response:** `204 No Content`

### Delete Reading Question
Deletes a specific reading question.

**Endpoint:** `DELETE /questions/:id`

**Response:** `204 No Content`

### Get Updated Reading Questions
Retrieves a list of reading questions that have been updated since a specific version.

**Endpoint:** `POST /questions/updates`

**Request Body:**
```json
{
  "questions": [
    {
      "readingQuestionId": "uuid",
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
      "title": "string",
      "passages": ["string"],
      "imageUrls": ["string"],
      "maxTime": int,
      "version": int
    }
  ]
}
```

### Get Multiple Reading Questions by IDs
Retrieves multiple reading questions by their IDs.

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
      "title": "string",
      "passages": ["string"],
      "imageUrls": ["string"],
      "maxTime": int,
      "version": int
    }
  ]
}
```

### Search Reading Questions
Search and filter reading questions with pagination.

**Endpoint:** `GET /questions/search`

**Query Parameters:**
- `page`: int (optional)
- `limit`: int (optional)
- `type`: string (optional)
- `topic`: string (optional)
- `instruction`: string (optional)
- `title`: string (optional)
- `passages`: string (optional)

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
        "title": "string",
        "passages": ["string"],
        "imageUrls": ["string"],
        "maxTime": int,
        "version": int
      }
    ],
    "total": int,
    "page": int,
    "limit": int
  }
}
```

### Delete All Reading Data
Deletes all reading questions and related data.

**Endpoint:** `DELETE /questions/all`

**Response:** `200 OK`
```json
{
  "message": "All reading data deleted successfully"
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