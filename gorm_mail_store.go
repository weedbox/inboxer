package inboxer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// GormMailStore implements the MailStore interface using GORM as the storage medium
type GormMailStore struct {
	db *gorm.DB
}

// MailEntity is the database model for Mail objects
type MailEntity struct {
	ID          string `gorm:"primaryKey"`
	SenderID    string `gorm:"index"`
	RecipientID string `gorm:"index"`
	Title       string
	Content     string    `gorm:"type:text"`
	Attachments string    `gorm:"type:text"` // JSON serialized attachments
	ReadStatus  bool      `gorm:"index"`
	CreateTime  time.Time `gorm:"index"`
	ExpireTime  time.Time `gorm:"index"`
	Tags        string    `gorm:"type:text"` // JSON serialized tags
	CreatedAt   time.Time // GORM's default timestamp
	UpdatedAt   time.Time // GORM's default timestamp
}

// TableName specifies the table name for the MailEntity
func (MailEntity) TableName() string {
	return "mails"
}

// NewGormMailStore creates a new GORM-based mail storage
func NewGormMailStore(db *gorm.DB) (*GormMailStore, error) {
	if db == nil {
		return nil, errors.New("database connection cannot be nil")
	}

	// Auto migrate the schema
	err := db.AutoMigrate(&MailEntity{})
	if err != nil {
		return nil, fmt.Errorf("failed to migrate database schema: %w", err)
	}

	return &GormMailStore{
		db: db,
	}, nil
}

// CreateMail creates a new mail and returns the mail ID
func (s *GormMailStore) CreateMail(ctx context.Context, mail *Mail) (string, error) {
	if mail == nil {
		return "", errors.New("mail cannot be nil")
	}

	// If mail has no ID, generate one
	if mail.ID == "" {
		// Use GORM's callback to generate a unique ID, or implement your own ID generation
		mail.ID = fmt.Sprintf("mail_%d", time.Now().UnixNano())
	}

	// Convert mail to entity
	entity, err := mailToEntity(mail)
	if err != nil {
		return "", fmt.Errorf("failed to convert mail to entity: %w", err)
	}

	// Start transaction with context
	tx := s.db.WithContext(ctx)
	result := tx.Create(entity)
	if result.Error != nil {
		return "", fmt.Errorf("failed to create mail: %w", result.Error)
	}

	return mail.ID, nil
}

// GetMail retrieves a mail by ID
func (s *GormMailStore) GetMail(ctx context.Context, mailID string) (*Mail, error) {
	if mailID == "" {
		return nil, errors.New("mail ID cannot be empty")
	}

	var entity MailEntity
	result := s.db.WithContext(ctx).First(&entity, "id = ?", mailID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("mail with ID %s not found", mailID)
		}
		return nil, fmt.Errorf("failed to get mail: %w", result.Error)
	}

	// Convert entity to mail
	mail, err := entityToMail(&entity)
	if err != nil {
		return nil, fmt.Errorf("failed to convert entity to mail: %w", err)
	}

	return mail, nil
}

// UpdateMail updates an existing mail
func (s *GormMailStore) UpdateMail(ctx context.Context, mail *Mail) error {
	if mail == nil || mail.ID == "" {
		return errors.New("mail cannot be nil and must have an ID")
	}

	// Check if mail exists
	var count int64
	result := s.db.WithContext(ctx).Model(&MailEntity{}).Where("id = ?", mail.ID).Count(&count)
	if result.Error != nil {
		return fmt.Errorf("failed to check mail existence: %w", result.Error)
	}
	if count == 0 {
		return fmt.Errorf("mail with ID %s not found", mail.ID)
	}

	// Convert mail to entity
	entity, err := mailToEntity(mail)
	if err != nil {
		return fmt.Errorf("failed to convert mail to entity: %w", err)
	}

	// Update mail
	result = s.db.WithContext(ctx).Save(entity)
	if result.Error != nil {
		return fmt.Errorf("failed to update mail: %w", result.Error)
	}

	return nil
}

