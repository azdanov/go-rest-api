# Go REST API

## Requirements

- [Go 1.24](https://golang.org/dl/)
- [Docker](https://www.docker.com/)
- [Docker Compose](https://docs.docker.com/compose/)
- [Task](https://taskfile.dev/)

## REST API Architecture Overview

```mermaid
flowchart

ExternalRequest([External Request]) --> A[HTTP Layer]
A[HTTP Layer]
B[Service Layer]
C[Repository Layer]
D["Client (External API)"]
E[(Database)]

A --> B
B --> C
B --> D
C --> E
```
