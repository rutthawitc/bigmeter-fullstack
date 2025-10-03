# Go Backend Development Guidelines
## Oracle to PostgreSQL Data Synchronization System

### Table of Contents
1. [Project Overview](#project-overview)
2. [Architecture Guidelines](#architecture-guidelines)
3. [Code Standards](#code-standards)
4. [Database Guidelines](#database-guidelines)
5. [Security & Safety](#security--safety)
6. [Error Handling](#error-handling)
7. [Testing Standards](#testing-standards)
8. [Performance Guidelines](#performance-guidelines)
9. [Monitoring & Logging](#monitoring--logging)
10. [Deployment & Operations](#deployment--operations)

---

## Project Overview

### Purpose
Backend system to synchronize data from Oracle Database to PostgreSQL and provide RESTful API access to the synchronized data.

### Core Requirements
- **Data Sync**: Reliable Oracle → PostgreSQL data synchronization
- **RESTful API**: HTTP API for data access and management
- **High Performance**: Handle concurrent requests efficiently
- **Data Integrity**: Ensure data consistency during sync operations
- **Monitoring**: Comprehensive logging and metrics

### Tech Stack
- **Language**: Go 1.21+
- **Web Framework**: Gin or Echo
- **Oracle Driver**: godror (thick) (`github.com/godror/godror`)
- **PostgreSQL Driver**: pgx or lib/pq
- **Migration Tool**: golang-migrate
- **Testing**: testify, sqlmock
- **Monitoring**: Prometheus + Grafana

---

## Architecture Guidelines

### Project Structure
```
project-root/
├── cmd/
│   ├── api/                 # API server main
│   └── sync/               # Data sync service main
├── internal/
│   ├── api/                # API handlers and routes
│   ├── config/             # Configuration management
│   ├── database/           # Database connections and queries
│   ├── models/             # Data models and structs
│   ├── services/           # Business logic
│   ├── sync/               # Data synchronization logic
│   └── utils/              # Utility functions
├── pkg/                    # Public libraries (if any)
├── migrations/             # Database migrations
├── configs/                # Configuration files
├── docker/                 # Docker configurations
├── scripts/                # Deployment and utility scripts
├── docs/                   # Documentation
└── tests/                  # Integration tests
```

### Service Architecture
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Oracle DB     │───▶│   Sync Service  │───▶│  PostgreSQL DB  │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │
                                ▼
                       ┌─────────────────┐
                       │   API Service   │
                       └─────────────────┘
                                │
                                ▼
                       ┌─────────────────┐
                       │   HTTP Clients  │
                       └─────────────────┘
```

### Design Patterns
- **Repository Pattern**: Database access abstraction
- **Service Layer**: Business logic separation
- **Dependency Injection**: Loose coupling
- **Factory Pattern**: Database connection management
- **Observer Pattern**: Event-driven sync notifications

---

## Code Standards

### General Rules
1. **Package Naming**: Use lowercase, short, descriptive names
2. **Function Naming**: Use camelCase, start with verb
3. **Variable Naming**: Use camelCase, descriptive names
4. **Constants**: Use UPPER_SNAKE_CASE for exported constants
5. **File Naming**: Use snake_case for file names

### Code Organization
```go
// Package declaration and imports
package main

import (
    "context"
    "fmt"
    "log"
    
    // Standard library first
    "database/sql"
    "net/http"
    
    // Third-party packages
    "github.com/gin-gonic/gin"
    _ "github.com/godror/godror" // driver name: "godror" (thick, Instant Client)
    
    // Local packages
    "your-project/internal/config"
    "your-project/internal/database"
)

// Constants
const (
    DefaultTimeout = 30 * time.Second
    MaxRetries     = 3
)

// Types
type UserService struct {
    db     *database.Manager
    logger *log.Logger
}

// Functions
func NewUserService(db *database.Manager, logger *log.Logger) *UserService {
    return &UserService{
        db:     db,
        logger: logger,
    }
}
```

### Interface Guidelines
```go
// Interface names should end with 'er' when possible
type Syncer interface {
    SyncUsers(ctx context.Context) error
    SyncOrders(ctx context.Context) error
}

type UserRepository interface {
    GetByID(ctx context.Context, id int64) (*User, error)
    Create(ctx context.Context, user *User) error
    Update(ctx context.Context, user *User) error
    Delete(ctx context.Context, id int64) error
}

// Keep interfaces small and focused
type Reader interface {
    Read(ctx context.Context, query string) ([]map[string]interface{}, error)
}

type Writer interface {
    Write(ctx context.Context, table string, data []map[string]interface{}) error
}
```

---

## Database Guidelines

### Connection Management
```go
type DBConfig struct {
    Oracle struct {
        Host     string `yaml:"host" validate:"required"`
        Port     int    `yaml:"port" validate:"required"`
        Service  string `yaml:"service" validate:"required"`
        User     string `yaml:"user" validate:"required"`
        Password string `yaml:"password" validate:"required"`
    } `yaml:"oracle"`
    
    PostgreSQL struct {
        Host     string `yaml:"host" validate:"required"`
        Port     int    `yaml:"port" validate:"required"`
        Database string `yaml:"database" validate:"required"`
        User     string `yaml:"user" validate:"required"`
        Password string `yaml:"password" validate:"required"`
        SSLMode  string `yaml:"ssl_mode" default:"require"`
    } `yaml:"postgresql"`
}

type DatabaseManager struct {
    oracle   *sql.DB
    postgres *sql.DB
    config   *DBConfig
}

// Connection Pool Settings (MANDATORY)
func (dm *DatabaseManager) configureConnection(db *sql.DB) {
    db.SetMaxOpenConns(25)           // Maximum open connections
    db.SetMaxIdleConns(10)           // Maximum idle connections  
    db.SetConnMaxLifetime(time.Hour) // Connection lifetime
    db.SetConnMaxIdleTime(30 * time.Minute) // Idle timeout
}
```

### Query Guidelines
```go
// Always use context with timeout
func (ur *UserRepository) GetUser(ctx context.Context, id int64) (*User, error) {
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()
    
    query := `SELECT id, name, email, created_at FROM users WHERE id = $1`
    
    var user User
    err := ur.db.QueryRowContext(ctx, query, id).Scan(
        &user.ID,
        &user.Name,
        &user.Email,
        &user.CreatedAt,
    )
    
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, ErrUserNotFound
        }
        return nil, fmt.Errorf("failed to get user: %w", err)
    }
    
    return &user, nil
}

// Use prepared statements for repeated queries
type UserRepository struct {
    db          *sql.DB
    insertStmt  *sql.Stmt
    updateStmt  *sql.Stmt
}

func NewUserRepository(db *sql.DB) (*UserRepository, error) {
    insertStmt, err := db.Prepare(`
        INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id
    `)
    if err != nil {
        return nil, fmt.Errorf("failed to prepare insert statement: %w", err)
    }
    
    updateStmt, err := db.Prepare(`
        UPDATE users SET name = $1, email = $2, updated_at = NOW() WHERE id = $3
    `)
    if err != nil {
        insertStmt.Close()
        return nil, fmt.Errorf("failed to prepare update statement: %w", err)
    }
    
    return &UserRepository{
        db:         db,
        insertStmt: insertStmt,
        updateStmt: updateStmt,
    }, nil
}
```

### Migration Standards
```sql
-- migrations/001_create_users_table.up.sql
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    oracle_id INTEGER UNIQUE, -- Reference to Oracle record
    sync_status VARCHAR(20) DEFAULT 'pending' CHECK (sync_status IN ('pending', 'synced', 'failed')),
    last_synced_at TIMESTAMP WITH TIME ZONE
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_users_oracle_id ON users(oracle_id);
CREATE INDEX IF NOT EXISTS idx_users_sync_status ON users(sync_status);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

-- migrations/001_create_users_table.down.sql
DROP TABLE IF EXISTS users;
```

---

## Security & Safety

### Input Validation
```go
import "github.com/go-playground/validator/v10"

type CreateUserRequest struct {
    Name  string `json:"name" validate:"required,min=2,max=100"`
    Email string `json:"email" validate:"required,email"`
}

func (h *UserHandler) CreateUser(c *gin.Context) {
    var req CreateUserRequest
    
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "Invalid request format",
            "details": err.Error(),
        })
        return
    }
    
    if err := h.validator.Struct(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "Validation failed",
            "details": err.Error(),
        })
        return
    }
    
    // Process validated request
}
```

### SQL Injection Prevention
```go
// ✅ CORRECT: Always use parameterized queries
func (ur *UserRepository) GetUsersByStatus(ctx context.Context, status string) ([]*User, error) {
    query := `SELECT id, name, email FROM users WHERE status = $1`
    
    rows, err := ur.db.QueryContext(ctx, query, status)
    if err != nil {
        return nil, fmt.Errorf("query failed: %w", err)
    }
    defer rows.Close()
    
    // Process rows...
    return users, nil
}

// ❌ WRONG: Never use string concatenation
func (ur *UserRepository) GetUsersByStatusWrong(ctx context.Context, status string) ([]*User, error) {
    query := fmt.Sprintf("SELECT id, name, email FROM users WHERE status = '%s'", status)
    // This is vulnerable to SQL injection!
}
```

### Sensitive Data Handling
```go
type Config struct {
    Database struct {
        Password string `yaml:"password" json:"-"` // Hide from JSON
    } `yaml:"database"`
}

// Use environment variables for secrets
func LoadConfig() (*Config, error) {
    config := &Config{}
    
    // Load from file
    data, err := os.ReadFile("config.yaml")
    if err != nil {
        return nil, err
    }
    
    if err := yaml.Unmarshal(data, config); err != nil {
        return nil, err
    }
    
    // Override with environment variables
    if password := os.Getenv("DB_PASSWORD"); password != "" {
        config.Database.Password = password
    }
    
    return config, nil
}
```

### Concurrency Safety
```go
type SafeCounter struct {
    mu    sync.RWMutex
    value int64
}

func (sc *SafeCounter) Increment() {
    sc.mu.Lock()
    defer sc.mu.Unlock()
    sc.value++
}

func (sc *SafeCounter) Get() int64 {
    sc.mu.RLock()
    defer sc.mu.RUnlock()
    return sc.value
}

// Use channels for safe communication between goroutines
type SyncJob struct {
    TableName string
    LastSync  time.Time
}

func (s *SyncService) ProcessJobs(ctx context.Context) error {
    jobs := make(chan SyncJob, 100)
    results := make(chan error, 100)
    
    // Start workers
    for i := 0; i < 5; i++ {
        go s.worker(ctx, jobs, results)
    }
    
    // Send jobs
    go func() {
        defer close(jobs)
        for _, job := range s.generateJobs() {
            select {
            case jobs <- job:
            case <-ctx.Done():
                return
            }
        }
    }()
    
    // Collect results
    var errors []error
    for i := 0; i < len(s.generateJobs()); i++ {
        if err := <-results; err != nil {
            errors = append(errors, err)
        }
    }
    
    if len(errors) > 0 {
        return fmt.Errorf("sync errors: %v", errors)
    }
    
    return nil
}
```

---

## Error Handling

### Error Types
```go
// Define custom error types
type ErrType string

const (
    ErrTypeValidation ErrType = "validation"
    ErrTypeNotFound   ErrType = "not_found"
    ErrTypeConflict   ErrType = "conflict"
    ErrTypeInternal   ErrType = "internal"
    ErrTypeExternal   ErrType = "external"
)

type AppError struct {
    Type    ErrType `json:"type"`
    Message string  `json:"message"`
    Code    string  `json:"code,omitempty"`
    Details string  `json:"details,omitempty"`
    Err     error   `json:"-"`
}

func (e *AppError) Error() string {
    if e.Err != nil {
        return fmt.Sprintf("%s: %s (%v)", e.Type, e.Message, e.Err)
    }
    return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

func (e *AppError) Unwrap() error {
    return e.Err
}

// Error constructors
func NewValidationError(message string, err error) *AppError {
    return &AppError{
        Type:    ErrTypeValidation,
        Message: message,
        Err:     err,
    }
}

func NewNotFoundError(resource string) *AppError {
    return &AppError{
        Type:    ErrTypeNotFound,
        Message: fmt.Sprintf("%s not found", resource),
    }
}
```

### Error Handling Middleware
```go
func ErrorHandler() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Next()
        
        if len(c.Errors) == 0 {
            return
        }
        
        err := c.Errors.Last().Err
        
        var appErr *AppError
        if errors.As(err, &appErr) {
            switch appErr.Type {
            case ErrTypeValidation:
                c.JSON(http.StatusBadRequest, appErr)
            case ErrTypeNotFound:
                c.JSON(http.StatusNotFound, appErr)
            case ErrTypeConflict:
                c.JSON(http.StatusConflict, appErr)
            default:
                c.JSON(http.StatusInternalServerError, appErr)
            }
        } else {
            // Unknown error
            c.JSON(http.StatusInternalServerError, gin.H{
                "error": "Internal server error",
            })
        }
    }
}
```

### Retry Logic
```go
func (s *SyncService) withRetry(ctx context.Context, operation func() error) error {
    const maxRetries = 3
    const baseDelay = time.Second
    
    for attempt := 0; attempt <= maxRetries; attempt++ {
        err := operation()
        if err == nil {
            return nil
        }
        
        // Don't retry certain errors
        if isNonRetryableError(err) {
            return err
        }
        
        if attempt == maxRetries {
            return fmt.Errorf("operation failed after %d attempts: %w", maxRetries+1, err)
        }
        
        // Exponential backoff
        delay := baseDelay * time.Duration(1<<uint(attempt))
        s.logger.Printf("Attempt %d failed: %v. Retrying in %v", attempt+1, err, delay)
        
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-time.After(delay):
            continue
        }
    }
    
    return nil
}

func isNonRetryableError(err error) bool {
    // Define errors that shouldn't be retried
    var appErr *AppError
    if errors.As(err, &appErr) {
        return appErr.Type == ErrTypeValidation || appErr.Type == ErrTypeNotFound
    }
    return false
}
```

---

## Testing Standards

### Unit Test Structure
```go
package services

import (
    "context"
    "testing"
    "time"
    
    "github.com/DATA-DOG/go-sqlmock"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestUserService_GetUser(t *testing.T) {
    tests := []struct {
        name        string
        userID      int64
        setupMock   func(sqlmock.Sqlmock)
        expected    *User
        expectedErr string
    }{
        {
            name:   "successful get user",
            userID: 1,
            setupMock: func(mock sqlmock.Sqlmock) {
                rows := sqlmock.NewRows([]string{"id", "name", "email", "created_at"}).
                    AddRow(1, "John Doe", "john@example.com", time.Now())
                
                mock.ExpectQuery("SELECT id, name, email, created_at FROM users WHERE id = \\$1").
                    WithArgs(1).
                    WillReturnRows(rows)
            },
            expected: &User{
                ID:    1,
                Name:  "John Doe",
                Email: "john@example.com",
            },
        },
        {
            name:   "user not found",
            userID: 999,
            setupMock: func(mock sqlmock.Sqlmock) {
                mock.ExpectQuery("SELECT id, name, email, created_at FROM users WHERE id = \\$1").
                    WithArgs(999).
                    WillReturnError(sql.ErrNoRows)
            },
            expectedErr: "user not found",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup
            db, mock, err := sqlmock.New()
            require.NoError(t, err)
            defer db.Close()
            
            tt.setupMock(mock)
            
            repo := NewUserRepository(db)
            service := NewUserService(repo)
            
            // Execute
            ctx := context.Background()
            result, err := service.GetUser(ctx, tt.userID)
            
            // Assert
            if tt.expectedErr != "" {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.expectedErr)
                assert.Nil(t, result)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.expected.ID, result.ID)
                assert.Equal(t, tt.expected.Name, result.Name)
                assert.Equal(t, tt.expected.Email, result.Email)
            }
            
            assert.NoError(t, mock.ExpectationsWereMet())
        })
    }
}
```

### Integration Test Guidelines
```go
func TestIntegration_UserSync(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }
    
    // Setup test databases
    oracleDB := setupTestOracleDB(t)
    defer oracleDB.Close()
    
    postgresDB := setupTestPostgresDB(t)
    defer postgresDB.Close()
    
    // Setup test data in Oracle
    seedOracleData(t, oracleDB)
    
    // Run sync
    syncService := NewSyncService(oracleDB, postgresDB)
    ctx := context.Background()
    
    err := syncService.SyncUsers(ctx)
    require.NoError(t, err)
    
    // Verify data in PostgreSQL
    users := getPostgresUsers(t, postgresDB)
    assert.Len(t, users, 3) // Expected number of synced users
}
```

### Benchmark Tests
```go
func BenchmarkUserService_GetUser(b *testing.B) {
    db, mock, err := sqlmock.New()
    require.NoError(b, err)
    defer db.Close()
    
    rows := sqlmock.NewRows([]string{"id", "name", "email", "created_at"}).
        AddRow(1, "John Doe", "john@example.com", time.Now())
    
    mock.ExpectQuery("SELECT (.+) FROM users WHERE id = \\$1").
        WithArgs(1).
        WillReturnRows(rows).
        Times(b.N)
    
    service := NewUserService(NewUserRepository(db))
    ctx := context.Background()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := service.GetUser(ctx, 1)
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

---

## Performance Guidelines

### Database Optimization
```go
// Use connection pooling
func setupDatabasePool(config *DBConfig) (*sql.DB, error) {
    db, err := sql.Open("postgres", config.PostgreSQL.DSN())
    if err != nil {
        return nil, err
    }
    
    // Production settings
    db.SetMaxOpenConns(25)                 // Maximum connections
    db.SetMaxIdleConns(10)                 // Keep some idle connections
    db.SetConnMaxLifetime(time.Hour)       // Recycle connections every hour
    db.SetConnMaxIdleTime(10 * time.Minute) // Close idle connections
    
    return db, nil
}

// Batch operations for better performance
func (ur *UserRepository) BatchInsert(ctx context.Context, users []*User) error {
    const batchSize = 1000
    
    for i := 0; i < len(users); i += batchSize {
        end := i + batchSize
        if end > len(users) {
            end = len(users)
        }
        
        batch := users[i:end]
        if err := ur.insertBatch(ctx, batch); err != nil {
            return fmt.Errorf("failed to insert batch %d-%d: %w", i, end-1, err)
        }
    }
    
    return nil
}

func (ur *UserRepository) insertBatch(ctx context.Context, users []*User) error {
    // Build bulk insert query
    valueStrings := make([]string, 0, len(users))
    valueArgs := make([]interface{}, 0, len(users)*3)
    
    for i, user := range users {
        valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d)", i*3+1, i*3+2, i*3+3))
        valueArgs = append(valueArgs, user.Name, user.Email, user.OracleID)
    }
    
    query := fmt.Sprintf("INSERT INTO users (name, email, oracle_id) VALUES %s",
        strings.Join(valueStrings, ","))
    
    _, err := ur.db.ExecContext(ctx, query, valueArgs...)
    return err
}
```

### Caching Strategy
```go
import "github.com/patrickmn/go-cache"

type CacheService struct {
    cache *cache.Cache
}

func NewCacheService() *CacheService {
    return &CacheService{
        cache: cache.New(15*time.Minute, 30*time.Minute),
    }
}

func (cs *CacheService) Get(key string) (interface{}, bool) {
    return cs.cache.Get(key)
}

func (cs *CacheService) Set(key string, value interface{}, expiration time.Duration) {
    cs.cache.Set(key, value, expiration)
}

// Service with caching
func (us *UserService) GetUser(ctx context.Context, id int64) (*User, error) {
    cacheKey := fmt.Sprintf("user:%d", id)
    
    // Check cache first
    if cached, found := us.cache.Get(cacheKey); found {
        return cached.(*User), nil
    }
    
    // Get from database
    user, err := us.repo.GetUser(ctx, id)
    if err != nil {
        return nil, err
    }
    
    // Cache the result
    us.cache.Set(cacheKey, user, 5*time.Minute)
    
    return user, nil
}
```

### Rate Limiting
```go
import "golang.org/x/time/rate"

type RateLimiter struct {
    limiter *rate.Limiter
}

func NewRateLimiter(rps int, burst int) *RateLimiter {
    return &RateLimiter{
        limiter: rate.NewLimiter(rate.Limit(rps), burst),
    }
}

func (rl *RateLimiter) Allow() bool {
    return rl.limiter.Allow()
}

// Middleware for rate limiting
func RateLimitMiddleware(limiter *RateLimiter) gin.HandlerFunc {
    return func(c *gin.Context) {
        if !limiter.Allow() {
            c.JSON(http.StatusTooManyRequests, gin.H{
                "error": "Rate limit exceeded",
            })
            c.Abort()
            return
        }
        c.Next()
    }
}
```

---

## Monitoring & Logging

### Structured Logging
```go
import (
    "log/slog"
    "os"
)

func SetupLogger() *slog.Logger {
    opts := &slog.HandlerOptions{
        Level: slog.LevelInfo,
    }
    
    var handler slog.Handler
    if os.Getenv("ENV") == "production" {
        handler = slog.NewJSONHandler(os.Stdout, opts)
    } else {
        handler = slog.NewTextHandler(os.Stdout, opts)
    }
    
    return slog.New(handler)
}

// Usage in services
func (us *UserService) GetUser(ctx context.Context, id int64) (*User, error) {
    us.logger.InfoContext(ctx, "getting user",
        slog.Int64("user_id", id),
        slog.String("operation", "get_user"),
    )
    
    user, err := us.repo.GetUser(ctx, id)
    if err != nil {
        us.logger.ErrorContext(ctx, "failed to get user",
            slog.Int64("user_id", id),
            slog.String("error", err.Error()),
        )
        return nil, err
    }
    
    us.logger.InfoContext(ctx, "user retrieved successfully",
        slog.Int64("user_id", id),
        slog.String("user_name", user.Name),
    )
    
    return user, nil
}
```

### Metrics Collection
```go
import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    syncDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "sync_duration_seconds",
            Help: "Duration of sync operations",
        },
        []string{"table", "status"},
    )
    
    activeConnections = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "database_connections_active",
            Help: "Number of active database connections",
        },
        []string{"database"},
    )
    
    httpRequestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "endpoint", "status"},
    )
)

// Middleware for metrics
func MetricsMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        
        c.Next()
        
        duration := time.Since(start)
        status := fmt.Sprintf("%d", c.Writer.Status())
        
        httpRequestsTotal.WithLabelValues(
            c.Request.Method,
            c.FullPath(),
            status,
        ).Inc()
        
        // Record duration if needed
        if duration > 100*time.Millisecond {
            slog.Warn("slow request",
                "method", c.Request.Method,
                "path", c.Request.URL.Path,
                "duration", duration,
                "status", c.Writer.Status(),
            )
        }
    }
}
```

### Health Checks
```go
type HealthChecker struct {
    oracleDB   *sql.DB
    postgresDB *sql.DB
}

func (hc *HealthChecker) Check(c *gin.Context) {
    ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
    defer cancel()
    
    health := gin.H{
        "status":    "ok",
        "timestamp": time.Now().Unix(),
        "checks":    gin.H{},
    }
    
    // Check Oracle connection
    if err := hc.oracleDB.PingContext(ctx); err != nil {
        health["status"] = "error"
        health["checks"].(gin.H)["oracle"] = gin.H{
            "status": "error",
            "error":  err.Error(),
        }
    } else {
        health["checks"].(gin.H)["oracle"] = gin.H{"status": "ok"}
    }
    
    // Check PostgreSQL connection
    if err := hc.postgresDB.PingContext(ctx); err != nil {
        health["status"] = "error"
        health["checks"].(gin.H)["postgres"] = gin.H{
            "status": "error",
            "error":  err.Error(),
        }
    } else {
        health["checks"].(gin.H)["postgres"] = gin.H{"status": "ok"}
    }
    
    statusCode := http.StatusOK
    if health["status"] == "error" {
        statusCode = http.StatusServiceUnavailable
    }
    
    c.JSON(statusCode, health)
}

---

## Deployment & Operations

### Docker Configuration
```dockerfile
# Multi-stage build for production
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/api

# Production stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

# Copy binary and config
COPY --from=builder /app/main .
COPY --from=builder /app/configs ./configs

# Create non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

EXPOSE 8089

CMD ["./main"]
```

### Docker Compose for Development
```yaml
version: '3.8'

services:
  app:
    build: .
    ports:
      - "8089:8089"
    environment:
      - ENV=development
      - DB_HOST=postgres
      - DB_PASSWORD=password
      - ORACLE_HOST=oracle
    depends_on:
      - postgres
      - oracle
      - redis
    volumes:
      - ./configs:/app/configs

  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: syncdb
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: password
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./migrations:/docker-entrypoint-initdb.d

  oracle:
    image: gvenzl/oracle-xe:21-slim
    environment:
      ORACLE_PASSWORD: password
    ports:
      - "1521:1521"
    volumes:
      - oracle_data:/opt/oracle/oradata

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

volumes:
  postgres_data:
  oracle_data:
```

### Environment Configuration
```go
// config/config.go
package config

import (
    "fmt"
    "os"
    "strconv"
    "time"

    "gopkg.in/yaml.v3"
)

type Config struct {
    Server   ServerConfig   `yaml:"server"`
    Database DatabaseConfig `yaml:"database"`
    Sync     SyncConfig     `yaml:"sync"`
    Redis    RedisConfig    `yaml:"redis"`
    Logging  LoggingConfig  `yaml:"logging"`
}

type ServerConfig struct {
    Port         int           `yaml:"port" default:"8089"`
    ReadTimeout  time.Duration `yaml:"read_timeout" default:"30s"`
    WriteTimeout time.Duration `yaml:"write_timeout" default:"30s"`
    IdleTimeout  time.Duration `yaml:"idle_timeout" default:"120s"`
}

type DatabaseConfig struct {
    Oracle     OracleConfig     `yaml:"oracle"`
    PostgreSQL PostgreSQLConfig `yaml:"postgresql"`
}

type OracleConfig struct {
    Host         string        `yaml:"host" validate:"required"`
    Port         int           `yaml:"port" validate:"required"`
    ServiceName  string        `yaml:"service_name" validate:"required"`
    Username     string        `yaml:"username" validate:"required"`
    Password     string        `yaml:"password" validate:"required"`
    MaxOpenConns int           `yaml:"max_open_conns" default:"25"`
    MaxIdleConns int           `yaml:"max_idle_conns" default:"10"`
    ConnLifetime time.Duration `yaml:"conn_lifetime" default:"1h"`
}

func (oc *OracleConfig) DSN() string {
    return fmt.Sprintf("%s/%s@%s:%d/%s",
        oc.Username, oc.Password, oc.Host, oc.Port, oc.ServiceName)
}

type PostgreSQLConfig struct {
    Host         string        `yaml:"host" validate:"required"`
    Port         int           `yaml:"port" validate:"required"`
    Database     string        `yaml:"database" validate:"required"`
    Username     string        `yaml:"username" validate:"required"`
    Password     string        `yaml:"password" validate:"required"`
    SSLMode      string        `yaml:"ssl_mode" default:"require"`
    MaxOpenConns int           `yaml:"max_open_conns" default:"25"`
    MaxIdleConns int           `yaml:"max_idle_conns" default:"10"`
    ConnLifetime time.Duration `yaml:"conn_lifetime" default:"1h"`
}

func (pc *PostgreSQLConfig) DSN() string {
    return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
        pc.Host, pc.Port, pc.Username, pc.Password, pc.Database, pc.SSLMode)
}

