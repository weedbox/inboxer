package inboxer

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// DefaultMailManager implements the MailManager interface
type DefaultMailManager struct {
	store       MailStore    // Storage backend
	cleanupTick *time.Ticker // Ticker for periodic cleanup
	cleanupStop chan bool    // Channel to stop cleanup goroutine
	mu          sync.Mutex   // Mutex for managing concurrent operations
}

// NewDefaultMailManager creates a new mail manager with the provided store
func NewDefaultMailManager(store MailStore) *DefaultMailManager {
	return &DefaultMailManager{
		store:       store,
		cleanupStop: make(chan bool),
	}
}

// SendMail sends a single mail
func (m *DefaultMailManager) SendMail(ctx context.Context, mail *Mail) (string, error) {
	if mail == nil {
		return "", errors.New("mail cannot be nil")
	}

	// Set default values if not provided
	m.prepareMailForSending(mail)

	// Store the mail
	return m.store.CreateMail(ctx, mail)
}

// SendBatchMail sends the same mail content to multiple recipients
func (m *DefaultMailManager) SendBatchMail(ctx context.Context, mail *Mail, recipientIDs []string) ([]string, error) {
	if mail == nil {
		return nil, errors.New("mail cannot be nil")
	}
	if len(recipientIDs) == 0 {
		return []string{}, nil
	}

	// Set default values for the template mail
	m.prepareMailForSending(mail)

	// Create a mail for each recipient
	mails := make([]*Mail, 0, len(recipientIDs))
	for _, recipientID := range recipientIDs {
		if recipientID == "" {
			continue
		}

		recipientMail := &Mail{
			SenderID:    mail.SenderID,
			RecipientID: recipientID,
			Title:       mail.Title,
			Content:     mail.Content,
			ReadStatus:  false,
			CreateTime:  mail.CreateTime,
			ExpireTime:  mail.ExpireTime,
			Tags:        make([]string, len(mail.Tags)),
		}

		// Copy tags
		copy(recipientMail.Tags, mail.Tags)

		// Copy attachments if any
		if mail.Attachments != nil {
			recipientMail.Attachments = make(map[string]interface{})
			for k, v := range mail.Attachments {
				recipientMail.Attachments[k] = v
			}
		}

		mails = append(mails, recipientMail)
	}

	// Store all mails in batch
	return m.store.CreateBatchMails(ctx, mails)
}

// SendSystemAnnouncement sends a system announcement to all players
// Note: In a real implementation, this would fetch all active player IDs from a player management system
// For this implementation, we're using a placeholder that simply tags the mail appropriately
func (m *DefaultMailManager) SendSystemAnnouncement(ctx context.Context, mail *Mail) (string, error) {
	if mail == nil {
		return "", errors.New("mail cannot be nil")
	}

	// Set default values
	m.prepareMailForSending(mail)

	// Mark as system announcement
	mail.SenderID = "system"
	mail.RecipientID = "all_players" // Special recipient ID for system announcements

	// Add system announcement tag if not already present
	hasAnnouncementTag := false
	for _, tag := range mail.Tags {
		if tag == "system_announcement" {
			hasAnnouncementTag = true
			break
		}
	}
	if !hasAnnouncementTag {
		mail.Tags = append(mail.Tags, "system_announcement")
	}

	// Store the announcement
	return m.store.CreateMail(ctx, mail)
}

// GetMailByID gets a mail by ID
func (m *DefaultMailManager) GetMailByID(ctx context.Context, mailID string) (*Mail, error) {
	if mailID == "" {
		return nil, errors.New("mail ID cannot be empty")
	}

	return m.store.GetMail(ctx, mailID)
}

// GetMailsByRecipient gets a user's mails with pagination
func (m *DefaultMailManager) GetMailsByRecipient(ctx context.Context, recipientID string, page, size int) ([]*Mail, int, error) {
	if recipientID == "" {
		return nil, 0, errors.New("recipient ID cannot be empty")
	}

	return m.store.GetMailsByRecipient(ctx, recipientID, page, size)
}

// QueryMails queries mails by conditions with pagination
func (m *DefaultMailManager) QueryMails(ctx context.Context, filter *MailFilter, page, size int) ([]*Mail, int, error) {
	if filter == nil {
		filter = &MailFilter{}
	}

	return m.store.QueryMails(ctx, filter, page, size)
}

// MarkAsRead marks a mail as read
func (m *DefaultMailManager) MarkAsRead(ctx context.Context, mailID string) error {
	if mailID == "" {
		return errors.New("mail ID cannot be empty")
	}

	// Get the mail first
	mail, err := m.store.GetMail(ctx, mailID)
	if err != nil {
		return err
	}

	// If already read, no need to update
	if mail.ReadStatus {
		return nil
	}

	// Mark as read and update
	mail.ReadStatus = true
	return m.store.UpdateMail(ctx, mail)
}

