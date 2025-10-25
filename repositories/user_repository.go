package repositories

import (
	"sync"
	"time"

	"messaging-app/domain"

	"github.com/google/uuid"
)

// MemoryUserRepository implements UserRepository with in-memory storage
type MemoryUserRepository struct {
	users map[string]*domain.User
	mutex sync.RWMutex
}

// NewMemoryUserRepository creates a new in-memory user repository (sorry for the out of creativity on naming)
func NewMemoryUserRepository() *MemoryUserRepository {
	return &MemoryUserRepository{
		users: make(map[string]*domain.User),
	}
}

// Create adds a new user to the repository
func (r *MemoryUserRepository) Create(user *domain.User) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Generate UUID if not provided
	if user.ID == "" {
		user.ID = uuid.New().String()
	}

	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now()
	}

	r.users[user.ID] = user
	return nil
}

// FindByID retrieves a user by their ID
func (r *MemoryUserRepository) FindByID(id string) (*domain.User, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	user, exists := r.users[id]
	if !exists {
		return nil, domain.ErrUserNotFound
	}

	return user, nil
}

// FindByUsername retrieves a user by their username
func (r *MemoryUserRepository) FindByUsername(username string) (*domain.User, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	for _, user := range r.users {
		if user.Username == username {
			return user, nil
		}
	}

	return nil, domain.ErrUserNotFound
}

// Exists checks if a username already exists
func (r *MemoryUserRepository) UsernameExists(username string) bool {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	for _, user := range r.users {
		if user.Username == username {
			return true
		}
	}

	return false
}
