# AI-Powered Rubric Editor with Regrading Feature

## Overview
This feature allows instructors to use Claude AI to intelligently update rubrics based on grading analytics and TA feedback, with automatic invalidation of existing grades requiring TAs to re-grade submissions.

## Implementation Summary

### 1. Database Changes
**File:** `server/internal/database/database.go`
- Added `needs_regrading` column to `grades` table (INTEGER DEFAULT 0)
- This flag marks grades that need to be re-evaluated after rubric changes

### 2. Backend Services

#### Rubric Service Enhancement
**File:** `server/internal/services/rubric.go`
- Added `MarkGradesForRegrading(rubricID int64)` method
- Finds all assignments using the rubric
- Sets `needs_regrading = 1` for all grades in those assignments
- Called automatically when `force_regrading` flag is set

#### HTTP Endpoints
**File:** `server/cmd/main.go`

**Updated:**
- `PUT /api/rubrics` - Now accepts `force_regrading` boolean parameter
- When true, automatically marks all related grades for regrading

**New:**
- `GET /api/grades/needs-regrading` - Returns all grades by the authenticated TA that need regrading
- Returns: grade ID, assignment info, student info, previous score, update timestamp

**Existing (Enhanced):**
- `POST /api/rubrics/{id}/ai-suggest` - Already existed, now integrated into edit workflow

### 3. Frontend Components

#### New Component: RubricAIEditor
**Files:** 
- `client/src/components/RubricAIEditor.js`
- `client/src/components/RubricAIEditor.css`

**Features:**
- **Step 1: Analysis Phase**
  - Shows current rubric preview
  - "Get AI Suggestions" button triggers Claude analysis
  - Loading state with spinner

- **Step 2: Comparison Mode**
  - Side-by-side diff view (original vs suggested)
  - Each criterion shows:
    - Badge: NEW / MODIFIED / UNCHANGED
    - Original criterion (if exists)
    - AI suggested criterion
    - Point values
    - AI reasoning for the change
  - Click to toggle selection (checkboxes)
  - Real-time weight validation (must sum to 100%)

- **Step 3: Apply Changes**
  - Confirmation dialog warning about regrading requirement
  - Sends `force_regrading: true` to backend
  - Success message with regrading notification

**Visual Design:**
- Merge-request inspired UI
- Color coding: Green (new), Red (modified), Gray (unchanged)
- Sticky footer with selection summary
- Weight validation warnings

#### Updated: InstructorDashboard
**File:** `client/src/components/InstructorDashboard.js`

**Changes:**
- Import `RubricAIEditor` component
- Changed "Edit" button to "🤖 AI Edit Rubric"
- `handleEditAssignment()` now shows AI editor instead of regular form
- New state: `showAIRubricEditor`
- New handlers: `handleAIRubricSave()`, `handleAIRubricCancel()`
- Modal size: `modal-extra-large` for wider AI editor view

#### Updated: TADashboard
**File:** `client/src/components/TADashboard.js`

**New Features:**
1. **Regrading Alert Banner** (on Overview tab)
   - Shows when `regradingNeeded.length > 0`
   - Prominent warning with count
   - "View Submissions for Regrading" button

2. **New Tab: "Regrading"**
   - Icon: 🔄
   - Shows count badge (animated if > 0)
   - Tab highlights with pulse animation when regrading needed

3. **Regrading List View**
   - Info banner explaining why regrading is needed
   - Cards for each submission:
     - Student info
     - Assignment name
     - Previous score
     - Rubric update date
     - "REGRADE NEEDED" badge (pulsing)
     - "Start Regrading" button

4. **Enhanced Stats**
   - Added "Needs Regrading" stat card (orange/red theme)
   - Shows regrading count prominently

**API Integration:**
- Fetches regrading data from `/api/grades/needs-regrading`
- Stores in `regradingNeeded` state
- Updates on component mount

### 4. Styling Enhancements

#### Modal Sizes
**File:** `client/src/components/courses/Modal.css`
- Added `.modal-extra-large` class (1200px max-width, 95% width)

#### TA Dashboard Styles
**File:** `client/src/components/TADashboard.css`

**New Styles:**
- `.regrading-alert` - Red/pink gradient alert banner
- `.stat-card.regrading-stat` - Orange stat card for regrading count
- `.regrading-info-banner` - Yellow info box explaining regrading
- `.grade-card.regrading` - Red-bordered cards for regrading items
- `.tab-button.highlight` - Pulsing animation for regrading tab
- `.tab-count.alert` - Bouncing badge animation
- `@keyframes pulse`, `pulse-badge`, `bounce` - Various animations

## User Workflow

### Instructor Workflow

1. **Navigate to Assignment**
   - Go to course → Assignments tab
   - Click "🤖 AI Edit Rubric" on any assignment

2. **Review Current Rubric**
   - See current criteria and weights
   - Click "✨ Get AI Suggestions"

3. **AI Analysis** (10-20 seconds)
   - Claude analyzes:
     - All grading data for assignments using this rubric
     - TA comments from grade discussions
     - Current rubric structure
   - Returns suggested improvements

