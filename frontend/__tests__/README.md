# Frontend Test Suite

Comprehensive test suite for the TAlytics frontend application.

## Test Coverage

### ✅ API Client Tests (`__tests__/lib/api.test.ts`)
- Authentication (register, login, logout)
- Course management
- Assignment operations
- Error handling
- Token management

### ✅ Component Tests
- **ProtectedRoute** (`__tests__/components/ProtectedRoute.test.tsx`)
  - Loading states
  - Authentication checks
  - Role-based access control
  - Redirects

- **Layout** (`__tests__/components/Layout.test.tsx`)
  - Navigation rendering
  - Role-based menu items
  - User display
  - Logout functionality

### ✅ Page Tests
- **Login** (`__tests__/app/login.test.tsx`)
  - Form rendering
  - Validation
  - Error handling
  - Success flow

- **Register** (`__tests__/app/register.test.tsx`)
  - Form rendering
  - Password validation
  - Role selection
  - Error handling

- **Dashboard** (`__tests__/app/dashboard.test.tsx`)
  - Role-specific content
  - Course loading
  - Empty states
  - Error handling

- **Courses** (`__tests__/app/courses.test.tsx`)
  - Course listing
  - Create course modal
  - Error handling
  - Null safety

- **Course Detail** (`__tests__/app/courses-detail.test.tsx`)
  - Assignment listing
  - Create assignment
  - Status badges
  - Error handling

- **Grading** (`__tests__/app/grading.test.tsx`)
  - Submission listing
  - Submission selection
  - Status badges
  - Error handling

### ✅ Auth Context Tests (`__tests__/lib/auth.test.tsx`)
- User state management
- Login/logout
- localStorage integration

## Running Tests

```bash
# Run all tests
npm test

# Run in watch mode
npm run test:watch

# Run with coverage
npm run test:coverage
```

## Test Structure

All tests follow this structure:
1. **Setup**: Mock dependencies, clear state
2. **Arrange**: Set up test data
3. **Act**: Perform actions
4. **Assert**: Verify results

## Key Testing Patterns

### Mocking API Calls
```typescript
jest.mock('@/lib/api')
const mockApi = api as jest.Mocked<typeof api>
mockApi.getCourses.mockResolvedValueOnce({ courses: [] })
```

### Testing Async Components
```typescript
await waitFor(() => {
  expect(screen.getByText('Expected Text')).toBeInTheDocument()
})
```

### Testing User Interactions
```typescript
await userEvent.type(input, 'text')
await userEvent.click(button)
```

## Common Issues & Solutions

### Issue: Component not rendering
**Solution**: Ensure localStorage is set with user data and token

### Issue: API mocks not working
**Solution**: Clear mocks in `beforeEach` and ensure proper typing

### Issue: Async operations timing out
**Solution**: Use `waitFor` with appropriate timeout

## Coverage Goals

- **Statements**: > 80%
- **Branches**: > 75%
- **Functions**: > 80%
- **Lines**: > 80%

## Adding New Tests

When adding new features:
1. Create test file in `__tests__/` directory
2. Follow existing test patterns
3. Mock external dependencies
4. Test happy path and error cases
5. Test edge cases (null, empty arrays, etc.)

## Best Practices

1. **Isolation**: Each test should be independent
2. **Clarity**: Test names should describe what they test
3. **Coverage**: Test both success and failure paths
4. **Mocking**: Mock external dependencies (API, router, etc.)
5. **Cleanup**: Clear state between tests

