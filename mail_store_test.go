package inboxer

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMemoryMailStore_CreateMail(t *testing.T) {
	// Initialize store
	store := NewMemoryMailStore()
	ctx := context.Background()

	// Create test mail
	now := time.Now()
	mail := &Mail{
		SenderID:    "system",
		RecipientID: "user1",
		Title:       "Test Mail",
		Content:     "This is a test mail",
		Attachments: map[string]interface{}{
			"coins": 100,
		},
		ReadStatus: false,
		CreateTime: now,
		ExpireTime: now.Add(24 * time.Hour),
		Tags:       []string{"test", "notification"},
	}

	// Test creating mail
	id, err := store.CreateMail(ctx, mail)
	assert.NoError(t, err)
	assert.NotEmpty(t, id)
	assert.Equal(t, id, mail.ID)

	// Test creating nil mail
	_, err = store.CreateMail(ctx, nil)
	assert.Error(t, err)
}

func TestMemoryMailStore_GetMail(t *testing.T) {
	// Initialize store
	store := NewMemoryMailStore()
	ctx := context.Background()

	// Create test mail
	now := time.Now()
	mail := &Mail{
		SenderID:    "system",
		RecipientID: "user1",
		Title:       "Test Mail",
		Content:     "This is a test mail",
		Attachments: map[string]interface{}{
			"coins": 100,
		},
		ReadStatus: false,
		CreateTime: now,
		ExpireTime: now.Add(24 * time.Hour),
		Tags:       []string{"test", "notification"},
	}

	id, err := store.CreateMail(ctx, mail)
	assert.NoError(t, err)

	// Test retrieving mail
	retrievedMail, err := store.GetMail(ctx, id)
	assert.NoError(t, err)
	assert.Equal(t, mail.ID, retrievedMail.ID)
	assert.Equal(t, mail.SenderID, retrievedMail.SenderID)
	assert.Equal(t, mail.RecipientID, retrievedMail.RecipientID)
	assert.Equal(t, mail.Title, retrievedMail.Title)
	assert.Equal(t, mail.Content, retrievedMail.Content)
	assert.Equal(t, mail.ReadStatus, retrievedMail.ReadStatus)
	assert.Equal(t, len(mail.Tags), len(retrievedMail.Tags))
	assert.Contains(t, retrievedMail.Attachments, "coins")
	assert.Equal(t, mail.Attachments["coins"], retrievedMail.Attachments["coins"])

	// Test retrieving non-existent mail
	_, err = store.GetMail(ctx, "non-existent-id")
	assert.Error(t, err)
}

func TestMemoryMailStore_UpdateMail(t *testing.T) {
	// Initialize store
	store := NewMemoryMailStore()
	ctx := context.Background()

	// Create test mail
	now := time.Now()
	mail := &Mail{
		SenderID:    "system",
		RecipientID: "user1",
		Title:       "Test Mail",
		Content:     "This is a test mail",
		ReadStatus:  false,
		CreateTime:  now,
		ExpireTime:  now.Add(24 * time.Hour),
	}

	id, err := store.CreateMail(ctx, mail)
	assert.NoError(t, err)

	// Update mail
	mail.Title = "Updated Title"
	mail.Content = "Updated Content"
	mail.ReadStatus = true
	mail.Attachments = map[string]interface{}{
		"item": "sword",
	}

	err = store.UpdateMail(ctx, mail)
	assert.NoError(t, err)

	// Verify update
	updatedMail, err := store.GetMail(ctx, id)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Title", updatedMail.Title)
	assert.Equal(t, "Updated Content", updatedMail.Content)
	assert.True(t, updatedMail.ReadStatus)
	assert.Contains(t, updatedMail.Attachments, "item")
	assert.Equal(t, "sword", updatedMail.Attachments["item"])

	// Test updating non-existent mail
	nonExistentMail := &Mail{
		ID:          "non-existent-id",
		SenderID:    "system",
		RecipientID: "user1",
		Title:       "Non-existent Mail",
	}
	err = store.UpdateMail(ctx, nonExistentMail)
	assert.Error(t, err)

	// Test updating nil mail
	err = store.UpdateMail(ctx, nil)
	assert.Error(t, err)
}