type SyncConfig struct {
    Interval        time.Duration `yaml:"interval" default:"5m"`
    BatchSize       int           `yaml:"batch_size" default:"1000"`
    MaxRetries      int           `yaml:"max_retries" default:"3"`
    RetryDelay      time.Duration `yaml:"retry_delay" default:"30s"`
    EnabledTables   []string      `yaml:"enabled_tables"`
    ConcurrentJobs  int           `yaml:"concurrent_jobs" default:"5"`
}

type RedisConfig struct {
    Host     string `yaml:"host" default:"localhost"`
    Port     int    `yaml:"port" default:"6379"`
    Password string `yaml:"password"`
    DB       int    `yaml:"db" default:"0"`
}

type LoggingConfig struct {
    Level  string `yaml:"level" default:"info"`
    Format string `yaml:"format" default:"json"`
}

func LoadConfig(configPath string) (*Config, error) {
    config := &Config{}

    // Load from YAML file
    if configPath != "" {
        data, err := os.ReadFile(configPath)
        if err != nil {
            return nil, fmt.Errorf("failed to read config file: %w", err)
        }

        if err := yaml.Unmarshal(data, config); err != nil {
            return nil, fmt.Errorf("failed to parse config file: %w", err)
        }
    }

    // Override with environment variables
    overrideWithEnv(config)

    return config, nil
}

