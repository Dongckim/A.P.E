# A.P.E API Documentation

Base URL: `http://localhost:9000/api`

All responses follow the format:
```json
{
  "data": { ... },
  "error": ""
}
```

On error:
```json
{
  "data": null,
  "error": "error message here"
}
```

---

## EC2 File Operations

### List directory

```
GET /api/ec2/files?path=/home/ubuntu&connection_id=conn1
```

**Response:**
```json
{
  "data": {
    "path": "/home/ubuntu",
    "items": [
      {
        "name": "app.py",
        "path": "/home/ubuntu/app.py",
        "size": 2048,
        "is_dir": false,
        "permissions": "-rw-r--r--",
        "modified_at": "2025-03-28T10:30:00Z"
      },
      {
        "name": "logs",
        "path": "/home/ubuntu/logs",
        "size": 4096,
        "is_dir": true,
        "permissions": "drwxr-xr-x",
        "modified_at": "2025-03-27T15:00:00Z"
      }
    ]
  }
}
```

### Read file content

```
GET /api/ec2/file?path=/home/ubuntu/app.py&connection_id=conn1
```

**Response:**
```json
{
  "data": {
    "path": "/home/ubuntu/app.py",
    "content": "import flask\n\napp = flask.Flask(__name__)\n...",
    "size": 2048,
    "encoding": "utf-8"
  }
}
```

### Save file content

```
PUT /api/ec2/file
Content-Type: application/json

{
  "connection_id": "conn1",
  "path": "/home/ubuntu/app.py",
  "content": "updated file content here"
}
```

### Upload file

```
POST /api/ec2/upload?path=/home/ubuntu/uploads&connection_id=conn1
Content-Type: multipart/form-data

file: (binary)
```

### Download file

```
GET /api/ec2/download?path=/home/ubuntu/app.py&connection_id=conn1
```

Returns binary file stream with appropriate Content-Disposition header.

### Delete file or folder

```
DELETE /api/ec2/file?path=/home/ubuntu/old-file.txt&connection_id=conn1
```

### Rename / Move

```
PATCH /api/ec2/rename
Content-Type: application/json

{
  "connection_id": "conn1",
  "old_path": "/home/ubuntu/old-name.txt",
  "new_path": "/home/ubuntu/new-name.txt"
}
```

### Create folder

```
POST /api/ec2/folder
Content-Type: application/json

{
  "connection_id": "conn1",
  "path": "/home/ubuntu/new-folder"
}
```

---

## S3 Operations

### List buckets

```
GET /api/s3/buckets
```

**Response:**
```json
{
  "data": {
    "buckets": [
      {
        "name": "my-app-assets",
        "region": "us-east-1",
        "created_at": "2024-01-15T08:00:00Z"
      }
    ]
  }
}
```

### List objects

```
GET /api/s3/objects?bucket=my-app-assets&prefix=images/
```

**Response:**
```json
{
  "data": {
    "bucket": "my-app-assets",
    "prefix": "images/",
    "items": [
      {
        "key": "images/logo.png",
        "size": 45000,
        "is_folder": false,
        "last_modified": "2025-03-20T12:00:00Z",
        "storage_class": "STANDARD"
      },
      {
        "key": "images/icons/",
        "size": 0,
        "is_folder": true,
        "last_modified": null,
        "storage_class": ""
      }
    ]
  }
}
```

### Upload object

```
POST /api/s3/upload?bucket=my-app-assets&key=images/new-logo.png
Content-Type: multipart/form-data

file: (binary)
```

### Download object

```
GET /api/s3/download?bucket=my-app-assets&key=images/logo.png
```

### Delete object

```
DELETE /api/s3/object?bucket=my-app-assets&key=images/old-logo.png
```

---

## Connection Management

### List active connections

```
GET /api/connections
```

**Response:**
```json
{
  "data": {
    "connections": [
      {
        "id": "conn1",
        "host": "54.123.45.67",
        "user": "ubuntu",
        "status": "connected",
        "connected_at": "2025-03-28T09:00:00Z"
      }
    ]
  }
}
```

### Add connection

```
POST /api/connections
Content-Type: application/json

{
  "host": "54.123.45.67",
  "user": "ubuntu",
  "key_path": "~/.ssh/my-key.pem",
  "port": 22
}
```

### Disconnect

```
DELETE /api/connections/conn1
```
