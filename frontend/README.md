# TAlytics Frontend

Modern Next.js frontend for the TAlytics AI-powered grading assistant.

## Tech Stack

- **Framework**: Next.js 16 (App Router)
- **Language**: TypeScript
- **Styling**: Tailwind CSS
- **State Management**: React Context API
- **API Client**: Custom fetch-based client

## Features

### ✅ Implemented

1. **Authentication**
   - Login page with JWT token management
   - Registration page with role selection
   - Protected routes with role-based access control
   - Automatic token refresh and logout

2. **Course Management**
   - Course dashboard with listing
   - Create new courses
   - Course detail pages
   - Assignment creation

3. **Grading Workspace**
   - View assigned submissions (for TAs)
   - Submission status tracking
   - Basic submission viewer

4. **Navigation & Layout**
   - Role-based navigation menu
   - Responsive layout component
   - User profile display

5. **Anomaly Dashboard**
   - Placeholder for anomaly reports
   - Protected access (professor/head_ta only)

## Project Structure

```
frontend/
├── app/
│   ├── login/          # Login page
│   ├── register/       # Registration page
│   ├── dashboard/      # Main dashboard
│   ├── courses/        # Course management
│   ├── grading/        # TA grading workspace
│   └── anomalies/      # Anomaly detection dashboard
├── components/
│   ├── Layout.tsx      # Main layout with navigation
│   └── ProtectedRoute.tsx  # Route protection
├── lib/
│   ├── api.ts          # API client
│   └── auth.tsx        # Auth context provider
└── public/             # Static assets
```

## Getting Started

### Prerequisites

- Node.js 18+ 
- Backend server running on `http://localhost:8080`

### Installation

```bash
cd frontend
npm install
```

### Environment Variables

Create `.env.local`:

```bash
NEXT_PUBLIC_API_URL=http://localhost:8080
```

### Development

```bash
npm run dev
```

Open [http://localhost:3000](http://localhost:3000)

### Build

```bash
npm run build
npm start
```

## API Integration

The frontend communicates with the backend API through the `lib/api.ts` client:

- All requests include JWT token from localStorage
- Automatic error handling
- Type-safe API responses
- Support for both JSON and FormData requests

## User Roles

### Professor
- Create and manage courses
- Create assignments
- Upload/create rubrics
- View anomaly reports
- Trigger TA notifications

### Head TA
- Same as professor (course creation optional)

### Grader TA
- View assigned submissions
- Grade submissions
- Access Claude chat (coming soon)

## Next Steps

### To Complete Implementation

1. **Enhanced Grading Workspace**
   - Full submission viewer with syntax highlighting
   - Rubric panel (side drawer/modal)
   - Claude chat interface for rubric questions
   - Real-time grading with Claude recommendations
   - Save progress functionality

2. **Rubric Builder**
   - Canvas-style rubric creation UI
   - Drag-and-drop criteria
   - PDF upload with preview
   - Rubric editing

3. **Anomaly Dashboard**
   - Statistical charts (Chart.js or Recharts)
   - TA severity visualization
   - Outlier grade display
   - Criterion inconsistency reports
   - Regrade risk indicators

4. **Assignment Management**
   - Full assignment detail page
   - Submission upload interface
   - TA distribution visualization
   - Progress tracking

5. **Additional Features**
   - File upload for submissions
   - PDF viewer for rubric PDFs
   - Real-time notifications
   - Export functionality
   - Dark mode support

## Design Principles

- **Minimal & Clean**: Simple, uncluttered interface
- **Visually Guided**: Clear visual hierarchy
- **Responsive**: Works on all screen sizes
- **Accessible**: Follows WCAG guidelines
- **Fast**: Optimized performance with Next.js

## Development Notes

- Uses Next.js App Router (not Pages Router)
- All components are client components ('use client')
- Authentication state managed via React Context
- Protected routes check both authentication and role
- API client handles all backend communication
