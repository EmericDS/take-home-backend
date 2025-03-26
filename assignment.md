# The Project

The goal of this project is to build a simple Go HTTP server that exposes the following API:
- an endpoint to upload documents (in raw text format)
- an endpoint to list uploaded documents
- an endpoint to download previously uploaded documents

## Upload Document Endpoint

Uploaded files must be stored on the local filesystem in a dedicated directory.

File metadata (an internal identifier, its name as well as its upload date) must be stored in a PostgreSQL database.

## List Documents Endpoint

The endpoint to list documents must output an array of the following JSON payload:

```json
{
    "id": "c7abc621-3bc8-42d0-bf8b-4d348cbcbc41",
    "name": "file.txt",
    "url": "http://localhost/dl/file.txt",
    "uploaded_at": "2009-11-10T23:00:00Z"
}
```

The URL returned in the list must allow to download the file.

# Going Further

Here are some ideas to go further, in no particular order:
- instead of storing documents on the local filesystem, use an S3-compatible object storage solution (e.g., [MinIO](https://min.io/docs/minio/container/index.html))
- add monitoring/telemetry
- add authentication
- add tests

Of course, these are just ideas and are optional. Feel free to add anything you want, or nothing. Keep in mind we much prefer a strong base rather than a bunch of poorly implemented features.

# What We Will Look For

We will mainly evaluate the following points:
- overall quality and readability
- maintainability
- documentation

# Getting started

You can either start from scratch by yourself or use the attached sources to have a Dockerized bootstrap environment.
