# Document Upload Service

A simple Go HTTP server that allows users to upload, list, and download text documents.

## Features

- Upload documents: Store files on the server with metadata in PostgreSQL
- List documents: Get a JSON list of all uploaded files with their metadata
- Download documents: Retrieve previously uploaded files

## Getting Started

### Prerequisites

- Docker and Docker Compose
- An internet connection to download dependencies

### Running the Service

Clone the repository and start the service with Docker Compose:


```bash
docker compose down
docker-compose up --build
```

The service will be available at http://localhost:8080

## API Documentation

### Homepage

- **URL**: `/`
- **Method**: `GET`
- **Description**: Displays a simple HTML form for uploading files

### Upload Document

- **URL**: `/upload`
- **Method**: `POST`
- **Content-Type**: `multipart/form-data`
- **Parameter**: `file` - The file to upload
- **Response**: `201 Created` on success
- **Example**:
  ```bash
  curl -X POST -F "file=@/path/to/file.txt" http://localhost:8080/upload
  ```

You can also use the HTML interface in the web page to upload documents at http://localhost:8080

### List Documents

- **URL**: `/documents`
- **Method**: `GET`
- **Response**: JSON array of document objects
- **Response Format**:
  ```json
  [
    {
      "id": "c7abc621-3bc8-42d0-bf8b-4d348cbcbc41",
      "name": "file.txt",
      "url": "http://localhost:8080/dl/c7abc621-3bc8-42d0-bf8b-4d348cbcbc41",
      "uploaded_at": "2009-11-10T23:00:00Z"
    }
  ]
  ```
- **Example**:
  ```bash
  curl http://localhost:8080/documents
  ```

### Download Document

- **URL**: `/dl/{id}`
- **Method**: `GET`
- **Parameters**: `id` - The unique identifier of the document
- **Response**: The file content with appropriate headers for download
- **Example**:
  ```bash
  curl -O http://localhost:8080/dl/c7abc621-3bc8-42d0-bf8b-4d348cbcbc41
  ```

## Project Structure

- `main.go`: Main application code
- `docker-compose.yml`: Docker Compose configuration
- `Dockerfile`: Docker build configuration

## Technical Details

### Storage

- Files are stored on the local filesystem in the `/app/uploads` directory
- File metadata (ID, name, upload date) is stored in PostgreSQL

### Database Schema

```sql
CREATE TABLE documents (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    uploaded_at TIMESTAMP WITH TIME ZONE NOT NULL
)
```
