# KitaDoc Backend

## Project Overview

This is the backend for the KitaDoc application, a tool for managing kindergarten documentation. It is written in Go and uses a SQLite database. The application provides a RESTful API for managing children, teachers, documentation entries, and assignments. It also includes features for audio recording analysis and document generation.

### Key Technologies

*   **Go:** The primary programming language.
*   **SQLite:** The database used for storing data.
*   **JWT:** Used for user authentication.
*   **Logrus:** Used for structured logging.
*   **Viper:** Used for configuration management.

### Architecture

The application follows a layered architecture:

*   **`handlers`:** Contains the HTTP handlers that receive and respond to API requests.
*   **`services`:** Contains the business logic of the application.
*   **`data`:** Contains the data access layer (DAL) for interacting with the database.
*   **`models`:** Contains the data structures used throughout the application.
*   **`middleware`:** Contains the middleware for handling cross-cutting concerns like authentication, logging, and CORS.

## Building and Running

The project uses a `Makefile` to automate common tasks.

### Build the application

```bash
make build
```

This will create a binary named `kitadoc-backend` in the `bin` directory.

### Run the application

```bash
make run-dev
```

This will start the backend server in development mode. The server will be available at `http://localhost:8070`.

### Run tests

```bash
make test
```

This will run all the tests in the project.

## Development Conventions

*   **Logging:** The application uses `logrus` for structured logging. The log level and format can be configured in the `config/config.yaml` file or through environment variables.
*   **Configuration:** The application uses `viper` for configuration management. Configuration can be provided through a `config.yaml` file, environment variables, or command-line flags.
*   **Database Migrations:** Database migrations are managed using `go-migrate`. Migration files are located in the `migrations` directory.
*   **Code Style:** The project uses `pre-commit` to enforce code style and formatting. Run `make pre-commit` to run the pre-commit hooks.
