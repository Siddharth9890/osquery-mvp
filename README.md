# Osquery MVP

A simple service that collects system data using osquery, stores it in a database, and exposes it through a web UI and API.

## Features

- Collects OS information and installed applications using osquery
- Stores data in a MySQL database (via Docker)
- Provides a clean web dashboard to visualize system information
- Exposes API endpoints for data retrieval
- Includes structured logging and request tracing
- Configurable via environment variables

## Prerequisites

- Go 1.18 or later
- Docker and Docker Compose
- osquery installed on your system

## Installation

### 1. Install osquery

#### macOS

```bash
brew update
brew install osquery
```

#### Other systems

Follow the installation instructions at [osquery.io/downloads](https://osquery.io/downloads/).

### 2. Clone the repository

```bash
git clone https://github.com/Siddharth9890/osquery-mvp.git
cd osquery-mvp
```

### 3. Configure environment variables

```bash
cp .env.example .env
```

Edit the `.env` file to configure database credentials and other settings if needed.

### 4. Start the database

```bash
make db-up
```

### 5. Build and run the application

```bash
make run
```

## Usage

### Web Dashboard

Access the web dashboard at:

```
http://localhost:8080
```

### API Endpoints

Get the latest system data:

```
http://localhost:8080/api/latest_data
```

## Logging

The application uses structured JSON logging with the following log levels:

- `debug`: Detailed information for debugging
- `info`: General information about application operation
- `warn`: Warning events that might need attention
- `error`: Error conditions that don't stop the application

Set the log level in the `.env` file:

```
LOG_LEVEL=info
```

## Data Collection

By default, data is collected:

- At application startup
- Every 15 minutes thereafter (configurable via REFRESH_INTERVAL in .env)

## Troubleshooting

- **Database Connection Issues**: Ensure Docker is running and the database container is healthy with `docker ps`
- **osquery Not Found**: Verify osquery is correctly installed by running `osqueryi --version` in your terminal
- **Application Won't Start**: Check the logs for detailed error messages

## Cleanup

To stop the database and clean up:

```bash
make db-down    # Stop the database
# or
make db-clean   # Stop the database and remove volumes
```
