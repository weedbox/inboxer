package inboxer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"
)

// MemoryMailStore implements the MailStore interface using memory as the storage medium
type MemoryMailStore struct {
	mu    sync.RWMutex
	mails map[string]*Mail
	idGen IDGenerator
}

// IDGenerator defines the interface for generating unique IDs
type IDGenerator interface {
	GenerateID() string
}

// SimpleIDGenerator is a simple implementation of the ID generator
type SimpleIDGenerator struct {
	counter int
	mu      sync.Mutex
}

// GenerateID generates a simple unique ID
func (g *SimpleIDGenerator) GenerateID() string {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.counter++
	return fmt.Sprintf("mail_%d_%d", time.Now().UnixNano(), g.counter)
}

// NewMemoryMailStore creates a new memory-based mail storage
func NewMemoryMailStore() *MemoryMailStore {
	return &MemoryMailStore{
		mails: make(map[string]*Mail),
		idGen: &SimpleIDGenerator{},
	}
}

// CreateMail creates a new mail and returns the mail ID
func (s *MemoryMailStore) CreateMail(ctx context.Context, mail *Mail) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if mail == nil {
		return "", errors.New("mail cannot be nil")
	}

	// Generate ID and copy the mail object
	if mail.ID == "" {
		mail.ID = s.idGen.GenerateID()
	}

	// Deep copy the mail object to avoid reference issues
	mailCopy := copyMail(mail)
	s.mails[mail.ID] = mailCopy

	return mail.ID, nil
}

// GetMail retrieves a mail by ID
func (s *MemoryMailStore) GetMail(ctx context.Context, mailID string) (*Mail, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	mail, exists := s.mails[mailID]
	if !exists {
		return nil, fmt.Errorf("mail with ID %s not found", mailID)
	}

	return copyMail(mail), nil
}

// UpdateMail updates an existing mail
func (s *MemoryMailStore) UpdateMail(ctx context.Context, mail *Mail) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if mail == nil || mail.ID == "" {
		return errors.New("mail cannot be nil and must have an ID")
	}

	if _, exists := s.mails[mail.ID]; !exists {
		return fmt.Errorf("mail with ID %s not found", mail.ID)
	}

	s.mails[mail.ID] = copyMail(mail)
	return nil
}

// DeleteMail deletes a mail by ID
func (s *MemoryMailStore) DeleteMail(ctx context.Context, mailID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.mails[mailID]; !exists {
		return fmt.Errorf("mail with ID %s not found", mailID)
	}

	delete(s.mails, mailID)
	return nil
}

// CreateBatchMails creates multiple mails in batch
func (s *MemoryMailStore) CreateBatchMails(ctx context.Context, mails []*Mail) ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(mails) == 0 {
		return []string{}, nil
	}

	ids := make([]string, 0, len(mails))
	for _, mail := range mails {
		if mail == nil {
			continue
		}

		if mail.ID == "" {
			mail.ID = s.idGen.GenerateID()
		}

		mailCopy := copyMail(mail)
		s.mails[mail.ID] = mailCopy
		ids = append(ids, mail.ID)
	}

	return ids, nil
}

// DeleteMailsByRecipient deletes all mails for a specific recipient
func (s *MemoryMailStore) DeleteMailsByRecipient(ctx context.Context, recipientID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if recipientID == "" {
		return errors.New("recipientID cannot be empty")
	}

	toDelete := []string{}
	for id, mail := range s.mails {
		if mail.RecipientID == recipientID {
			toDelete = append(toDelete, id)
		}
	}

	for _, id := range toDelete {
		delete(s.mails, id)
	}

	return nil
}

// DeleteExpiredMails deletes all expired mails
func (s *MemoryMailStore) DeleteExpiredMails(ctx context.Context, beforeTime time.Time) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	toDelete := []string{}
	for id, mail := range s.mails {
		if !mail.ExpireTime.IsZero() && mail.ExpireTime.Before(beforeTime) {
			toDelete = append(toDelete, id)
		}
	}

	for _, id := range toDelete {
		delete(s.mails, id)
	}

	return len(toDelete), nil
}

// GetMailsByRecipient retrieves mails for a specific recipient with pagination
func (s *MemoryMailStore) GetMailsByRecipient(ctx context.Context, recipientID string, page, size int) ([]*Mail, int, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	// Collect all matching mails
	matchedMails := []*Mail{}
	for _, mail := range s.mails {
		if mail.RecipientID == recipientID {
			matchedMails = append(matchedMails, copyMail(mail))
		}
	}

	// Sort by creation time (newest first)
	sort.Slice(matchedMails, func(i, j int) bool {
		return matchedMails[i].CreateTime.After(matchedMails[j].CreateTime)
	})

	// Calculate total and pagination
	total := len(matchedMails)
	start := (page - 1) * size
	end := start + size

	if start >= total {
		return []*Mail{}, total, nil
	}
	if end > total {
		end = total
	}

	return matchedMails[start:end], total, nil
}