// MarkAllAsRead marks all user's mails as read
func (m *DefaultMailManager) MarkAllAsRead(ctx context.Context, recipientID string) error {
	if recipientID == "" {
		return errors.New("recipient ID cannot be empty")
	}

	// Fetch all user's mails
	// Note: This uses pagination internally but processes all pages to mark everything as read
	// In a real production system with many mails, this might need a direct DB update instead
	page := 1
	pageSize := 100
	totalProcessed := 0

	for {
		mails, total, err := m.store.GetMailsByRecipient(ctx, recipientID, page, pageSize)
		if err != nil {
			return err
		}

		// No more mails to process
		if len(mails) == 0 {
			break
		}

		// Mark each unread mail as read
		for _, mail := range mails {
			if !mail.ReadStatus {
				mail.ReadStatus = true
				if err := m.store.UpdateMail(ctx, mail); err != nil {
					return err
				}
			}
		}

		totalProcessed += len(mails)
		if totalProcessed >= total {
			break
		}

		page++
	}

	return nil
}

// DeleteMail deletes a mail
func (m *DefaultMailManager) DeleteMail(ctx context.Context, mailID string) error {
	if mailID == "" {
		return errors.New("mail ID cannot be empty")
	}

	return m.store.DeleteMail(ctx, mailID)
}

// DeleteMailsByRecipient deletes all user's mails
func (m *DefaultMailManager) DeleteMailsByRecipient(ctx context.Context, recipientID string) error {
	if recipientID == "" {
		return errors.New("recipient ID cannot be empty")
	}

	return m.store.DeleteMailsByRecipient(ctx, recipientID)
}

// DeleteExpiredMails deletes all expired mails
func (m *DefaultMailManager) DeleteExpiredMails(ctx context.Context) (int, error) {
	return m.store.DeleteExpiredMails(ctx, time.Now())
}

// CountUnreadMails counts unread mails for a recipient
func (m *DefaultMailManager) CountUnreadMails(ctx context.Context, recipientID string) (int, error) {
	if recipientID == "" {
		return 0, errors.New("recipient ID cannot be empty")
	}

	return m.store.CountUnreadMails(ctx, recipientID)
}

// CountMailsWithAttachments counts mails with attachments for a recipient
func (m *DefaultMailManager) CountMailsWithAttachments(ctx context.Context, recipientID string) (int, error) {
	if recipientID == "" {
		return 0, errors.New("recipient ID cannot be empty")
	}

	return m.store.CountMailsWithAttachments(ctx, recipientID)
}

// ScheduleCleanup sets up automatic cleanup of expired mails
func (m *DefaultMailManager) ScheduleCleanup(ctx context.Context, duration time.Duration) error {
	if duration <= 0 {
		return errors.New("cleanup duration must be positive")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Stop existing cleanup if running
	if m.cleanupTick != nil {
		m.cleanupTick.Stop()
		m.cleanupStop <- true
	}

	// Create new ticker
	m.cleanupTick = time.NewTicker(duration)

	// Start cleanup goroutine
	go func() {
		for {
			select {
			case <-m.cleanupTick.C:
				// Execute cleanup in a new context since the original might have expired
				cleanupCtx := context.Background()
				count, err := m.DeleteExpiredMails(cleanupCtx)
				if err != nil {
					// In a real system, we'd log this error
					fmt.Printf("Error during automatic mail cleanup: %v\n", err)
				} else if count > 0 {
					fmt.Printf("Automatic cleanup removed %d expired mails\n", count)
				}
			case <-m.cleanupStop:
				return
			}
		}
	}()

	return nil
}

// ExportMailLogs exports mail logs based on filter
func (m *DefaultMailManager) ExportMailLogs(ctx context.Context, filter *MailFilter) (string, error) {
	if filter == nil {
		filter = &MailFilter{}
	}

	return m.store.ExportMailLogs(ctx, filter)
}

// prepareMailForSending sets default values for a mail before sending
func (m *DefaultMailManager) prepareMailForSending(mail *Mail) {
	now := time.Now()

	// Set creation time if not set
	if mail.CreateTime.IsZero() {
		mail.CreateTime = now
	}

	// Initialize tags if nil
	if mail.Tags == nil {
		mail.Tags = []string{}
	}

	// Initialize attachments if nil
	if mail.Attachments == nil {
		mail.Attachments = make(map[string]interface{})
	}

	// Ensure read status is false for new mails
	mail.ReadStatus = false
}