func overrideWithEnv(config *Config) {
    // Server
    if port := os.Getenv("PORT"); port != "" {
        if p, err := strconv.Atoi(port); err == nil {
            config.Server.Port = p
        }
    }

    // Oracle
    if host := os.Getenv("ORACLE_HOST"); host != "" {
        config.Database.Oracle.Host = host
    }
    if password := os.Getenv("ORACLE_PASSWORD"); password != "" {
        config.Database.Oracle.Password = password
    }

    // PostgreSQL
    if host := os.Getenv("PG_HOST"); host != "" {
        config.Database.PostgreSQL.Host = host
    }
    if password := os.Getenv("PG_PASSWORD"); password != "" {
        config.Database.PostgreSQL.Password = password
    }

    // Redis
    if host := os.Getenv("REDIS_HOST"); host != "" {
        config.Redis.Host = host
    }
    if password := os.Getenv("REDIS_PASSWORD"); password != "" {
        config.Redis.Password = password
    }
}
```

### Makefile for Development
```makefile
# Makefile
.PHONY: build run test clean docker-up docker-down migrate-up migrate-down

# Variables
APP_NAME=sync-api
DOCKER_COMPOSE=docker-compose
MIGRATE=migrate
DATABASE_URL=postgres://postgres:password@localhost:5432/syncdb?sslmode=disable