// QueryMails queries mails by filter conditions with pagination
func (s *MemoryMailStore) QueryMails(ctx context.Context, filter *MailFilter, page, size int) ([]*Mail, int, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	matchedMails := []*Mail{}
	now := time.Now()

	for _, mail := range s.mails {
		if !matchMail(mail, filter, now) {
			continue
		}
		matchedMails = append(matchedMails, copyMail(mail))
	}

	// Sort by creation time (newest first)
	sort.Slice(matchedMails, func(i, j int) bool {
		return matchedMails[i].CreateTime.After(matchedMails[j].CreateTime)
	})

	// Calculate total and pagination
	total := len(matchedMails)
	start := (page - 1) * size
	end := start + size

	if start >= total {
		return []*Mail{}, total, nil
	}
	if end > total {
		end = total
	}

	return matchedMails[start:end], total, nil
}

// CountUnreadMails counts the number of unread mails for a specific recipient
func (s *MemoryMailStore) CountUnreadMails(ctx context.Context, recipientID string) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	count := 0
	for _, mail := range s.mails {
		if mail.RecipientID == recipientID && !mail.ReadStatus {
			count++
		}
	}

	return count, nil
}

// CountMailsWithAttachments counts the number of mails with attachments for a specific recipient
func (s *MemoryMailStore) CountMailsWithAttachments(ctx context.Context, recipientID string) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	count := 0
	for _, mail := range s.mails {
		if mail.RecipientID == recipientID && mail.Attachments != nil && len(mail.Attachments) > 0 {
			count++
		}
	}

	return count, nil
}

// ExportMailLogs exports mail logs based on filter
func (s *MemoryMailStore) ExportMailLogs(ctx context.Context, filter *MailFilter) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	matchedMails := []*Mail{}
	now := time.Now()

	for _, mail := range s.mails {
		if !matchMail(mail, filter, now) {
			continue
		}
		matchedMails = append(matchedMails, copyMail(mail))
	}

	// Sort by creation time (newest first)
	sort.Slice(matchedMails, func(i, j int) bool {
		return matchedMails[i].CreateTime.After(matchedMails[j].CreateTime)
	})

	// Convert mails to JSON format
	data, err := json.MarshalIndent(matchedMails, "", "  ")
	if err != nil {
		return "", fmt.Errorf("error marshaling mails to JSON: %w", err)
	}

	return string(data), nil
}

// Helper function: Deep copy a mail object
func copyMail(mail *Mail) *Mail {
	if mail == nil {
		return nil
	}

	mailCopy := &Mail{
		ID:          mail.ID,
		SenderID:    mail.SenderID,
		RecipientID: mail.RecipientID,
		Title:       mail.Title,
		Content:     mail.Content,
		ReadStatus:  mail.ReadStatus,
		CreateTime:  mail.CreateTime,
		ExpireTime:  mail.ExpireTime,
	}

	// Copy tags
	if mail.Tags != nil {
		mailCopy.Tags = make([]string, len(mail.Tags))
		copy(mailCopy.Tags, mail.Tags)
	}

	// Copy attachments
	if mail.Attachments != nil {
		mailCopy.Attachments = make(map[string]interface{})
		for k, v := range mail.Attachments {
			mailCopy.Attachments[k] = v
		}
	}

	return mailCopy
}

// Helper function: Check if a mail matches the filter conditions
func matchMail(mail *Mail, filter *MailFilter, now time.Time) bool {
	if filter == nil {
		return true
	}

	// Filter by sender
	if filter.SenderID != "" && mail.SenderID != filter.SenderID {
		return false
	}

	// Filter by recipient
	if filter.RecipientID != "" && mail.RecipientID != filter.RecipientID {
		return false
	}

	// Filter by read status
	if filter.ReadStatus != nil && mail.ReadStatus != *filter.ReadStatus {
		return false
	}

	// Filter by creation time range
	if filter.StartTime != nil && mail.CreateTime.Before(*filter.StartTime) {
		return false
	}
	if filter.EndTime != nil && mail.CreateTime.After(*filter.EndTime) {
		return false
	}

	// Filter by expired mails
	if filter.ExpiredOnly && (mail.ExpireTime.IsZero() || !mail.ExpireTime.Before(now)) {
		return false
	}

	// Filter by tags
	if len(filter.Tags) > 0 {
		hasTag := false
		for _, filterTag := range filter.Tags {
			for _, mailTag := range mail.Tags {
				if filterTag == mailTag {
					hasTag = true
					break
				}
			}
			if hasTag {
				break
			}
		}
		if !hasTag {
			return false
		}
	}

	return true
}