// DeleteMail deletes a mail by ID
func (s *GormMailStore) DeleteMail(ctx context.Context, mailID string) error {
	if mailID == "" {
		return errors.New("mail ID cannot be empty")
	}

	result := s.db.WithContext(ctx).Delete(&MailEntity{}, "id = ?", mailID)
	if result.Error != nil {
		return fmt.Errorf("failed to delete mail: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("mail with ID %s not found", mailID)
	}

	return nil
}

// CreateBatchMails creates multiple mails in batch
func (s *GormMailStore) CreateBatchMails(ctx context.Context, mails []*Mail) ([]string, error) {
	if len(mails) == 0 {
		return []string{}, nil
	}

	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	ids := make([]string, 0, len(mails))
	entities := make([]MailEntity, 0, len(mails))

	for _, mail := range mails {
		if mail == nil {
			continue
		}

		// If mail has no ID, generate one
		if mail.ID == "" {
			mail.ID = fmt.Sprintf("mail_%d_%d", time.Now().UnixNano(), len(ids))
		}

		entity, err := mailToEntity(mail)
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to convert mail to entity: %w", err)
		}

		entities = append(entities, *entity)
		ids = append(ids, mail.ID)
	}

	// Create all mails in a batch
	if len(entities) > 0 {
		result := tx.Create(&entities)
		if result.Error != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to create batch mails: %w", result.Error)
		}
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return ids, nil
}

// DeleteMailsByRecipient deletes all mails for a specific recipient
func (s *GormMailStore) DeleteMailsByRecipient(ctx context.Context, recipientID string) error {
	if recipientID == "" {
		return errors.New("recipientID cannot be empty")
	}

	result := s.db.WithContext(ctx).Delete(&MailEntity{}, "recipient_id = ?", recipientID)
	if result.Error != nil {
		return fmt.Errorf("failed to delete mails by recipient: %w", result.Error)
	}

	return nil
}

// DeleteExpiredMails deletes all expired mails
func (s *GormMailStore) DeleteExpiredMails(ctx context.Context, beforeTime time.Time) (int, error) {
	result := s.db.WithContext(ctx).Delete(&MailEntity{}, "expire_time != ? AND expire_time < ?", time.Time{}, beforeTime)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to delete expired mails: %w", result.Error)
	}

	return int(result.RowsAffected), nil
}

// GetMailsByRecipient retrieves mails for a specific recipient with pagination
func (s *GormMailStore) GetMailsByRecipient(ctx context.Context, recipientID string, page, size int) ([]*Mail, int, error) {
	if recipientID == "" {
		return nil, 0, errors.New("recipientID cannot be empty")
	}

	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}

	// Query for total count
	var total int64
	result := s.db.WithContext(ctx).Model(&MailEntity{}).Where("recipient_id = ?", recipientID).Count(&total)
	if result.Error != nil {
		return nil, 0, fmt.Errorf("failed to count mails by recipient: %w", result.Error)
	}

	// No records found
	if total == 0 {
		return []*Mail{}, 0, nil
	}

	// Calculate offset
	offset := (page - 1) * size

	// Query for mail entities with pagination
	var entities []MailEntity
	result = s.db.WithContext(ctx).
		Where("recipient_id = ?", recipientID).
		Order("create_time DESC").
		Offset(offset).
		Limit(size).
		Find(&entities)

	if result.Error != nil {
		return nil, 0, fmt.Errorf("failed to get mails by recipient: %w", result.Error)
	}

	// Convert entities to mails
	mails := make([]*Mail, 0, len(entities))
	for _, entity := range entities {
		mail, err := entityToMail(&entity)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to convert entity to mail: %w", err)
		}
		mails = append(mails, mail)
	}

	return mails, int(total), nil
}

// QueryMails queries mails by filter conditions with pagination
func (s *GormMailStore) QueryMails(ctx context.Context, filter *MailFilter, page, size int) ([]*Mail, int, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}

	tx := s.db.WithContext(ctx).Model(&MailEntity{})

	// Apply filters
	if filter != nil {
		if filter.SenderID != "" {
			tx = tx.Where("sender_id = ?", filter.SenderID)
		}
		if filter.RecipientID != "" {
			tx = tx.Where("recipient_id = ?", filter.RecipientID)
		}
		if filter.ReadStatus != nil {
			tx = tx.Where("read_status = ?", *filter.ReadStatus)
		}
		if filter.StartTime != nil {
			tx = tx.Where("create_time >= ?", *filter.StartTime)
		}
		if filter.EndTime != nil {
			tx = tx.Where("create_time <= ?", *filter.EndTime)
		}
		if filter.ExpiredOnly {
			now := time.Now()
			tx = tx.Where("expire_time != ? AND expire_time < ?", time.Time{}, now)
		}
		if len(filter.Tags) > 0 {
			// This is a simplistic approach - in a real database you might use a more optimized
			// query for tag filtering, especially for databases that support JSON operations
			for _, tag := range filter.Tags {
				tx = tx.Where("tags LIKE ?", "%"+tag+"%")
			}
		}
	}

	// Count total matching records
	var total int64
	result := tx.Count(&total)
	if result.Error != nil {
		return nil, 0, fmt.Errorf("failed to count filtered mails: %w", result.Error)
	}

	// No records found
	if total == 0 {
		return []*Mail{}, 0, nil
	}

	// Calculate offset
	offset := (page - 1) * size

	// Query for mail entities with pagination
	var entities []MailEntity
	result = tx.Order("create_time DESC").Offset(offset).Limit(size).Find(&entities)
	if result.Error != nil {
		return nil, 0, fmt.Errorf("failed to query mails: %w", result.Error)
	}

	// Convert entities to mails
	mails := make([]*Mail, 0, len(entities))
	for _, entity := range entities {
		mail, err := entityToMail(&entity)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to convert entity to mail: %w", err)
		}
		mails = append(mails, mail)
	}

	return mails, int(total), nil
}