# Build the application
build:
	go build -o bin/$(APP_NAME) ./cmd/api

# Run the application
run:
	go run ./cmd/api

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Run integration tests
test-integration:
	go test -v -tags=integration ./tests/integration/...

# Run benchmarks
benchmark:
	go test -bench=. -benchmem ./...

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

# Docker commands
docker-up:
	$(DOCKER_COMPOSE) up -d

docker-down:
	$(DOCKER_COMPOSE) down

docker-logs:
	$(DOCKER_COMPOSE) logs -f

docker-build:
	$(DOCKER_COMPOSE) build

# Database migrations
migrate-up:
	$(MIGRATE) -path migrations -database $(DATABASE_URL) up

migrate-down:
	$(MIGRATE) -path migrations -database $(DATABASE_URL) down

migrate-force:
	$(MIGRATE) -path migrations -database $(DATABASE_URL) force $(VERSION)

migrate-version:
	$(MIGRATE) -path migrations -database $(DATABASE_URL) version

# Linting and formatting
fmt:
	go fmt ./...

lint:
	golangci-lint run

# Security scanning
security:
	gosec ./...

# Dependency management
deps:
	go mod tidy
	go mod vendor

# Generate mocks for testing
generate-mocks:
	go generate ./...

# Performance profiling
profile-cpu:
	go test -cpuprofile=cpu.prof -bench=. ./...
	go tool pprof cpu.prof

