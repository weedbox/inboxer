package inboxer

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewDefaultMailManager(t *testing.T) {
	store := NewMemoryMailStore()
	manager := NewDefaultMailManager(store)

	assert.NotNil(t, manager)
	assert.NotNil(t, manager.store)
	assert.NotNil(t, manager.cleanupStop)
}

func TestSendMail(t *testing.T) {
	// Initialize store and manager
	store := NewMemoryMailStore()
	manager := NewDefaultMailManager(store)
	ctx := context.Background()

	// Test sending a valid mail
	mail := &Mail{
		SenderID:    "system",
		RecipientID: "user1",
		Title:       "Test Mail",
		Content:     "This is a test mail",
		Attachments: map[string]interface{}{
			"coins": 100,
		},
	}

	id, err := manager.SendMail(ctx, mail)
	assert.NoError(t, err)
	assert.NotEmpty(t, id)

	// Verify the mail was stored correctly
	storedMail, err := manager.GetMailByID(ctx, id)
	assert.NoError(t, err)
	assert.Equal(t, mail.SenderID, storedMail.SenderID)
	assert.Equal(t, mail.RecipientID, storedMail.RecipientID)
	assert.Equal(t, mail.Title, storedMail.Title)
	assert.Equal(t, mail.Content, storedMail.Content)
	assert.Equal(t, false, storedMail.ReadStatus) // Should be set to unread
	assert.NotZero(t, storedMail.CreateTime)      // Should have a creation time
	assert.Contains(t, storedMail.Attachments, "coins")
	assert.Equal(t, 100, storedMail.Attachments["coins"])

	// Test sending nil mail
	_, err = manager.SendMail(ctx, nil)
	assert.Error(t, err)
}

func TestSendBatchMail(t *testing.T) {
	// Initialize store and manager
	store := NewMemoryMailStore()
	manager := NewDefaultMailManager(store)
	ctx := context.Background()

	// Create a template mail
	mail := &Mail{
		SenderID: "system",
		Title:    "Batch Test Mail",
		Content:  "This is a batch test mail",
		Attachments: map[string]interface{}{
			"coins": 50,
		},
		Tags: []string{"batch", "test"},
	}

	// Define recipients
	recipients := []string{"user1", "user2", "user3"}

	// Send batch mail
	ids, err := manager.SendBatchMail(ctx, mail, recipients)
	assert.NoError(t, err)
	assert.Equal(t, len(recipients), len(ids))

	// Verify each recipient received the mail
	for i, id := range ids {
		storedMail, err := manager.GetMailByID(ctx, id)
		assert.NoError(t, err)
		assert.Equal(t, "system", storedMail.SenderID)
		assert.Equal(t, recipients[i], storedMail.RecipientID)
		assert.Equal(t, mail.Title, storedMail.Title)
		assert.Equal(t, mail.Content, storedMail.Content)
		assert.Equal(t, false, storedMail.ReadStatus)
		assert.NotZero(t, storedMail.CreateTime)
		assert.Equal(t, len(mail.Tags), len(storedMail.Tags))
		assert.Contains(t, storedMail.Attachments, "coins")
		assert.Equal(t, 50, storedMail.Attachments["coins"])
	}

	// Test with empty recipients
	emptyIds, err := manager.SendBatchMail(ctx, mail, []string{})
	assert.NoError(t, err)
	assert.Empty(t, emptyIds)

	// Test with nil mail
	_, err = manager.SendBatchMail(ctx, nil, recipients)
	assert.Error(t, err)
}

func TestSendSystemAnnouncement(t *testing.T) {
	// Initialize store and manager
	store := NewMemoryMailStore()
	manager := NewDefaultMailManager(store)
	ctx := context.Background()

	// Create announcement mail
	mail := &Mail{
		Title:   "System Announcement",
		Content: "This is a system-wide announcement",
		Tags:    []string{"important"},
	}

	// Send system announcement
	id, err := manager.SendSystemAnnouncement(ctx, mail)
	assert.NoError(t, err)
	assert.NotEmpty(t, id)

	// Verify announcement was created correctly
	announcement, err := manager.GetMailByID(ctx, id)
	assert.NoError(t, err)
	assert.Equal(t, "system", announcement.SenderID)
	assert.Equal(t, "all_players", announcement.RecipientID)
	assert.Equal(t, mail.Title, announcement.Title)
	assert.Equal(t, mail.Content, announcement.Content)
	assert.Contains(t, announcement.Tags, "system_announcement")
	assert.Contains(t, announcement.Tags, "important")

	// Test with nil mail
	_, err = manager.SendSystemAnnouncement(ctx, nil)
	assert.Error(t, err)
}

