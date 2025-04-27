# Online Survey Platform API Server

A RESTful API server built with Go and MongoDB for creating, editing, deleting and collecting responses for surveys. Surveys are uniquely identified by a 5-character token for public access, support multiple question types (Textbox, Multiple Choice, Likert Scale), and store participant responses with user identification.

## Table of Contents
- [Features](#features)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Configuration](#configuration)
- [Running the Server](#running-the-server)
- [API Endpoints](#api-endpoints)
- [Data Structures](#data-structures)
- [Example Usage](#example-usage)


## Features
- Create, update, delete, and retrieve surveys with unique 5-character tokens
- Public survey access via token
- Response collection with user identification
- Paginated survey list retrieval
- Environment variable configuration via `.env`
- Input validation for question types and responses

## Prerequisites
- **Go**: Version 1.16 or higher
- **MongoDB**: Version 4.4 or higher (local or cloud instance)
- **Git**: For cloning the repository
- **Dependencies**:
  - `go.mongodb.org/mongo-driver/v2`
  - `github.com/gorilla/mux`
  - `github.com/joho/godotenv`

## Installation
1. Clone the repository:
   ```bash
   git clone https://github.com/r7chardgh/OSP_backend.git
   cd OSP_backend
   ```

2. Install dependencies:
   ```bash
   go get go.mongodb.org/mongo-driver/v2/mongo
   go get github.com/gorilla/mux
   go get github.com/joho/godotenv
   ```

3. Ensure MongoDB is running:
   - For a local instance, start MongoDB (default port: 27017).
   - For a cloud instance (e.g., MongoDB Atlas), obtain the connection URI.

## Configuration
1. Create a `.env` file in the project root:
   ```env
   MONGODB_URI=mongodb://localhost:27017
   ```
   Replace `mongodb://localhost:27017` with your MongoDB connection string if using a remote instance.

   You may want to find the quick-start guide from MongoDB official Website
   https://www.mongodb.com/docs/drivers/go/current/quick-start/

2. The API connects to a database named `OSP_backend` with collections `surveys` and `responses`.

## Running the Server
1. Start the server:
   ```bash
   go run main.go
   ```
2. The server will be available at `http://localhost:5050`.

## API Endpoints
All endpoints return JSON responses and expect JSON payloads where applicable. The base URL is `http://localhost:5050`.

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/surveys?page={page}&limit={limit}` | List all surveys (paginated) |
| `POST` | `/surveys` | Create a new survey |
| `PUT` | `/surveys/{survey_id}` | Update an existing survey |
| `DELETE` | `/surveys/{survey_id}` | Delete a survey |
| `GET` | `/surveys/token/{token}` | Retrieve a survey by token |
| `POST` | `/responses/{survey_id}` | Submit responses for a survey |
| `GET` | `/responses` | Get all responses across surveys |
| `GET` | `/responses/{survey_id}` | Get responses for a specific survey |

### Endpoint Details

#### GET /surveys
List all surveys with pagination.
- **Query Parameters**:
  - `page` (int, optional): Page number (default: 1)
  - `limit` (int, optional): Items per page (default: 10)
- **Response**: `200 OK`
  ```json
  [
      {
          "token": "aB2c9",
          "title": "Employee Feedback"
      }
  ]
  ```

#### POST /surveys
Create a new survey.
- **Body**:
  ```json
  {
      "title": "string",
      "questions": [
          {
              "question_title": "string",
              "question_type": "Textbox|Multiple Choice|Likert Scale",
              "answers": ["string"]
          }
      ]
  }
  ```
- **Response**: `201 Created`
  ```json
  {
      "id": "ObjectID",
      "title": "string",
      "token": "string",
      "created_at": "timestamp",
      "updated_at": "timestamp",
      "questions": [
        {"question_title":"string","question_type":"string","answers":["string"]}
      ]
  }
  ```

#### PUT /surveys/{survey_id}
Update a survey by ID.
- **Path Parameters**:
  - `survey_id` (ObjectID): Survey ID
- **Body**:
  ```json
  {
      "title": "string",
      "questions": [
        {"question_title":"string","question_type":"string","answers":["string"]}
      ]
  }
  ```
- **Response**: `200 OK`
  ```json
  { "message": "survey updated" }
  ```

#### DELETE /surveys/{survey_id}
Delete a survey by ID.
- **Path Parameters**:
  - `survey_id` (ObjectID): Survey ID
- **Response**: `200 OK`
  ```json
  { "message": "survey deleted" }
  ```

#### GET /surveys/token/{token}
Retrieve a survey by its public token.
- **Path Parameters**:
  - `token` (string): 5-character survey token
- **Response**: `200 OK`
  ```json
  {
      "id": "ObjectID",
      "title": "string",
      "token": "string",
      "created_at": "timestamp",
      "updated_at": "timestamp",
      "questions": [
        {"question_title":"string","question_type":"string","answers":["string"]}
      ]
  }
  ```

#### POST /responses/{survey_id}
Submit responses for a survey.
- **Path Parameters**:
  - `survey_id` (ObjectID): Survey ID
- **Body**:
  ```json
  [
      {
          "question_id": "ObjectID",
          "response_text": "string"
      }
  ]
  ```
- **Response**: `201 Created`
  ```json
  [
      {
          "question_id": "ObjectID",
          "response_text": "string"
      }
  ]
  ```

#### GET /responses
Retrieve all responses across all surveys.
- **Response**: `200 OK`
  ```json
  [
      {
          "id": "ObjectID",
          "user_id": "ObjectID",
          "created_at": "timestamp",
          "survey_id": "ObjectID",
          "question_id": "ObjectID",
          "response_text": "string"
      }
  ]
  ```

#### GET /responses/{survey_id}
Retrieve responses for a specific survey.
- **Path Parameters**:
  - `survey_id` (ObjectID): Survey ID
- **Response**: `200 OK`
  ```json
  [
      {
          "id": "ObjectID",
          "user_id": "ObjectID",
          "created_at": "timestamp",
          "survey_id": "ObjectID",
          "question_id": "ObjectID",
          "response_text": "string"
      }
  ]
  ```

## Data Structures

### Survey
```json
{
    "id": "ObjectID",
    "token": "string (5 characters)",
    "created_at": "timestamp",
    "updated_at": "timestamp",
    "title": "string",
    "questions": [
        {
            "id": "ObjectID",
            "question_title": "string",
            "question_type": "Textbox|Multiple Choice|Likert Scale",
            "answers": ["string"]
        }
    ]
}
```

### SurveysList
```json
{
    "token": "string",
    "title": "string"
}
```

### Response
```json
{
    "id": "ObjectID",
    "user_id": "ObjectID",
    "created_at": "timestamp",
    "survey_id": "ObjectID",
    "question_id": "ObjectID",
    "response_text": "string"
}
```

### ResponseInput
```json
{
    "question_id": "ObjectID",
    "response_text": "string"
}
```

## Example Usage
Below are example `curl` commands for interacting with the API.
Download curl from Official website https://curl.se/download.html if you would like to follow the guide
or
Use Any RESTful API client app to test it, e.g. Postman, Thunder Client (VSCode plugin)

### Create a Survey
```bash
curl -X POST http://localhost:5050/surveys \
-H "Content-Type: application/json" \
-d '{
    "title": "My Survey",
    "questions": [
        {
            "question_title": "How satisfied are you with our service?",
            "question_type": "Likert Scale",
            "answers": ["1", "2", "3", "4", "5"]
        },
        {
            "question_title": "What is your favorite frontend framework?",
            "question_type": "Multiple Choice",
            "answers": ["React", "Vue", "Next", "Nuxt"]
        }
    ]
}'
```

**Response**:
```json
{
    "id": "507f1f77bcf86cd799439011",
    "token": "Xy2aB",
    "created_at": "2025-04-27T10:00:00Z",
    "updated_at": "2025-04-27T10:00:00Z",
    "title": "My Survey",
    "questions": [...]
}
```

### Get Survey by Token
```bash
curl http://localhost:5050/surveys/token/Xy2aB
```

**Response**:
```json
{
    "id": "507f1f77bcf86cd799439011",
    "token": "Xy2aB",
    "created_at": "2025-04-27T10:00:00Z",
    "updated_at": "2025-04-27T10:00:00Z",
    "title": "My Survey",
    "questions": [...]
}
```

### Submit Responses by knowing survey id
```bash
curl -X POST http://localhost:5050/responses/507f1f77bcf86cd799439011 \
-H "Content-Type: application/json" \
-d '[
    {
        "question_id": "507f1f77bcf86cd799439012",
        "response_text": "4"
    },
    {
        "question_id": "507f1f77bcf86cd799439013",
        "response_text": "Next"
    }
]'
```

**Response**:
```json
[
    {
        "question_id": "507f1f77bcf86cd799439012",
        "response_text": "4"
    },
    {
        "question_id": "507f1f77bcf86cd799439013",
        "response_text": "Next"
    }
]
```

### List Surveys (page no. 1 and 10 items shown on one page)
```bash
curl http://localhost:5050/surveys?page=1&limit=10
```

**Response**:
```json
[
    {
        "token": "Xy2aB",
        "title": "My Survey"
    }
]
```
