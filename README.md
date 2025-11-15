# TAlytics

Submission for the Operations and Staff Support Track, Claude for Good 2025.  
Team Members: Zaid Contractor and Philip Naveen.

---

## Overview

TAlytics is a platform for data-driven grading and analytics, integrating artificial intelligence and modern web technologies. The project provides tools and automation for educational staff to efficiently design rubrics, manage grading, and analyze student assessment data.

**Tech Stack**

- Java Script + React
- GoLang + gRPC + Protobuf
- Anthropic Claude API
- Local RDBMS (SQLite)

## Project Structure

- [`GradingSuperSystem/`](https://github.com/zaidcontractor/TAlytics/tree/main/GradingSuperSystem):  
  The main folder housing all application source code and supporting files.
  - `.github/`, `.vscode/`: Configuration files for development and contribution.
  - [`AI_RUBRIC_EDITOR_FEATURE.md`](https://github.com/zaidcontractor/TAlytics/blob/main/GradingSuperSystem/AI_RUBRIC_EDITOR_FEATURE.md): Technical documentation on the AI-powered rubric editor.
  - [`README.md`](https://github.com/zaidcontractor/TAlytics/blob/main/GradingSuperSystem/README.md): In-depth explanation of features and usage.
  - `node_modules/`, [`package.json`](https://github.com/zaidcontractor/TAlytics/blob/main/GradingSuperSystem/package.json), [`package-lock.json`](https://github.com/zaidcontractor/TAlytics/blob/main/GradingSuperSystem/package-lock.json): Dependencies for the JavaScript front-end.

### Client Application
Located in [`GradingSuperSystem/client/`](https://github.com/zaidcontractor/TAlytics/tree/main/GradingSuperSystem/client):
- Implements the web-based user interface using JavaScript, React, and CSS.
- [`src/`](https://github.com/zaidcontractor/TAlytics/tree/main/GradingSuperSystem/client/src): Source code for React components, pages, and client logic.
- [`public/`](https://github.com/zaidcontractor/TAlytics/tree/main/GradingSuperSystem/client/public): Static assets (HTML, icons, etc).
- [`README.md`](https://github.com/zaidcontractor/TAlytics/blob/main/GradingSuperSystem/client/README.md): Front-end usage documentation.

### Server Application
Located in [`GradingSuperSystem/server/`](https://github.com/zaidcontractor/TAlytics/tree/main/GradingSuperSystem/server):
- Written in Go, responsible for API services, grading logic, and database management.
- [`cmd/`](https://github.com/zaidcontractor/TAlytics/tree/main/GradingSuperSystem/server/cmd): Entrypoints for running server binaries.
- [`internal/`](https://github.com/zaidcontractor/TAlytics/tree/main/GradingSuperSystem/server/internal): Core backend logic.
- [`proto/`](https://github.com/zaidcontractor/TAlytics/tree/main/GradingSuperSystem/server/proto): Protocol Buffers for gRPC communication.
- Files such as [`talytics.pb.go`](https://github.com/zaidcontractor/TAlytics/blob/main/GradingSuperSystem/server/talytics.pb.go) and [`talytics_grpc.pb.go`](https://github.com/zaidcontractor/TAlytics/blob/main/GradingSuperSystem/server/talytics_grpc.pb.go): Auto-generated gRPC code.
- [`talytics.db`](https://github.com/zaidcontractor/TAlytics/blob/main/GradingSuperSystem/server/talytics.db): Local SQLite database for persistence.

## Key Features

- **AI-Powered Rubric Editor** ([see documentation](https://github.com/zaidcontractor/TAlytics/blob/main/GradingSuperSystem/AI_RUBRIC_EDITOR_FEATURE.md)):  
  Automate rubric creation and standardization using generative AI.
- **Modern Web Interface**:  
  React-driven, responsive dashboard for educators and staff.
- **Backend API & gRPC Services**:  
  Fast Go server handles complex grading, reporting, and analytics.
- **Data Storage**:  
  All assessment data managed with SQLite.
- **Extensible Protocols**:  
  Use of Protocol Buffers for scalable API endpoint definition.

## Getting Started

1. **Clone the Repository:**
   ```bash
   git clone https://github.com/zaidcontractor/TAlytics.git
   cd TAlytics/GradingSuperSystem
   ```

2. **Set up the client (React app):**
   ```bash
   cd client
   npm install
   npm start
   ```

3. **Set up the server (Go backend):**
   ```bash
   cd ../server
   go run main
   ```
   Or compile and run the `talytics-server` binary.

## Documentation

- **[GradingSuperSystem/README.md](https://github.com/zaidcontractor/TAlytics/blob/main/GradingSuperSystem/README.md)**: Full technical documentation.
- **[AI_RUBRIC_EDITOR_FEATURE.md](https://github.com/zaidcontractor/TAlytics/blob/main/GradingSuperSystem/AI_RUBRIC_EDITOR_FEATURE.md)**: Details on the rubric editor's AI integration.

## Contributing

All contributions are welcome!  
Please refer to the `.github/` and individual README files for contribution standards.

---

*This project is part of the Claude for Good 2025 hackathon.*