func TestMemoryMailStore_DeleteMail(t *testing.T) {
	// Initialize store
	store := NewMemoryMailStore()
	ctx := context.Background()

	// Create test mail
	mail := &Mail{
		SenderID:    "system",
		RecipientID: "user1",
		Title:       "Test Mail",
		Content:     "This is a test mail",
	}

	id, err := store.CreateMail(ctx, mail)
	assert.NoError(t, err)

	// Delete mail
	err = store.DeleteMail(ctx, id)
	assert.NoError(t, err)

	// Verify deletion
	_, err = store.GetMail(ctx, id)
	assert.Error(t, err)

	// Test deleting non-existent mail
	err = store.DeleteMail(ctx, "non-existent-id")
	assert.Error(t, err)
}

func TestMemoryMailStore_CreateBatchMails(t *testing.T) {
	// Initialize store
	store := NewMemoryMailStore()
	ctx := context.Background()

	// Create test mails
	now := time.Now()
	mails := []*Mail{
		{
			SenderID:    "system",
			RecipientID: "user1",
			Title:       "Test Mail 1",
			Content:     "This is test mail 1",
			CreateTime:  now,
		},
		{
			SenderID:    "system",
			RecipientID: "user2",
			Title:       "Test Mail 2",
			Content:     "This is test mail 2",
			CreateTime:  now,
		},
		{
			SenderID:    "player1",
			RecipientID: "user3",
			Title:       "Test Mail 3",
			Content:     "This is test mail 3",
			CreateTime:  now,
		},
	}

	// Test batch creation
	ids, err := store.CreateBatchMails(ctx, mails)
	assert.NoError(t, err)
	assert.Equal(t, len(mails), len(ids))

	// Verify mails were created
	for i, id := range ids {
		mail, err := store.GetMail(ctx, id)
		assert.NoError(t, err)
		assert.Equal(t, mails[i].Title, mail.Title)
		assert.Equal(t, mails[i].RecipientID, mail.RecipientID)
	}

	// Test with empty array
	emptyIds, err := store.CreateBatchMails(ctx, []*Mail{})
	assert.NoError(t, err)
	assert.Empty(t, emptyIds)

	// Test with nil in array
	mailsWithNil := []*Mail{
		{
			SenderID:    "system",
			RecipientID: "user4",
			Title:       "Valid Mail",
		},
		nil,
	}
	idsWithNil, err := store.CreateBatchMails(ctx, mailsWithNil)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(idsWithNil))
}

