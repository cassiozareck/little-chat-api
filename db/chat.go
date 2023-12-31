package db

import (
	"database/sql"
	"github.com/cassiozareck/realchat/shared"
	"log"
	"time"
)

type ChatDB interface {
	CreateChat() (uint32, error)
	ChatExists(chatID uint32) (bool, error)
	Store(msg shared.Message) error
	GetMessages(chatID uint32) ([]shared.Message, error)
}

type ChatDBImp struct {
	sql *sql.DB
}

func NewChatDBImp(sql *sql.DB) *ChatDBImp {
	return &ChatDBImp{sql}
}

// CreateChat creates a new chat and returns its id.
func (c *ChatDBImp) CreateChat() (uint32, error) {
	var chatID uint32
	err := c.sql.QueryRow("INSERT INTO chat DEFAULT VALUES RETURNING id").Scan(&chatID)
	if err != nil {
		return 0, err
	}
	return chatID, nil
}

// ChatExists checks if a chat with the given id exists.
func (c *ChatDBImp) ChatExists(chatID uint32) (bool, error) {
	var exists bool
	err := c.sql.QueryRow("SELECT exists (SELECT 1 FROM chat WHERE id = $1)", chatID).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// Store will store a message in the database
func (c *ChatDBImp) Store(msg shared.Message) error {
	_, err := c.sql.Exec("INSERT INTO message (sender_id, text, timestamp, chat_id) VALUES ($1, $2, $3, $4)",
		msg.SenderID, msg.Text, msg.Timestamp, msg.ChatID)
	if shared.LOG {
		log.Printf("Stored message: %v", msg)
	}
	if err != nil {
		return err
	}
	return nil
}

// GetMessages will get all messages from a chat
func (c *ChatDBImp) GetMessages(chatID uint32) ([]shared.Message, error) {
	rows, err := c.queryMessages(chatID)
	if err != nil {
		return nil, err
	}
	defer c.closeRows(rows)

	return c.scanMessages(rows)
}

// queryMessages performs the SQL query to get messages for a chat.
func (c *ChatDBImp) queryMessages(chatID uint32) (*sql.Rows, error) {
	return c.sql.Query("SELECT id, message.text,  timestamp, chat_id, sender_id FROM message WHERE chat_id = $1", chatID)
}

// closeRows closes the SQL rows and logs any error.
func (c *ChatDBImp) closeRows(rows *sql.Rows) {
	if err := rows.Close(); err != nil {
		log.Printf("Failed to close rows: %v", err)
	}
}

// scanMessages scans the SQL rows into Message objects.
func (c *ChatDBImp) scanMessages(rows *sql.Rows) ([]shared.Message, error) {
	var messages []shared.Message
	for rows.Next() {
		msg, err := c.scanMessage(rows)
		if err != nil {
			return nil, err
		}
		messages = append(messages, *msg)
	}
	return messages, nil
}

// scanMessage scans a single SQL row into a Message object.
func (c *ChatDBImp) scanMessage(rows *sql.Rows) (*shared.Message, error) {
	var id, senderID, chatID uint32
	var text, timestamp string
	if err := rows.Scan(&id, &text, &timestamp, &chatID, &senderID); err != nil {
		return nil, err
	}

	t, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return nil, err
	}
	message := shared.Message{ID: id, SenderID: senderID, ChatID: chatID, Text: text, Timestamp: t}

	if shared.LOG {
		log.Printf("Scanned message: %v", message)
	}

	return &message, nil
}
