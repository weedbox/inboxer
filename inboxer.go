package inboxer

import (
	"context"
	"time"
)

// Mail represents the basic structure of system mail
type Mail struct {
	ID          string                 // Unique mail ID
	SenderID    string                 // Sender ID (system or player)
	RecipientID string                 // Recipient ID
	Title       string                 // Mail title
	Content     string                 // Mail content
	Attachments map[string]interface{} // Attachments (items, coins, etc.)
	ReadStatus  bool                   // Read status
	CreateTime  time.Time              // Creation time
	ExpireTime  time.Time              // Expiration time
	Tags        []string               // Tags (can be used for mail categorization)
}

// MailFilter defines conditions for filtering mails
type MailFilter struct {
	SenderID    string     // Filter by sender
	RecipientID string     // Filter by recipient
	ReadStatus  *bool      // Filter by read status
	StartTime   *time.Time // Filter by creation time (start)
	EndTime     *time.Time // Filter by creation time (end)
	ExpiredOnly bool       // Query only expired mails
	Tags        []string   // Filter by tags
}

// MailManager defines the interface for managing game system mails
type MailManager interface {
	// Mail sending operations
	SendMail(ctx context.Context, mail *Mail) (string, error)                               // Send a single mail, returns mail ID
	SendBatchMail(ctx context.Context, mail *Mail, recipientIDs []string) ([]string, error) // Send the same mail content to multiple recipients
	SendSystemAnnouncement(ctx context.Context, mail *Mail) (string, error)                 // Send system announcement (to all players)

	// Mail query operations
	GetMailByID(ctx context.Context, mailID string) (*Mail, error)                                     // Get mail by ID
	GetMailsByRecipient(ctx context.Context, recipientID string, page, size int) ([]*Mail, int, error) // Get user's mails with pagination
	QueryMails(ctx context.Context, filter *MailFilter, page, size int) ([]*Mail, int, error)          // Query mails by conditions

	// Mail action operations
	MarkAsRead(ctx context.Context, mailID string) error         // Mark mail as read
	MarkAllAsRead(ctx context.Context, recipientID string) error // Mark all user's mails as read

	// Mail management operations
	DeleteMail(ctx context.Context, mailID string) error                  // Delete mail
	DeleteMailsByRecipient(ctx context.Context, recipientID string) error // Delete all user's mails
	DeleteExpiredMails(ctx context.Context) (int, error)                  // Delete all expired mails, returns deletion count

	// Mail statistics operations
	CountUnreadMails(ctx context.Context, recipientID string) (int, error)          // Get unread mail count
	CountMailsWithAttachments(ctx context.Context, recipientID string) (int, error) // Get count of mails with attachments

	// System operations
	ScheduleCleanup(ctx context.Context, duration time.Duration) error      // Set interval for automatic expired mail cleanup
	ExportMailLogs(ctx context.Context, filter *MailFilter) (string, error) // Export mail logs
}
