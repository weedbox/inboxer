# Inboxer

Inboxer is a flexible, lightweight in-game mail system written in Go. It provides a comprehensive solution for handling in-game notifications, messages, and rewards for games and applications.

## Features

- **Core Functionality**
  - Create, retrieve, update, and delete mail messages
  - Support for mail attachments (items, coins, etc.)
  - Mail tagging for categorization
  - Message expiration management
  - Read status tracking

- **Advanced Features**
  - Batch mail operations for sending to multiple recipients
  - System-wide announcements
  - Powerful query filtering
  - Pagination support for large mailboxes
  - Automatic cleanup of expired messages
  - Mail logging and export

## Installation

```bash
go get github.com/yourusername/inboxer
```

## Quick Start

```go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/yourusername/inboxer"
)

func main() {
	// Create a memory-based mail store
	store := inboxer.NewMemoryMailStore()
	
	// Initialize mail manager with the store
	manager := inboxer.NewDefaultMailManager(store)
	
	// Create a context
	ctx := context.Background()
	
	// Create a mail message
	mail := &inboxer.Mail{
		SenderID:    "system",
		RecipientID: "player123",
		Title:       "Welcome to the game!",
		Content:     "Thank you for joining our game. Here's a welcome gift!",
		Attachments: map[string]interface{}{
			"coins": 100,
			"item":  "starter_weapon",
		},
		Tags: []string{"welcome", "reward"},
	}
	
	// Send the mail
	mailID, err := manager.SendMail(ctx, mail)
	if err != nil {
		fmt.Printf("Error sending mail: %v\n", err)
		return
	}
	
	fmt.Printf("Mail sent successfully with ID: %s\n", mailID)
	
	// Schedule automatic cleanup of expired mails
	manager.ScheduleCleanup(ctx, 24*time.Hour)
}
```

## Component Overview

### Mail Structure

The `Mail` struct represents an individual mail message:

```go
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
	Tags        []string               // Tags for mail categorization
}
```

### Mail Store

The `MailStore` interface defines the storage layer operations:

```go
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
```

### Mail Manager

The `MailManager` interface provides high-level operations for the mail system:

```go
type MailManager interface {
	// Mail sending operations
	SendMail(ctx context.Context, mail *Mail) (string, error)
	SendBatchMail(ctx context.Context, mail *Mail, recipientIDs []string) ([]string, error)
	SendSystemAnnouncement(ctx context.Context, mail *Mail) (string, error)
	
	// Mail query operations
	GetMailByID(ctx context.Context, mailID string) (*Mail, error)
	GetMailsByRecipient(ctx context.Context, recipientID string, page, size int) ([]*Mail, int, error)
	QueryMails(ctx context.Context, filter *MailFilter, page, size int) ([]*Mail, int, error)
	
	// Mail action operations
	MarkAsRead(ctx context.Context, mailID string) error
	MarkAllAsRead(ctx context.Context, recipientID string) error
	
	// Mail management operations
	DeleteMail(ctx context.Context, mailID string) error
	DeleteMailsByRecipient(ctx context.Context, recipientID string) error
	DeleteExpiredMails(ctx context.Context) (int, error)
	
	// Mail statistics operations
	CountUnreadMails(ctx context.Context, recipientID string) (int, error)
	CountMailsWithAttachments(ctx context.Context, recipientID string) (int, error)
	
	// System operations
	ScheduleCleanup(ctx context.Context, duration time.Duration) error
	ExportMailLogs(ctx context.Context, filter *MailFilter) (string, error)
}
```

## Advanced Usage

### Filtering Mails

The `MailFilter` struct allows for sophisticated mail querying:

```go
filter := &inboxer.MailFilter{
	SenderID:    "system",
	RecipientID: "player123",
	ReadStatus:  &readStatus, // Pass a bool pointer
	StartTime:   &startTime,  // Pass a time.Time pointer
	EndTime:     &endTime,    // Pass a time.Time pointer
	ExpiredOnly: false,
	Tags:        []string{"important", "notification"},
}

mails, count, err := manager.QueryMails(ctx, filter, 1, 10)
```

### Batch Operations

Send the same mail to multiple recipients:

```go
recipients := []string{"player1", "player2", "player3"}
mailIDs, err := manager.SendBatchMail(ctx, mail, recipients)
```

### System Announcements

Send a message to all players:

```go
announcement := &inboxer.Mail{
	Title:   "Server Maintenance",
	Content: "The server will be down for maintenance on Saturday, 10:00 UTC.",
	Tags:    []string{"announcement", "important"},
}

announcementID, err := manager.SendSystemAnnouncement(ctx, announcement)
```

## Storage Implementations

### Memory Store

The package includes a memory-based implementation of the `MailStore` interface. This is perfect for testing or small applications:

```go
store := inboxer.NewMemoryMailStore()
manager := inboxer.NewDefaultMailManager(store)
```

### Custom Storage

To implement your own storage backend (e.g., for a database), implement the `MailStore` interface with your custom logic.

## License

This project is licensed under the Apache License 2.0. See the LICENSE file for details.
