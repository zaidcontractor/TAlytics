package services

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/talytics/server/internal/database"
	pb "github.com/talytics/server/proto"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type UserService struct {
	pb.UnimplementedUserServiceServer
	db        *database.Database
	jwtSecret []byte
}

type Claims struct {
	UserID int64  `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func NewUserService(db *database.Database) *UserService {
	// In production, this should be loaded from environment variables
	secret := []byte("your-256-bit-secret")
	return &UserService{
		db:        db,
		jwtSecret: secret,
	}
}

func (s *UserService) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.AuthResponse, error) {
	// Validate input
	if req.Email == "" || req.Name == "" || req.Password == "" || req.Role == "" {
		return nil, errors.New("all fields are required")
	}

	if req.Role != "instructor" && req.Role != "ta" {
		return nil, errors.New("role must be either 'instructor' or 'ta'")
	}

	// Check if user already exists
	var count int
	err := s.db.DB.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", req.Email).Scan(&count)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, errors.New("user with this email already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Insert user
	result, err := s.db.DB.Exec(`
		INSERT INTO users (email, name, password_hash, role, created_at, updated_at)
		VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, req.Email, req.Name, string(hashedPassword), req.Role)
	if err != nil {
		return nil, err
	}

	userID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	// Create user object
	user := &pb.User{
		Id:    userID,
		Email: req.Email,
		Name:  req.Name,
		Role:  req.Role,
		CreatedAt: timestamppb.Now(),
		UpdatedAt: timestamppb.Now(),
	}

	// Generate JWT token
	token, err := s.generateToken(userID, req.Email, req.Role)
	if err != nil {
		return nil, err
	}

	return &pb.AuthResponse{
		User:    user,
		Token:   token,
		Message: "User registered successfully",
	}, nil
}

func (s *UserService) Login(ctx context.Context, req *pb.LoginRequest) (*pb.AuthResponse, error) {
	if req.Email == "" || req.Password == "" {
		return nil, errors.New("email and password are required")
	}

	// Get user from database
	var user pb.User
	var passwordHash string
	var createdAt, updatedAt time.Time

	err := s.db.DB.QueryRow(`
		SELECT id, email, name, password_hash, role, created_at, updated_at
		FROM users WHERE email = ?
	`, req.Email).Scan(&user.Id, &user.Email, &user.Name, &passwordHash, &user.Role, &createdAt, &updatedAt)

	if err == sql.ErrNoRows {
		return nil, errors.New("invalid email or password")
	}
	if err != nil {
		return nil, err
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password))
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	user.CreatedAt = timestamppb.New(createdAt)
	user.UpdatedAt = timestamppb.New(updatedAt)

	// Generate JWT token
	token, err := s.generateToken(user.Id, user.Email, user.Role)
	if err != nil {
		return nil, err
	}

	return &pb.AuthResponse{
		User:    &user,
		Token:   token,
		Message: "Login successful",
	}, nil
}

func (s *UserService) VerifyToken(ctx context.Context, req *pb.VerifyTokenRequest) (*pb.UserResponse, error) {
	claims, err := s.validateToken(req.Token)
	if err != nil {
		return nil, err
	}

	// Get user from database
	var user pb.User
	var createdAt, updatedAt time.Time

	err = s.db.DB.QueryRow(`
		SELECT id, email, name, role, created_at, updated_at
		FROM users WHERE id = ?
	`, claims.UserID).Scan(&user.Id, &user.Email, &user.Name, &user.Role, &createdAt, &updatedAt)

	if err == sql.ErrNoRows {
		return nil, errors.New("user not found")
	}
	if err != nil {
		return nil, err
	}

	user.CreatedAt = timestamppb.New(createdAt)
	user.UpdatedAt = timestamppb.New(updatedAt)

	return &pb.UserResponse{
		User:    &user,
		Message: "Token is valid",
	}, nil
}

func (s *UserService) GetProfile(ctx context.Context, req *pb.GetProfileRequest) (*pb.UserResponse, error) {
	claims, err := s.validateToken(req.Token)
	if err != nil {
		return nil, err
	}

	// Get user from database
	var user pb.User
	var createdAt, updatedAt time.Time

	err = s.db.DB.QueryRow(`
		SELECT id, email, name, role, created_at, updated_at
		FROM users WHERE id = ?
	`, claims.UserID).Scan(&user.Id, &user.Email, &user.Name, &user.Role, &createdAt, &updatedAt)

	if err == sql.ErrNoRows {
		return nil, errors.New("user not found")
	}
	if err != nil {
		return nil, err
	}

	user.CreatedAt = timestamppb.New(createdAt)
	user.UpdatedAt = timestamppb.New(updatedAt)

	return &pb.UserResponse{
		User:    &user,
		Message: "Profile retrieved successfully",
	}, nil
}

func (s *UserService) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	// In a production system, you'd invalidate the token by adding it to a blacklist
	// or removing it from a whitelist stored in the database
	
	// For now, just return success - client will remove token from storage
	return &pb.LogoutResponse{
		Message: "Logged out successfully",
	}, nil
}

func (s *UserService) generateToken(userID int64, email, role string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

func (s *UserService) validateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return s.jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}