4. **Review Suggestions**
   - See side-by-side comparison
   - Each criterion shows:
     - What changed
     - Why it changed (AI reasoning)
     - Point allocation
   - Click criteria to select/deselect
   - Real-time validation (must = 100%)

5. **Apply Changes**
   - Review selection summary
   - See warning: "This will require TAs to re-grade X submissions"
   - Confirm changes
   - Rubric updated
   - All existing grades marked for regrading

### TA Workflow

1. **See Notification**
   - Red alert banner on Overview tab
   - "Regrading" tab badge shows count (animated)
   - Tab pulses to draw attention

2. **Navigate to Regrading Tab**
   - Click "Regrading" tab or "View Submissions" button
   - See info banner explaining rubric change

3. **View Submissions**
   - List of all submissions needing regrading
   - Each shows:
     - Student name
     - Assignment name
     - Previous score (for reference)
     - When rubric was updated

4. **Start Regrading**
   - Click "Start Regrading" button
   - Opens grading viewer with new rubric
   - Grade according to updated criteria
   - Submit grade (clears `needs_regrading` flag)

## Data Flow

```
Instructor edits rubric with AI
    ↓
Frontend: RubricAIEditor sends force_regrading: true
    ↓
Backend: PUT /api/rubrics/{id}
    ↓
RubricService.UpdateRubric()
    ↓
RubricService.MarkGradesForRegrading(rubricID)
    ↓
Database: UPDATE grades SET needs_regrading = 1
    ↓
TA logs in / navigates to dashboard
    ↓
Frontend: GET /api/grades/needs-regrading
    ↓
Backend: Returns grades WHERE needs_regrading = 1 AND grader_id = current_user
    ↓
Frontend: Shows alert + regrading tab
    ↓
TA re-grades submission
    ↓
Frontend: POST /api/grades (updates grade)
    ↓
Backend: Sets needs_regrading = 0 automatically
```

## Claude AI Integration

### Prompt Structure
The AI suggestion endpoint (`/api/rubrics/{id}/ai-suggest`) sends Claude:

**Context:**
- Current rubric (criteria names, weights)
- Grading analytics (score distributions, means, variance)
- ALL TA comments from assignments using this rubric

**Request:**
```
Analyze this rubric based on actual usage data and suggest improvements.
Consider:
1. Are criterion names clear based on TA usage?
2. Should any criteria be split for clarity?
3. Should any criteria be merged?
4. Are point allocations appropriate?
5. What issues do TA comments reveal?

Return JSON with:
- name: criterion name
- max_points: point value
- index: position
- change_reason: explanation of changes
```

**Response Format:**
```json
[
  {
    "name": "Code Quality and Style",
    "max_points": 25,
    "index": 0,
    "change_reason": "Split from original 'Code' criterion for clarity based on TA comments about ambiguity"
  }
]
```

## Security Considerations

1. **Authorization:**
   - Only instructors can trigger AI suggestions
   - Only instructors can update rubrics
   - Only course instructors can modify course rubrics
   - TAs can only see their own regrading assignments

2. **Validation:**
   - Weights must sum to 100% (client + server validation)
   - At least one criterion required
   - Rubric ownership verified before updates

3. **Data Integrity:**
   - Transaction-based updates
   - Cascade rules prevent orphaned data
   - Grade history preserved (old scores visible)

## Performance Considerations

1. **AI Analysis:**
   - Can take 10-20 seconds for Claude API
   - Shows loading state to user
   - Timeout handling needed (not yet implemented)

2. **Regrading Queries:**
   - Indexed on `needs_regrading` flag
   - Indexed on `grader_id`
   - JOIN optimization for submission/assignment names

3. **Bulk Updates:**
   - Uses single UPDATE per assignment
   - Could be optimized with batch updates for very large datasets

## Future Enhancements

1. **Notifications:**
   - Email TAs when regrading is required
   - In-app notification system
   - Push notifications for mobile

2. **Regrading Progress:**
   - Track % of regrading completed
   - Dashboard widget showing progress
   - Deadline system for regrading

3. **Rubric Versioning:**
   - Keep history of all rubric versions
   - Allow comparing different versions
   - Roll back to previous versions

4. **AI Improvements:**
   - Use Claude 3 Opus for better analysis
   - Multi-step prompting for deeper insights
   - Learning from instructor accept/reject patterns

5. **Batch Regrading:**
   - Allow TAs to mark multiple submissions for batch processing
   - Bulk grade adjustments based on rubric changes

## Testing Checklist

- [ ] Instructor can open AI editor
- [ ] AI suggestions are fetched and displayed
- [ ] Side-by-side comparison shows correctly
- [ ] Selection toggles work
- [ ] Weight validation functions
- [ ] Apply changes shows confirmation
- [ ] Grades are marked for regrading
- [ ] TAs see regrading alert
- [ ] Regrading tab shows correct count
- [ ] Regrading list displays correctly
- [ ] Re-grading clears the flag
- [ ] Permissions are enforced
- [ ] Error handling works

## Database Migration Note

If you have existing data, you need to add the `needs_regrading` column:

```sql
ALTER TABLE grades ADD COLUMN needs_regrading INTEGER DEFAULT 0;
```

This is automatically handled on fresh database creation.
