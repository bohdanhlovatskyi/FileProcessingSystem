## System for uploading and processing files

- consists of two microservices:
  - file uploading and retrieval (Files API)
  - file processing (Processing API)

### Files API
- exposes HTTP endpoint for uploading files
- once file is uploaded, it should be saved to the file system, and its ID should be sent to the Processing API via a RabbitMQ queue.
  
### Processing API
- accepts file ID from a RabbitMQ queue
- reduces the image size
- overwrites the file in the file system