func TestGetMailByID(t *testing.T) {
	// Initialize store and manager
	store := NewMemoryMailStore()
	manager := NewDefaultMailManager(store)
	ctx := context.Background()

	// Create a test mail
	mail := &Mail{
		SenderID:    "system",
		RecipientID: "user1",
		Title:       "Test Mail",
		Content:     "This is a test mail",
	}

	// Send mail
	id, err := manager.SendMail(ctx, mail)
	assert.NoError(t, err)

	// Test retrieving mail by valid ID
	retrievedMail, err := manager.GetMailByID(ctx, id)
	assert.NoError(t, err)
	assert.Equal(t, id, retrievedMail.ID)
	assert.Equal(t, mail.Title, retrievedMail.Title)

	// Test retrieving mail with empty ID
	_, err = manager.GetMailByID(ctx, "")
	assert.Error(t, err)

	// Test retrieving non-existent mail
	_, err = manager.GetMailByID(ctx, "non-existent-id")
	assert.Error(t, err)
}

func TestGetMailsByRecipient(t *testing.T) {
	// Initialize store and manager
	store := NewMemoryMailStore()
	manager := NewDefaultMailManager(store)
	ctx := context.Background()

	// Create test mails for different recipients
	now := time.Now()
	oldestTime := now.Add(-2 * time.Hour)
	olderTime := now.Add(-1 * time.Hour)
	newerTime := now.Add(1 * time.Hour)

	mails := []*Mail{
		{
			SenderID:    "system",
			RecipientID: "user1",
			Title:       "Oldest Mail",
			Content:     "Content 1",
			CreateTime:  oldestTime,
		},
		{
			SenderID:    "system",
			RecipientID: "user1",
			Title:       "Older Mail",
			Content:     "Content 2",
			CreateTime:  olderTime,
		},
		{
			SenderID:    "system",
			RecipientID: "user1",
			Title:       "Newer Mail",
			Content:     "Content 3",
			CreateTime:  newerTime,
		},
		{
			SenderID:    "system",
			RecipientID: "user2",
			Title:       "Other User Mail",
			Content:     "Content 4",
			CreateTime:  now,
		},
	}

	// Send all mails
	for _, mail := range mails {
		_, err := manager.SendMail(ctx, mail)
		assert.NoError(t, err)
	}

	// Test getting all mails for user1
	user1Mails, count, err := manager.GetMailsByRecipient(ctx, "user1", 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, 3, count)
	assert.Equal(t, 3, len(user1Mails))

	// Verify sort order (newest first)
	assert.Equal(t, "Newer Mail", user1Mails[0].Title)
	assert.Equal(t, "Older Mail", user1Mails[1].Title)
	assert.Equal(t, "Oldest Mail", user1Mails[2].Title)

	// Test pagination
	user1MailsPage1, count, err := manager.GetMailsByRecipient(ctx, "user1", 1, 2)
	assert.NoError(t, err)
	assert.Equal(t, 3, count) // Total count should still be 3
	assert.Equal(t, 2, len(user1MailsPage1))

	// Test with empty recipient ID
	_, _, err = manager.GetMailsByRecipient(ctx, "", 1, 10)
	assert.Error(t, err)
}