profile-mem:
	go test -memprofile=mem.prof -bench=. ./...
	go tool pprof mem.prof

# Development setup
setup: deps migrate-up
	@echo "Development environment setup complete"

# Production build
build-prod:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o bin/$(APP_NAME) ./cmd/api
```

### Kubernetes Deployment
```yaml
# k8s/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: sync-api
  labels:
    app: sync-api
spec:
  replicas: 3
  selector:
    matchLabels:
      app: sync-api
  template:
    metadata:
      labels:
        app: sync-api
    spec:
      containers:
      - name: sync-api
        image: your-registry/sync-api:latest
        ports:
        - containerPort: 8089
        env:
        - name: ENV
          value: "production"
        - name: PG_HOST
          valueFrom:
            secretKeyRef:
              name: database-secret
              key: pg-host
        - name: PG_PASSWORD
          valueFrom:
            secretKeyRef:
              name: database-secret
              key: pg-password
        - name: ORACLE_HOST
          valueFrom:
            secretKeyRef:
              name: database-secret
              key: oracle-host
        - name: ORACLE_PASSWORD
          valueFrom:
            secretKeyRef:
              name: database-secret
              key: oracle-password
        livenessProbe:
          httpGet:
            path: /health
            port: 8089
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8089
          initialDelaySeconds: 5
          periodSeconds: 5
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"