// CountUnreadMails counts the number of unread mails for a specific recipient
func (s *GormMailStore) CountUnreadMails(ctx context.Context, recipientID string) (int, error) {
	if recipientID == "" {
		return 0, errors.New("recipientID cannot be empty")
	}

	var count int64
	result := s.db.WithContext(ctx).Model(&MailEntity{}).Where("recipient_id = ? AND read_status = ?", recipientID, false).Count(&count)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to count unread mails: %w", result.Error)
	}

	return int(count), nil
}

// CountMailsWithAttachments counts the number of mails with attachments for a specific recipient
func (s *GormMailStore) CountMailsWithAttachments(ctx context.Context, recipientID string) (int, error) {
	if recipientID == "" {
		return 0, errors.New("recipientID cannot be empty")
	}

	var count int64
	result := s.db.WithContext(ctx).Model(&MailEntity{}).
		Where("recipient_id = ? AND attachments != ? AND attachments != '[]' AND attachments != '{}'", recipientID, "").
		Count(&count)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to count mails with attachments: %w", result.Error)
	}

	return int(count), nil
}

// ExportMailLogs exports mail logs based on filter
func (s *GormMailStore) ExportMailLogs(ctx context.Context, filter *MailFilter) (string, error) {
	// Reuse the QueryMails function to get filtered mails
	// Set a high limit to get all matching mails
	mails, _, err := s.QueryMails(ctx, filter, 1, 10000)
	if err != nil {
		return "", fmt.Errorf("failed to query mails for export: %w", err)
	}

	// Convert mails to JSON
	data, err := json.MarshalIndent(mails, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal mails to JSON: %w", err)
	}

	return string(data), nil
}

// Helper function: Convert Mail to MailEntity
func mailToEntity(mail *Mail) (*MailEntity, error) {
	entity := &MailEntity{
		ID:          mail.ID,
		SenderID:    mail.SenderID,
		RecipientID: mail.RecipientID,
		Title:       mail.Title,
		Content:     mail.Content,
		ReadStatus:  mail.ReadStatus,
		CreateTime:  mail.CreateTime,
		ExpireTime:  mail.ExpireTime,
	}

	// Serialize attachments to JSON
	if mail.Attachments != nil {
		attachmentsJSON, err := json.Marshal(mail.Attachments)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal attachments: %w", err)
		}
		entity.Attachments = string(attachmentsJSON)
	} else {
		entity.Attachments = "{}"
	}

	// Serialize tags to JSON
	if mail.Tags != nil {
		tagsJSON, err := json.Marshal(mail.Tags)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal tags: %w", err)
		}
		entity.Tags = string(tagsJSON)
	} else {
		entity.Tags = "[]"
	}

	return entity, nil
}

// Helper function: Convert MailEntity to Mail
func entityToMail(entity *MailEntity) (*Mail, error) {
	mail := &Mail{
		ID:          entity.ID,
		SenderID:    entity.SenderID,
		RecipientID: entity.RecipientID,
		Title:       entity.Title,
		Content:     entity.Content,
		ReadStatus:  entity.ReadStatus,
		CreateTime:  entity.CreateTime,
		ExpireTime:  entity.ExpireTime,
	}

	// Deserialize attachments from JSON
	if entity.Attachments != "" {
		var attachments map[string]interface{}
		if err := json.Unmarshal([]byte(entity.Attachments), &attachments); err != nil {
			return nil, fmt.Errorf("failed to unmarshal attachments: %w", err)
		}
		mail.Attachments = attachments
	}

	// Deserialize tags from JSON
	if entity.Tags != "" {
		var tags []string
		if err := json.Unmarshal([]byte(entity.Tags), &tags); err != nil {
			return nil, fmt.Errorf("failed to unmarshal tags: %w", err)
		}
		mail.Tags = tags
	}

	return mail, nil
}
