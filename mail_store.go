package inboxer

import (
	"context"
	"time"
)

// MailStore defines the interface for mail storage, used for persistent storage of mail data
type MailStore interface {
	// Basic CRUD operations
	CreateMail(ctx context.Context, mail *Mail) (string, error)
	GetMail(ctx context.Context, mailID string) (*Mail, error)
	UpdateMail(ctx context.Context, mail *Mail) error
	DeleteMail(ctx context.Context, mailID string) error

	// Batch operations
	CreateBatchMails(ctx context.Context, mails []*Mail) ([]string, error)
	DeleteMailsByRecipient(ctx context.Context, recipientID string) error
	DeleteExpiredMails(ctx context.Context, beforeTime time.Time) (int, error)

	// Query operations
	GetMailsByRecipient(ctx context.Context, recipientID string, page, size int) ([]*Mail, int, error)
	QueryMails(ctx context.Context, filter *MailFilter, page, size int) ([]*Mail, int, error)

	// Count operations
	CountUnreadMails(ctx context.Context, recipientID string) (int, error)
	CountMailsWithAttachments(ctx context.Context, recipientID string) (int, error)

	// System operations
	ExportMailLogs(ctx context.Context, filter *MailFilter) (string, error)
}