---
apiVersion: v1
kind: Service
metadata:
  name: sync-api-service
spec:
  selector:
    app: sync-api
  ports:
    - protocol: TCP
      port: 80
      targetPort: 8089
  type: LoadBalancer

---
apiVersion: v1
kind: Secret
metadata:
  name: database-secret
type: Opaque
data:
  pg-host: <base64-encoded-host>
  pg-password: <base64-encoded-password>
  oracle-host: <base64-encoded-host>
  oracle-password: <base64-encoded-password>
```

### CI/CD Pipeline (GitHub Actions)
```yaml
# .github/workflows/ci.yml
name: CI/CD Pipeline

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest
    
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_PASSWORD: password
          POSTGRES_DB: testdb
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 5432:5432
    
    steps:
    - uses: actions/checkout@v4
    
    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: 1.21
    
    - name: Cache dependencies
      uses: actions/cache@v3
      with:
        path: ~/go/pkg/mod
        key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
        restore-keys: |
          ${{ runner.os }}-go-
    
    - name: Install dependencies
      run: go mod download
    
    - name: Run tests
      run: go test -v -coverprofile=coverage.out ./...
      env:
        DATABASE_URL: postgres://postgres:password@localhost:5432/testdb?sslmode=disable
    
    - name: Upload coverage
      uses: codecov/codecov-action@v3
      with:
        file: ./coverage.out
    
    - name: Run linter
      uses: golangci/golangci-lint-action@v3
      with:
        version: latest
    
    - name: Security scan
      run: |
        go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
        gosec ./...

  build:
    needs: test
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'
    
    steps:
    - uses: actions/checkout@v4
    
    - name: Set up Docker Buildx
      uses: docker/setup-buildx-action@v3
    
    - name: Login to Container Registry
      uses: docker/login-action@v3
      with:
        registry: your-registry.com
        username: ${{ secrets.REGISTRY_USERNAME }}
        password: ${{ secrets.REGISTRY_PASSWORD }}
    
    - name: Build and push
      uses: docker/build-push-action@v5
      with:
        context: .
        push: true
        tags: your-registry.com/sync-api:${{ github.sha }}
        cache-from: type=gha
        cache-to: type=gha,mode=max

  deploy:
    needs: build
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'
    
    steps:
    - name: Deploy to Kubernetes
      run: |
        # Add your deployment script here
        echo "Deploying to production..."