func TestMemoryMailStore_DeleteMailsByRecipient(t *testing.T) {
	// Initialize store
	store := NewMemoryMailStore()
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

	_, err := store.CreateBatchMails(ctx, mails)
	assert.NoError(t, err)

	// Delete user1's mails
	err = store.DeleteMailsByRecipient(ctx, "user1")
	assert.NoError(t, err)

	// Check user1's mails are gone
	user1Mails, count, err := store.GetMailsByRecipient(ctx, "user1", 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, 0, count)
	assert.Empty(t, user1Mails)

	// Check user2's mail is still there
	user2Mails, count, err := store.GetMailsByRecipient(ctx, "user2", 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
	assert.Equal(t, "User2 Mail", user2Mails[0].Title)

	// Test with empty recipient ID
	err = store.DeleteMailsByRecipient(ctx, "")
	assert.Error(t, err)
}

func TestMemoryMailStore_DeleteExpiredMails(t *testing.T) {
	// Initialize store
	store := NewMemoryMailStore()
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

	_, err := store.CreateBatchMails(ctx, mails)
	assert.NoError(t, err)

	// Delete expired mails
	count, err := store.DeleteExpiredMails(ctx, now)
	assert.NoError(t, err)
	assert.Equal(t, 2, count)

	// Check that only non-expired mails remain
	allMails, total, err := store.QueryMails(ctx, &MailFilter{}, 1, 10)
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

func TestMemoryMailStore_GetMailsByRecipient(t *testing.T) {
	// Initialize store
	store := NewMemoryMailStore()
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
			CreateTime:  oldestTime,
		},
		{
			SenderID:    "system",
			RecipientID: "user1",
			Title:       "Older Mail",
			CreateTime:  olderTime,
		},
		{
			SenderID:    "system",
			RecipientID: "user1",
			Title:       "Newer Mail",
			CreateTime:  newerTime,
		},
		{
			SenderID:    "system",
			RecipientID: "user2",
			Title:       "Other User Mail",
			CreateTime:  now,
		},
	}

	_, err := store.CreateBatchMails(ctx, mails)
	assert.NoError(t, err)

	// Test getting all mails for user1
	user1Mails, count, err := store.GetMailsByRecipient(ctx, "user1", 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, 3, count)
	assert.Equal(t, 3, len(user1Mails))

	// Verify sort order (newest first)
	assert.Equal(t, "Newer Mail", user1Mails[0].Title)
	assert.Equal(t, "Older Mail", user1Mails[1].Title)
	assert.Equal(t, "Oldest Mail", user1Mails[2].Title)

	// Test pagination
	user1MailsPage1, count, err := store.GetMailsByRecipient(ctx, "user1", 1, 2)
	assert.NoError(t, err)
	assert.Equal(t, 3, count) // Total count should still be 3
	assert.Equal(t, 2, len(user1MailsPage1))
	assert.Equal(t, "Newer Mail", user1MailsPage1[0].Title)
	assert.Equal(t, "Older Mail", user1MailsPage1[1].Title)

	user1MailsPage2, count, err := store.GetMailsByRecipient(ctx, "user1", 2, 2)
	assert.NoError(t, err)
	assert.Equal(t, 3, count)
	assert.Equal(t, 1, len(user1MailsPage2))
	assert.Equal(t, "Oldest Mail", user1MailsPage2[0].Title)

	// Test with invalid pagination values
	user1MailsInvalidPage, count, err := store.GetMailsByRecipient(ctx, "user1", 0, 0)
	assert.NoError(t, err)
	assert.Equal(t, 3, count)
	assert.Equal(t, 3, len(user1MailsInvalidPage))

	// Test with non-existent user
	nonExistentUserMails, count, err := store.GetMailsByRecipient(ctx, "non-existent-user", 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, 0, count)
	assert.Empty(t, nonExistentUserMails)
}

