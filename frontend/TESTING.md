# Frontend Testing Guide

## Overview

This document describes the comprehensive test suite created to ensure frontend reliability and catch issues early.

## Test Suite Summary

### ✅ Created Test Files (10 test files, 60+ tests)

1. **API Client Tests** - Tests all API interactions
2. **Auth Context Tests** - Tests authentication state management
3. **ProtectedRoute Tests** - Tests route protection and role-based access
4. **Layout Tests** - Tests navigation and UI components
5. **Login Page Tests** - Tests login functionality
6. **Register Page Tests** - Tests registration functionality
7. **Dashboard Tests** - Tests dashboard for different roles
8. **Courses Page Tests** - Tests course management
9. **Course Detail Tests** - Tests assignment management
10. **Grading Page Tests** - Tests TA grading workspace

## What's Tested

### Authentication & Authorization
- ✅ User registration with validation
- ✅ User login with error handling
- ✅ JWT token storage and retrieval
- ✅ Role-based access control
- ✅ Protected route redirects
- ✅ Logout functionality

### API Integration
- ✅ All API endpoints (15+ endpoints)
- ✅ Request/response handling
- ✅ Error handling (401, 400, 500)
- ✅ Token injection in headers
- ✅ FormData uploads

### UI Components
- ✅ Form validation
- ✅ Loading states
- ✅ Error messages
- ✅ Empty states
- ✅ Modal interactions
- ✅ Navigation rendering

### Data Handling
- ✅ Null safety (courses, assignments, submissions)
- ✅ Empty array handling
- ✅ API response parsing
- ✅ State updates after mutations

### User Flows
- ✅ Complete registration flow
- ✅ Complete login flow
- ✅ Course creation flow
- ✅ Assignment creation flow
- ✅ Submission viewing flow

## Running Tests

```bash
# Run all tests once
npm test

# Run tests in watch mode (recommended during development)
npm run test:watch

# Run tests with coverage report
npm run test:coverage
```

## Test Results

Current status: **60 tests** covering:
- ✅ API client (15 tests)
- ✅ Authentication (8 tests)
- ✅ Components (6 tests)
- ✅ Pages (31 tests)

## Key Features of Test Suite

### 1. Comprehensive Coverage
- Tests cover all major user flows
- Tests both success and error paths
- Tests edge cases (null, empty, invalid data)

### 2. Robust Mocking
- API calls are mocked
- Next.js router is mocked
- localStorage is mocked
- All external dependencies isolated

### 3. Real User Interactions
- Uses `@testing-library/user-event` for realistic interactions
- Tests form submissions
- Tests button clicks
- Tests input changes

### 4. Async Handling
- Proper `waitFor` usage for async operations
- Timeout handling
- Loading state verification

### 5. Error Scenarios
- Network errors
- API errors (400, 401, 500)
- Validation errors
- Authentication failures

## Common Issues Caught by Tests

1. **Null Safety**: Tests ensure components handle null/undefined data
2. **Loading States**: Tests verify loading indicators appear
3. **Error Messages**: Tests verify errors are displayed correctly
4. **Role Access**: Tests ensure role-based restrictions work
5. **Form Validation**: Tests verify validation rules
6. **State Updates**: Tests verify UI updates after API calls

## Adding New Tests

When adding new features:

1. **Create test file**: `__tests__/app/your-feature.test.tsx`
2. **Mock dependencies**: Use `jest.mock()` for API, router, etc.
3. **Test happy path**: Test successful operations
4. **Test error cases**: Test failure scenarios
5. **Test edge cases**: Test null, empty, invalid inputs
6. **Test user interactions**: Use `userEvent` for realistic testing

## Example Test Structure

```typescript
describe('YourComponent', () => {
  beforeEach(() => {
    jest.clearAllMocks()
    localStorage.clear()
  })

  it('should render correctly', () => {
    // Test rendering
  })

  it('should handle user interaction', async () => {
    // Test user actions
  })

  it('should handle errors', async () => {
    // Test error scenarios
  })
})
```

## Continuous Integration

These tests should be run:
- Before every commit
- In CI/CD pipeline
- Before deploying to production

## Next Steps

To improve test coverage further:
1. Add integration tests for complete workflows
2. Add E2E tests with Playwright/Cypress
3. Add visual regression tests
4. Add accessibility tests
5. Add performance tests

## Maintenance

- Update tests when API changes
- Update tests when UI changes
- Add tests for new features
- Remove tests for deprecated features
- Keep test descriptions clear and up-to-date

