# GoTinyLink - URL Shortener

## Overview

GoTinyLink is a URL Shortener application built with a **Golang backend** and a **Next.js frontend**. It provides functionality to shorten URLs, track their access count, set expiration dates, and display statistics. The application uses **PostgreSQL** as its database, **Redis** for caching, and ensures smooth frontend-backend communication with **CORS** support.

---

## Features

### Backend Features

- Create unique shortened URLs.
- Redirect to the original URL when accessing a short code.
- Track the number of accesses for each shortened URL.
- Update and delete existing shortened URLs.
- Automatically remove expired URLs using a scheduled task.
- Cache frequently accessed URLs with Redis for better performance.
- Retrieve detailed statistics for each shortened URL.
- Rate limiting with IP blacklist to prevent abuse and ensure fair usage.

### Frontend Features

- User-friendly interface to create and manage URLs.
- Display original URLs, their short codes, creation timestamps, and access counts.

---

## System Architecture

1. **Frontend**: Built with Next.js and interacts with the backend using API calls.
2. **Backend**: Manages API requests, database operations, and caching.
3. **PostgreSQL Database**: Stores URL data along with expiration timestamps.
4. **Redis Cache**: Caches frequently accessed URLs for faster redirection.

---

## Repository Structure

```
GoTinyLink/
├── backend/            # Backend (Golang) service
├── frontend/           # Frontend (Next.js) application
├── README.md           # Project documentation
└── LICENSE             # License
```

---

## Backend Setup

### Prerequisites

- **Go 1.20+**
- **PostgreSQL 12+**
- **Redis 6+**
- `.env` file containing:
  ```env
  DATABASE_URL=<your_postgresql_connection_string>
  REDIS_HOST=<your_redis_host>
  REDIS_PORT=<your_redis_port>
  REDIS_PASSWORD=<your_redis_password>
  PORT=8080
  HASH_SALT=your_hash_salt
  ALLOW_ORIGINS=your_allow_origins
  ```

### Installation and Setup

1. Navigate to the backend directory:
   ```bash
   cd server
   ```
2. Copy the example environment file to .env:
   ```bash
   cp .env.example .env
   ```
3. Install dependencies:
   ```bash
   go mod tidy
   ```
4. Run the backend service:
   ```bash
   go run main.go
   ```
   The server runs on `http://localhost:8080` by default.

### Scheduled Tasks

Expired URLs are automatically cleaned using a scheduled task implemented with Go's `time.Ticker`. You can customize the cleanup interval in the backend configuration.

---

## Frontend Setup

### Prerequisites

- **Node.js 16+**
- **npm** or **yarn**
- `.env` file in the frontend directory containing:
  ```env
  NEXT_PUBLIC_API_BASE_URL=http://localhost:8080
  NEXT_PUBLIC_DOMAIN_URL=http://localhost:3000
  ```

### Installation and Setup

1. Navigate to the frontend directory:
   ```bash
   cd go-tinyLink
   ```
2. Copy the example environment file to .env:
   ```bash
   cp .env.example .env
   ```
3. Install dependencies:
   ```bash
   npm install
   # or
   yarn install
   ```
4. Run the development server:
   ```bash
   npm run dev
   # or
   yarn dev
   ```
   The application runs on `http://localhost:3000` by default.

---

## API Endpoints

### Base URL

`http://localhost:8080`

### Endpoints

- **POST /shorten**

  - Create a new shortened URL.
  - Payload:
    ```json
    {
      "url": "https://exampleHaha/fasfasvdcdcda.com"
    }
    ```
  - Response:
    ```json
    {
      "success": true,
      "url": "https://exampleHaha/fasfasvdcdcda.com",
      "shortCode": "abc123",
      "accessCount": 0,
      "expiresAt": "2023-10-10T10:00:00Z"
    }
    ```

- **GET /shorten/:shortCode**

  - Redirect to the original URL.

- **GET /shorten/:shortCode/stats**

  - Get statistics for a shortened URL.

- **PUT /shorten/:shortCode**

  - Update the original URL.
  - Payload:
    ```json
    {
      "url": "https://exampleHaha/fasfasvdcdcda.com"
    }
    ```

- **DELETE /shorten/:shortCode**
  - Delete the shortened URL.

---

## Contribution

1. Fork the repository.
2. Create a new branch:
   ```bash
   git checkout -b feature/new-feature
   ```
3. Commit your changes:
   ```bash
   git commit -m "Add new feature"
   ```
4. Push the branch:
   ```bash
   git push origin feature/new-feature
   ```
5. Open a pull request.

---

## License

This project is licensed under the [MIT License](LICENSE).

---

## Contact

For any inquiries or suggestions, feel free to reach out:

- **Email**: [tsaqiffatih@gmail.com](mailto:tsaqiffatih@gmail.com)
- **GitHub**: [tsaqiffatih](https://github.com/tsaqiffatih)