func TestMemoryMailStore_QueryMails(t *testing.T) {
	// Initialize store
	store := NewMemoryMailStore()
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

	_, err := store.CreateBatchMails(ctx, mails)
	assert.NoError(t, err)

	// Test filtering by sender
	systemMails, count, err := store.QueryMails(ctx, &MailFilter{SenderID: "system"}, 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, 2, count)
	for _, mail := range systemMails {
		assert.Equal(t, "system", mail.SenderID)
	}

	// Test filtering by recipient
	user1Mails, count, err := store.QueryMails(ctx, &MailFilter{RecipientID: "user1"}, 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, 2, count)
	for _, mail := range user1Mails {
		assert.Equal(t, "user1", mail.RecipientID)
	}

	// Test filtering by read status
	readMails, count, err := store.QueryMails(ctx, &MailFilter{ReadStatus: &readStatus}, 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, 2, count)
	for _, mail := range readMails {
		assert.True(t, mail.ReadStatus)
	}

	// Test filtering by time range
	timeRangeMails, count, err := store.QueryMails(ctx, &MailFilter{
		StartTime: &yesterdayTime,
		EndTime:   &now,
	}, 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, 3, count)

	// Verify that the correct mails are included in the time range
	timeRangeMailTitles := []string{}
	for _, mail := range timeRangeMails {
		timeRangeMailTitles = append(timeRangeMailTitles, mail.Title)
		// Verify each mail is within the time range
		assert.True(t, mail.CreateTime.Equal(yesterdayTime) || mail.CreateTime.After(yesterdayTime))
		assert.True(t, mail.CreateTime.Equal(now) || mail.CreateTime.Before(now))
	}

	// Check if the expected mails are in the results
	assert.Contains(t, timeRangeMailTitles, "System Notification") // From yesterdayTime
	assert.Contains(t, timeRangeMailTitles, "Player Message")      // From now
	assert.Contains(t, timeRangeMailTitles, "Expired Mail")        // From yesterdayTime

	// Test filtering by expired mails
	expiredMails, count, err := store.QueryMails(ctx, &MailFilter{ExpiredOnly: true}, 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
	assert.Equal(t, "Expired Mail", expiredMails[0].Title)

	// Test filtering by tags
	taggedMails, count, err := store.QueryMails(ctx, &MailFilter{Tags: []string{"important"}}, 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
	assert.Equal(t, "Admin Notice", taggedMails[0].Title)

	// Test with multiple filters
	complexQueryMails, count, err := store.QueryMails(ctx, &MailFilter{
		RecipientID: "user1",
		ReadStatus:  &unreadStatus,
		Tags:        []string{"player"},
	}, 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
	assert.Equal(t, "Player Message", complexQueryMails[0].Title)

	// Test pagination
	allMails, totalCount, err := store.QueryMails(ctx, &MailFilter{}, 1, 2)
	assert.NoError(t, err)
	assert.Equal(t, 4, totalCount)
	assert.Equal(t, 2, len(allMails))

	// Test with invalid page
	outOfBoundsMails, count, err := store.QueryMails(ctx, &MailFilter{}, 10, 10)
	assert.NoError(t, err)
	assert.Equal(t, 4, count)
	assert.Empty(t, outOfBoundsMails)
}

func TestMemoryMailStore_CountUnreadMails(t *testing.T) {
	// Initialize store
	store := NewMemoryMailStore()
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
			ReadStatus:  true,
		},
		{
			SenderID:    "system",
			RecipientID: "user2",
			Title:       "Other User Unread Mail",
			ReadStatus:  false,
		},
	}

	_, err := store.CreateBatchMails(ctx, mails)
	assert.NoError(t, err)

	// Test count for user1
	user1Count, err := store.CountUnreadMails(ctx, "user1")
	assert.NoError(t, err)
	assert.Equal(t, 2, user1Count)

	// Test count for user2
	user2Count, err := store.CountUnreadMails(ctx, "user2")
	assert.NoError(t, err)
	assert.Equal(t, 1, user2Count)

	// Test count for non-existent user
	nonExistentCount, err := store.CountUnreadMails(ctx, "non-existent-user")
	assert.NoError(t, err)
	assert.Equal(t, 0, nonExistentCount)
}

func TestMemoryMailStore_CountMailsWithAttachments(t *testing.T) {
	// Initialize store
	store := NewMemoryMailStore()
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

	_, err := store.CreateBatchMails(ctx, mails)
	assert.NoError(t, err)

	// Test count for user1
	user1Count, err := store.CountMailsWithAttachments(ctx, "user1")
	assert.NoError(t, err)
	assert.Equal(t, 2, user1Count) // Empty attachments map should not count

	// Test count for user2
	user2Count, err := store.CountMailsWithAttachments(ctx, "user2")
	assert.NoError(t, err)
	assert.Equal(t, 1, user2Count)

	// Test count for non-existent user
	nonExistentCount, err := store.CountMailsWithAttachments(ctx, "non-existent-user")
	assert.NoError(t, err)
	assert.Equal(t, 0, nonExistentCount)
}

func TestMemoryMailStore_ExportMailLogs(t *testing.T) {
	// Initialize store
	store := NewMemoryMailStore()
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

	_, err := store.CreateBatchMails(ctx, mails)
	assert.NoError(t, err)

	// Test exporting all mails
	allLogsJSON, err := store.ExportMailLogs(ctx, &MailFilter{})
	assert.NoError(t, err)
	assert.NotEmpty(t, allLogsJSON)
	assert.Contains(t, allLogsJSON, "System Mail")
	assert.Contains(t, allLogsJSON, "Player Mail")

	// Test exporting filtered logs
	systemLogsJSON, err := store.ExportMailLogs(ctx, &MailFilter{SenderID: "system"})
	assert.NoError(t, err)
	assert.NotEmpty(t, systemLogsJSON)
	assert.Contains(t, systemLogsJSON, "System Mail")
	assert.NotContains(t, systemLogsJSON, "Player Mail")

	// Test exporting with tag filter
	playerLogsJSON, err := store.ExportMailLogs(ctx, &MailFilter{Tags: []string{"player"}})
	assert.NoError(t, err)
	assert.NotEmpty(t, playerLogsJSON)
	assert.Contains(t, playerLogsJSON, "Player Mail")
	assert.NotContains(t, playerLogsJSON, "System Mail")
}
