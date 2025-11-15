package proto

import (
"google.golang.org/protobuf/types/known/timestamppb"
)

// Submission message
type Submission struct {
Id           int64                  `json:"id"`
AssignmentId int64                  `json:"assignment_id"`
StudentId    string                 `json:"student_id"`
StudentName  string                 `json:"student_name"`
FilePath     string                 `json:"file_path"`
FileName     string                 `json:"file_name"`
UploadedAt   *timestamppb.Timestamp `json:"uploaded_at"`
}

// UploadSubmissionRequest message
type UploadSubmissionRequest struct {
AssignmentId int64  `json:"assignment_id"`
StudentId    string `json:"student_id"`
StudentName  string `json:"student_name"`
FileData     []byte `json:"file_data"`
}

// GetSubmissionRequest message
type GetSubmissionRequest struct {
Id int64 `json:"id"`
}

// ListSubmissionsRequest message
type ListSubmissionsRequest struct {
AssignmentId int64 `json:"assignment_id"`
}

// DeleteSubmissionRequest message
type DeleteSubmissionRequest struct {
Id int64 `json:"id"`
}

// SubmissionResponse message
type SubmissionResponse struct {
Submission *Submission `json:"submission"`
Message    string      `json:"message"`
}

// ListSubmissionsResponse message
type ListSubmissionsResponse struct {
Submissions []*Submission `json:"submissions"`
}

// SubmissionFileResponse message
type SubmissionFileResponse struct {
FileData []byte `json:"file_data"`
FileName string `json:"file_name"`
}

// DeleteSubmissionResponse message
type DeleteSubmissionResponse struct {
Message string `json:"message"`
}

// Unimplemented server
type UnimplementedSubmissionServiceServer struct{}

func (UnimplementedSubmissionServiceServer) mustEmbedUnimplementedSubmissionServiceServer() {}