func TestQueryMails(t *testing.T) {
	// Initialize store and manager
	store := NewMemoryMailStore()
	manager := NewDefaultMailManager(store)
	ctx := context.Background()

	// Create test mails with various attributes
	now := time.Now()
	yesterdayTime := now.Add(-24 * time.Hour)
	tomorrowTime := now.Add(24 * time.Hour)
	expiredTime := now.Add(-1 * time.Hour)

	readStatus := true
	unreadStatus := false

	mails := []*Mail{
		{
			SenderID:    "system",
			RecipientID: "user1",
			Title:       "System Notification",
			ReadStatus:  readStatus,
			CreateTime:  yesterdayTime,
			Tags:        []string{"system", "notification"},
		},
		{
			SenderID:    "player1",
			RecipientID: "user1",
			Title:       "Player Message",
			ReadStatus:  unreadStatus,
			CreateTime:  now,
			Tags:        []string{"player", "message"},
		},
		{
			SenderID:    "system",
			RecipientID: "user2",
			Title:       "Expired Mail",
			ReadStatus:  unreadStatus,
			CreateTime:  yesterdayTime,
			ExpireTime:  expiredTime,
			Tags:        []string{"system", "expired"},
		},
		{
			SenderID:    "admin",
			RecipientID: "user2",
			Title:       "Admin Notice",
			ReadStatus:  readStatus,
			CreateTime:  tomorrowTime,
			Tags:        []string{"admin", "important"},
		},
	}

	// Send all mails
	for _, mail := range mails {
		_, err := manager.SendMail(ctx, mail)
		assert.NoError(t, err)
	}

	// Test filtering by sender
	systemMails, count, err := manager.QueryMails(ctx, &MailFilter{SenderID: "system"}, 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, 2, count)
	for _, mail := range systemMails {
		assert.Equal(t, "system", mail.SenderID)
	}

	// Test filtering by recipient
	user1Mails, count, err := manager.QueryMails(ctx, &MailFilter{RecipientID: "user1"}, 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, 2, count)
	for _, mail := range user1Mails {
		assert.Equal(t, "user1", mail.RecipientID)
	}

	// Test filtering by tags
	taggedMails, count, err := manager.QueryMails(ctx, &MailFilter{Tags: []string{"important"}}, 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
	assert.Equal(t, "Admin Notice", taggedMails[0].Title)

	// Test with multiple filters
	complexQueryMails, count, err := manager.QueryMails(ctx, &MailFilter{
		RecipientID: "user1",
		ReadStatus:  &unreadStatus,
		Tags:        []string{"player"},
	}, 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
	assert.Equal(t, "Player Message", complexQueryMails[0].Title)

	// Test with nil filter
	allMails, count, err := manager.QueryMails(ctx, nil, 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, 4, count)
	assert.Equal(t, 4, len(allMails))

	// Check if all sent mails are included in the result
	mailTitles := []string{}
	for _, mail := range allMails {
		mailTitles = append(mailTitles, mail.Title)
	}
	assert.Contains(t, mailTitles, "System Notification")
	assert.Contains(t, mailTitles, "Player Message")
	assert.Contains(t, mailTitles, "Expired Mail")
	assert.Contains(t, mailTitles, "Admin Notice")

	// Verify sorting is correct (by creation time, newest first)
	assert.Equal(t, "Admin Notice", allMails[0].Title)   // tomorrowTime
	assert.Equal(t, "Player Message", allMails[1].Title) // now

	// Both of these mails have yesterdayTime, so their order might vary
	// Just check that both exist in positions 2 and 3
	lastTwoTitles := []string{allMails[2].Title, allMails[3].Title}
	assert.Contains(t, lastTwoTitles, "System Notification")
	assert.Contains(t, lastTwoTitles, "Expired Mail")
}

func TestMarkAsRead(t *testing.T) {
	// Initialize store and manager
	store := NewMemoryMailStore()
	manager := NewDefaultMailManager(store)
	ctx := context.Background()

	// Create an unread mail
	mail := &Mail{
		SenderID:    "system",
		RecipientID: "user1",
		Title:       "Unread Mail",
		Content:     "This is an unread mail",
		ReadStatus:  false,
	}

	// Send mail
	id, err := manager.SendMail(ctx, mail)
	assert.NoError(t, err)

	// Verify mail is unread
	unreadMail, err := manager.GetMailByID(ctx, id)
	assert.NoError(t, err)
	assert.False(t, unreadMail.ReadStatus)

	// Mark mail as read
	err = manager.MarkAsRead(ctx, id)
	assert.NoError(t, err)

	// Verify mail is now read
	readMail, err := manager.GetMailByID(ctx, id)
	assert.NoError(t, err)
	assert.True(t, readMail.ReadStatus)

	// Mark already read mail as read again (should not error)
	err = manager.MarkAsRead(ctx, id)
	assert.NoError(t, err)

	// Test with empty mail ID
	err = manager.MarkAsRead(ctx, "")
	assert.Error(t, err)

	// Test with non-existent mail ID
	err = manager.MarkAsRead(ctx, "non-existent-id")
	assert.Error(t, err)
}

func TestMarkAllAsRead(t *testing.T) {
	// Initialize store and manager
	store := NewMemoryMailStore()
	manager := NewDefaultMailManager(store)
	ctx := context.Background()

	// Create several unread mails for the same recipient
	mails := []*Mail{
		{
			SenderID:    "system",
			RecipientID: "user1",
			Title:       "Unread Mail 1",
			ReadStatus:  false,
		},
		{
			SenderID:    "system",
			RecipientID: "user1",
			Title:       "Unread Mail 2",
			ReadStatus:  false,
		},
		{
			SenderID:    "system",
			RecipientID: "user1",
			Title:       "Read Mail",
			ReadStatus:  true,
		},
		{
			SenderID:    "system",
			RecipientID: "user2",
			Title:       "Other User Mail",
			ReadStatus:  false,
		},
	}

	// Send all mails
	for _, mail := range mails {
		_, err := manager.SendMail(ctx, mail)
		assert.NoError(t, err)
	}

	// Mark all mails as read for user1
	err := manager.MarkAllAsRead(ctx, "user1")
	assert.NoError(t, err)

	// Verify all mails for user1 are now read
	user1Mails, _, err := manager.GetMailsByRecipient(ctx, "user1", 1, 10)
	assert.NoError(t, err)
	for _, mail := range user1Mails {
		assert.True(t, mail.ReadStatus)
	}

	// Verify user2's mail is still unread
	user2Mails, _, err := manager.GetMailsByRecipient(ctx, "user2", 1, 10)
	assert.NoError(t, err)
	assert.False(t, user2Mails[0].ReadStatus)

	// Test with empty recipient ID
	err = manager.MarkAllAsRead(ctx, "")
	assert.Error(t, err)
}

func TestDeleteMail(t *testing.T) {
	// Initialize store and manager
	store := NewMemoryMailStore()
	manager := NewDefaultMailManager(store)
	ctx := context.Background()

	// Create a test mail
	mail := &Mail{
		SenderID:    "system",
		RecipientID: "user1",
		Title:       "Test Mail",
		Content:     "This is a test mail",
	}

	// Send mail
	id, err := manager.SendMail(ctx, mail)
	assert.NoError(t, err)

	// Delete mail
	err = manager.DeleteMail(ctx, id)
	assert.NoError(t, err)

	// Verify mail is deleted
	_, err = manager.GetMailByID(ctx, id)
	assert.Error(t, err)

	// Test with empty mail ID
	err = manager.DeleteMail(ctx, "")
	assert.Error(t, err)

	// Test with non-existent mail ID
	err = manager.DeleteMail(ctx, "non-existent-id")
	assert.Error(t, err)
}

func TestDeleteMailsByRecipient(t *testing.T) {
	// Initialize store and manager
	store := NewMemoryMailStore()
	manager := NewDefaultMailManager(store)
	ctx := context.Background()

	// Create test mails for different recipients
	mails := []*Mail{
		{
			SenderID:    "system",
			RecipientID: "user1",
			Title:       "User1 Mail 1",
		},
		{
			SenderID:    "system",
			RecipientID: "user1",
			Title:       "User1 Mail 2",
		},
		{
			SenderID:    "system",
			RecipientID: "user2",
			Title:       "User2 Mail",
		},
	}

	// Send all mails
	for _, mail := range mails {
		_, err := manager.SendMail(ctx, mail)
		assert.NoError(t, err)
	}

	// Delete all mails for user1
	err := manager.DeleteMailsByRecipient(ctx, "user1")
	assert.NoError(t, err)

	// Verify user1's mails are deleted
	user1Mails, count, err := manager.GetMailsByRecipient(ctx, "user1", 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, 0, count)
	assert.Empty(t, user1Mails)

	// Verify user2's mail is still there
	user2Mails, count, err := manager.GetMailsByRecipient(ctx, "user2", 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
	assert.Equal(t, "User2 Mail", user2Mails[0].Title)

	// Test with empty recipient ID
	err = manager.DeleteMailsByRecipient(ctx, "")
	assert.Error(t, err)
}

func TestDeleteExpiredMails(t *testing.T) {
	// Initialize store and manager
	store := NewMemoryMailStore()
	manager := NewDefaultMailManager(store)
	ctx := context.Background()

	// Create mails with different expiration times
	now := time.Now()
	pastExpiry := now.Add(-1 * time.Hour)
	futureExpiry := now.Add(24 * time.Hour)

	mails := []*Mail{
		{
			SenderID:    "system",
			RecipientID: "user1",
			Title:       "Expired Mail 1",
			ExpireTime:  pastExpiry,
		},
		{
			SenderID:    "system",
			RecipientID: "user1",
			Title:       "Expired Mail 2",
			ExpireTime:  pastExpiry,
		},
		{
			SenderID:    "system",
			RecipientID: "user2",
			Title:       "Future Mail",
			ExpireTime:  futureExpiry,
		},
		{
			SenderID:    "system",
			RecipientID: "user3",
			Title:       "No Expiry Mail",
		},
	}

	// Send all mails
	for _, mail := range mails {
		_, err := manager.SendMail(ctx, mail)
		assert.NoError(t, err)
	}

	// Delete expired mails
	count, err := manager.DeleteExpiredMails(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 2, count)

	// Verify only expired mails are deleted
	allMails, total, err := manager.QueryMails(ctx, &MailFilter{}, 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, 2, total)

	// Verify the remaining mail titles
	titles := []string{}
	for _, mail := range allMails {
		titles = append(titles, mail.Title)
	}
	assert.Contains(t, titles, "Future Mail")
	assert.Contains(t, titles, "No Expiry Mail")
	assert.NotContains(t, titles, "Expired Mail 1")
	assert.NotContains(t, titles, "Expired Mail 2")
}

func TestCountUnreadMails(t *testing.T) {
	// Initialize store and manager
	store := NewMemoryMailStore()
	manager := NewDefaultMailManager(store)
	ctx := context.Background()

	// Create test mails with different read statuses
	mails := []*Mail{
		{
			SenderID:    "system",
			RecipientID: "user1",
			Title:       "Unread Mail 1",
			ReadStatus:  false,
		},
		{
			SenderID:    "system",
			RecipientID: "user1",
			Title:       "Unread Mail 2",
			ReadStatus:  false,
		},
		{
			SenderID:    "system",
			RecipientID: "user1",
			Title:       "Read Mail",
			ReadStatus:  true, // This will be set to false by SendMail
		},
		{
			SenderID:    "system",
			RecipientID: "user2",
			Title:       "Other User Unread Mail",
			ReadStatus:  false,
		},
	}

	// Send all mails and keep track of IDs
	mailIDs := make(map[string]string)
	for _, mail := range mails {
		id, err := manager.SendMail(ctx, mail)
		assert.NoError(t, err)
		mailIDs[mail.Title] = id
	}

	// Mark the "Read Mail" as read to match our original intent
	err := manager.MarkAsRead(ctx, mailIDs["Read Mail"])
	assert.NoError(t, err)

	// Test count for user1
	user1Count, err := manager.CountUnreadMails(ctx, "user1")
	assert.NoError(t, err)
	assert.Equal(t, 2, user1Count)

	// Test count for user2
	user2Count, err := manager.CountUnreadMails(ctx, "user2")
	assert.NoError(t, err)
	assert.Equal(t, 1, user2Count)

	// Test count for non-existent user
	nonExistentCount, err := manager.CountUnreadMails(ctx, "non-existent-user")
	assert.NoError(t, err)
	assert.Equal(t, 0, nonExistentCount)

	// Test with empty recipient ID
	_, err = manager.CountUnreadMails(ctx, "")
	assert.Error(t, err)
}

func TestCountMailsWithAttachments(t *testing.T) {
	// Initialize store and manager
	store := NewMemoryMailStore()
	manager := NewDefaultMailManager(store)
	ctx := context.Background()

	// Create test mails with and without attachments
	mails := []*Mail{
		{
			SenderID:    "system",
			RecipientID: "user1",
			Title:       "Mail With Attachment 1",
			Attachments: map[string]interface{}{
				"coins": 100,
			},
		},
		{
			SenderID:    "system",
			RecipientID: "user1",
			Title:       "Mail With Attachment 2",
			Attachments: map[string]interface{}{
				"item": "sword",
			},
		},
		{
			SenderID:    "system",
			RecipientID: "user1",
			Title:       "Mail Without Attachment",
		},
		{
			SenderID:    "system",
			RecipientID: "user1",
			Title:       "Mail With Empty Attachment",
			Attachments: map[string]interface{}{},
		},
		{
			SenderID:    "system",
			RecipientID: "user2",
			Title:       "Other User Mail With Attachment",
			Attachments: map[string]interface{}{
				"coins": 50,
			},
		},
	}

	// Send all mails
	for _, mail := range mails {
		_, err := manager.SendMail(ctx, mail)
		assert.NoError(t, err)
	}

	// Test count for user1
	user1Count, err := manager.CountMailsWithAttachments(ctx, "user1")
	assert.NoError(t, err)
	assert.Equal(t, 2, user1Count) // Empty attachments map should not count

	// Test count for user2
	user2Count, err := manager.CountMailsWithAttachments(ctx, "user2")
	assert.NoError(t, err)
	assert.Equal(t, 1, user2Count)

	// Test count for non-existent user
	nonExistentCount, err := manager.CountMailsWithAttachments(ctx, "non-existent-user")
	assert.NoError(t, err)
	assert.Equal(t, 0, nonExistentCount)

	// Test with empty recipient ID
	_, err = manager.CountMailsWithAttachments(ctx, "")
	assert.Error(t, err)
}

func TestScheduleCleanup(t *testing.T) {
	// Initialize store and manager
	store := NewMemoryMailStore()
	manager := NewDefaultMailManager(store)
	ctx := context.Background()

	// Create expired mails
	now := time.Now()
	pastExpiry := now.Add(-1 * time.Hour)

	mails := []*Mail{
		{
			SenderID:    "system",
			RecipientID: "user1",
			Title:       "Expired Mail 1",
			ExpireTime:  pastExpiry,
		},
		{
			SenderID:    "system",
			RecipientID: "user1",
			Title:       "Expired Mail 2",
			ExpireTime:  pastExpiry,
		},
	}

	// Send all mails
	for _, mail := range mails {
		_, err := manager.SendMail(ctx, mail)
		assert.NoError(t, err)
	}

	// Schedule cleanup with short duration
	err := manager.ScheduleCleanup(ctx, 100*time.Millisecond)
	assert.NoError(t, err)

	// Wait for cleanup to run
	time.Sleep(200 * time.Millisecond)

	// Verify expired mails were deleted
	allMails, total, err := manager.QueryMails(ctx, &MailFilter{}, 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, 0, total)
	assert.Empty(t, allMails)

	// Clean up
	if manager.cleanupTick != nil {
		manager.cleanupTick.Stop()
		manager.cleanupStop <- true
	}

	// Test with negative duration
	err = manager.ScheduleCleanup(ctx, -1*time.Second)
	assert.Error(t, err)
}

func TestExportMailLogs(t *testing.T) {
	// Initialize store and manager
	store := NewMemoryMailStore()
	manager := NewDefaultMailManager(store)
	ctx := context.Background()

	// Create some test mails
	now := time.Now()
	mails := []*Mail{
		{
			SenderID:    "system",
			RecipientID: "user1",
			Title:       "System Mail",
			Content:     "System Content",
			CreateTime:  now,
			Tags:        []string{"system"},
		},
		{
			SenderID:    "player1",
			RecipientID: "user2",
			Title:       "Player Mail",
			Content:     "Player Content",
			CreateTime:  now,
			Tags:        []string{"player"},
		},
	}

	// Send all mails
	for _, mail := range mails {
		_, err := manager.SendMail(ctx, mail)
		assert.NoError(t, err)
	}

	// Test exporting all mails
	allLogsJSON, err := manager.ExportMailLogs(ctx, nil)
	assert.NoError(t, err)
	assert.NotEmpty(t, allLogsJSON)
	assert.Contains(t, allLogsJSON, "System Mail")
	assert.Contains(t, allLogsJSON, "Player Mail")

	// Test exporting filtered logs
	systemLogsJSON, err := manager.ExportMailLogs(ctx, &MailFilter{SenderID: "system"})
	assert.NoError(t, err)
	assert.NotEmpty(t, systemLogsJSON)
	assert.Contains(t, systemLogsJSON, "System Mail")
	assert.NotContains(t, systemLogsJSON, "Player Mail")

	// Test exporting with tag filter
	playerLogsJSON, err := manager.ExportMailLogs(ctx, &MailFilter{Tags: []string{"player"}})
	assert.NoError(t, err)
	assert.NotEmpty(t, playerLogsJSON)
	assert.Contains(t, playerLogsJSON, "Player Mail")
	assert.NotContains(t, playerLogsJSON, "System Mail")
}
