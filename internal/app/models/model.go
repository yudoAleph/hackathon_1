package models

import (
	"time"
)

// User represents a user in the system
type User struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	FullName  string    `gorm:"type:varchar(255);not null;index:idx_users_full_name" json:"full_name" binding:"required"`
	Email     string    `gorm:"type:varchar(255);not null;uniqueIndex:idx_users_email" json:"email" binding:"required,email"`
	Phone     *string   `gorm:"type:varchar(20);index:idx_users_phone" json:"phone,omitempty"` // Optional field
	Password  string    `gorm:"type:varchar(255);not null" json:"-"`                           // Excluded from JSON
	AvatarURL *string   `gorm:"type:varchar(255)" json:"avatar_url,omitempty"`
	CreatedAt time.Time `gorm:"autoCreateTime;index:idx_users_created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	// Relations
	Contacts []Contact `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"contacts,omitempty"`
}

// TableName overrides the table name for User model
func (User) TableName() string {
	return "users"
}

// Contact represents a contact entry for a user
type Contact struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint      `gorm:"not null;index:idx_contacts_user_id,idx_contacts_user_favorite,idx_contacts_user_created" json:"user_id"`
	FullName  string    `gorm:"type:varchar(255);not null;index:idx_contacts_full_name" json:"full_name" binding:"required"`
	Phone     string    `gorm:"type:varchar(20);not null;index:idx_contacts_phone" json:"phone" binding:"required"`
	Email     *string   `gorm:"type:varchar(255);index:idx_contacts_email" json:"email,omitempty"`
	Favorite  bool      `gorm:"default:false;index:idx_contacts_favorite,idx_contacts_user_favorite" json:"favorite"`
	CreatedAt time.Time `gorm:"autoCreateTime;index:idx_contacts_created_at,idx_contacts_user_created" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	// Relations
	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
}

// TableName overrides the table name for Contact model
func (Contact) TableName() string {
	return "contacts"
}

// UserResponse represents the user data sent to clients (without sensitive data)
type UserResponse struct {
	ID        uint      `json:"id"`
	FullName  string    `json:"full_name"`
	Email     string    `json:"email"`
	Phone     *string   `json:"phone,omitempty"` // Optional field
	AvatarURL *string   `json:"avatar_url,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ToResponse converts User to UserResponse
func (u *User) ToResponse() *UserResponse {
	return &UserResponse{
		ID:        u.ID,
		FullName:  u.FullName,
		Email:     u.Email,
		Phone:     u.Phone,
		AvatarURL: u.AvatarURL,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

// ContactResponse represents the contact data sent to clients
type ContactResponse struct {
	ID        uint      `json:"id"`
	UserID    uint      `json:"user_id"`
	FullName  string    `json:"full_name"`
	Phone     string    `json:"phone"`
	Email     *string   `json:"email,omitempty"`
	Favorite  bool      `json:"favorite"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ToResponse converts Contact to ContactResponse
func (c *Contact) ToResponse() *ContactResponse {
	return &ContactResponse{
		ID:        c.ID,
		UserID:    c.UserID,
		FullName:  c.FullName,
		Phone:     c.Phone,
		Email:     c.Email,
		Favorite:  c.Favorite,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}
