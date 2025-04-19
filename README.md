# Go REST API

# REST API Architecture Overview

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
