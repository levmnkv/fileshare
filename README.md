# FileShare MVP

A lightweight, blazing fast, and beautifully designed web service for sharing files and text across a local network. Built specifically with Raspberry Pi 4 (ARM64) in mind, this project uses a Go backend and a dependency-free Vanilla HTML/CSS/JS frontend.

## Features

- **Text Sharing (Notepad)**: A synchronized text area for quick copy-pasting of snippets, URLs, or notes across the network.
- **File Storage**: Drag-and-drop file upload, file listing, downloading, and deletion.
- **Responsive Premium Design**: Stunning glassmorphism UI with dark mode, fluid animations, and Toast notifications.
- **Lightweight & Fast**: Zero external dependencies on the frontend. Compiled to a single statically linked binary on the backend.
- **Docker Ready**: Multi-stage `Dockerfile` optimized for ARM64 and Alpine Linux.

## Tech Stack

- **Backend**: Go (Golang)
- **Frontend**: HTML5, Vanilla JavaScript, CSS3
- **Containerization**: Docker

## Installation & Usage (Docker)

The easiest way to run the service is via Docker. By default, the application is configured to store files in `/mnt/docker-data` on your host machine.

### Prerequisites
- Docker installed on your target machine (e.g., Raspberry Pi).

### Quick Start

1. Clone the repository:
   ```bash
   git clone https://github.com/levmnkv/fileshare.git
   cd fileshare
   ```

2. Build the Docker image:
   ```bash
   docker build -t fileshare-app .
   ```

3. Run the container (make sure the volume path exists or let Docker create it):
   ```bash
   docker run -d \
     -p 8080:8080 \
     -v /mnt/docker-data:/mnt/docker-data \
     --name fileshare fileshare-app
   ```

4. Access the application in your browser:
   `http://<YOUR_DEVICE_IP>:8080`

## Configuration

You can customize the upload directory using the `UPLOAD_DIR` environment variable. By default, it falls back to `/mnt/docker-data`.

If running locally without Docker:
```bash
export UPLOAD_DIR="./uploads"
go run main.go
```

## Development

The project follows Test-Driven Development (TDD) principles. You can run the test suite using standard Go tooling:

```bash
go test -v ./...
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