```

---

## Configuration Management

### Environment-Specific Configs
```yaml
# configs/development.yaml
server:
  port: 8089
  read_timeout: 30s
  write_timeout: 30s

database:
  oracle:
    host: localhost
    port: 1521
    service_name: xe
    username: system
    password: password
    max_open_conns: 10
    max_idle_conns: 5
  postgresql:
    host: localhost
    port: 5432
    database: syncdb_dev
    username: postgres
    password: password
    ssl_mode: disable
    max_open_conns: 10
    max_idle_conns: 5

sync:
  interval: 30s
  batch_size: 100
  max_retries: 3
  concurrent_jobs: 2
  enabled_tables:
    - users
    - orders
    - products

logging:
  level: debug
  format: text

---
# configs/production.yaml
server:
  port: 8089
  read_timeout: 30s
  write_timeout: 30s

database:
  oracle:
    max_open_conns: 25
    max_idle_conns: 10
    conn_lifetime: 1h
  postgresql:
    max_open_conns: 25
    max_idle_conns: 10
    conn_lifetime: 1h
    ssl_mode: require

sync:
  interval: 5m
  batch_size: 1000
  max_retries: 5
  concurrent_jobs: 5
  enabled_tables:
    - users
    - orders
    - products
    - customers
    - inventory

logging:
  level: info
  format: json
```

---

## API Documentation Standards

### OpenAPI Specification
```yaml
# api/openapi.yaml
openapi: 3.0.3
info:
  title: Sync API
  description: Oracle to PostgreSQL Data Synchronization API
  version: 1.0.0
  contact:
    name: API Support
    email: support@example.com

servers:
  - url: http://localhost:8089/api/v1
    description: Development server
  - url: https://api.example.com/v1
    description: Production server

paths:
  /users:
    get:
      summary: List users
      parameters:
        - name: limit
          in: query
          schema:
            type: integer
            minimum: 1
            maximum: 100
            default: 20
        - name: offset
          in: query
          schema:
            type: integer
            minimum: 0
            default: 0
        - name: status
          in: query
          schema:
            type: string
            enum: [active, inactive]
      responses:
        '200':
          description: List of users
          content:
            application/json:
              schema:
                type: object
                properties:
                  data:
                    type: array
                    items:
                      $ref: '#/components/schemas/User'
                  meta:
                    $ref: '#/components/schemas/PaginationMeta'

components:
  schemas:
    User:
      type: object
      properties:
        id:
          type: integer
          format: int64
        name:
          type: string
        email:
          type: string
          format: email
        status:
          type: string
          enum: [active, inactive]
        created_at:
          type: string
          format: date-time
        updated_at:
          type: string
          format: date-time
    
    PaginationMeta:
      type: object
      properties:
        total:
          type: integer
        limit:
          type: integer
        offset:
          type: integer
        has_more:
          type: boolean
```

### API Handler Documentation
```go
// handlers/user.go

// GetUsers godoc
// @Summary      List users
// @Description  Get list of users with pagination
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        limit  query     int  false  "Number of items to return" minimum(1) maximum(100) default(20)
// @Param        offset query     int  false  "Number of items to skip" minimum(0) default(0)
// @Param        status query     string  false  "Filter by status" Enums(active, inactive)
// @Success      200  {object}  response.PaginatedResponse{data=[]models.User}
// @Failure      400  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /users [get]
func (h *UserHandler) GetUsers(c *gin.Context) {
    // Implementation
}
```

---

## Final Checklist

### Pre-Development
- [ ] Set up development environment with Docker
- [ ] Configure database connections (Oracle + PostgreSQL)
- [ ] Set up migration system
- [ ] Create initial project structure
- [ ] Configure logging and monitoring

### Development Phase
- [ ] Implement database models and repositories
- [ ] Create sync service with retry logic
- [ ] Build RESTful API with proper error handling
- [ ] Add input validation and sanitization
- [ ] Implement caching where appropriate
- [ ] Add comprehensive tests (unit + integration)

### Security & Performance
- [ ] Implement rate limiting
- [ ] Add authentication/authorization if needed
- [ ] Configure connection pooling
- [ ] Add batch processing for large datasets
- [ ] Implement proper error handling and logging

### Testing & Quality
- [ ] Unit tests with >80% coverage
- [ ] Integration tests with test databases
- [ ] Load testing for API endpoints
- [ ] Security testing (gosec, dependency scanning)
- [ ] Code review checklist

### Deployment
- [ ] Create Docker images
- [ ] Set up Kubernetes manifests
- [ ] Configure CI/CD pipeline
- [ ] Set up monitoring and alerting
- [ ] Create runbooks and documentation

### Operations
- [ ] Health check endpoints
- [ ] Metrics collection
- [ ] Log aggregation
- [ ] Backup and disaster recovery
- [ ] Performance monitoring

---

## Additional Resources

### Recommended Libraries
- **Web Framework**: `github.com/gin-gonic/gin`
- **Oracle Driver**: `github.com/sijms/go-ora/v2`
- **PostgreSQL Driver**: `github.com/jackc/pgx/v5`
- **Migrations**: `github.com/golang-migrate/migrate`
- **Validation**: `github.com/go-playground/validator/v10`
- **Configuration**: `gopkg.in/yaml.v3`
- **Logging**: `log/slog` (Go 1.21+)
- **Testing**: `github.com/stretchr/testify`
- **Mocking**: `github.com/DATA-DOG/go-sqlmock`
- **Caching**: `github.com/patrickmn/go-cache`
- **Metrics**: `github.com/prometheus/client_golang`

### Development Tools
- **Linting**: `golangci-lint`
- **Security**: `gosec`
- **Documentation**: `swaggo/swag`
- **Profiling**: `go tool pprof`
- **Testing**: `go test -race`

### Best Practices Summary
1. Always use context with timeouts
2. Implement proper error handling with custom error types
3. Use structured logging with appropriate levels
4. Write comprehensive tests with good coverage
5. Follow Go naming conventions and style guides
6. Implement graceful shutdown for services
7. Use dependency injection for better testability
8. Monitor application metrics and performance
9. Follow security best practices for database access
10. Document APIs and maintain up-to-date documentation

This guide provides a comprehensive foundation for building a robust, scalable, and maintainable Go backend system for Oracle to PostgreSQL data synchronization.
```
