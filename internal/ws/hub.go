package ws

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/CollabTED/CollabTed-Backend/internal/services"
	"github.com/CollabTED/CollabTed-Backend/internal/sse"
	"github.com/CollabTED/CollabTed-Backend/pkg/logger"
	"github.com/CollabTED/CollabTed-Backend/pkg/redis"
	"github.com/CollabTED/CollabTed-Backend/pkg/types"
	"github.com/CollabTED/CollabTed-Backend/prisma/db"
	"github.com/gorilla/websocket"
)

var msgSrv = services.NewMessageService()
var wrkSrv = services.NewWorkspaceService()

type MessageType string

const (
	MessageTypeBroadcast    MessageType = "broadcast"
	MessageTypeDelete       MessageType = "delete"
	MessageTypeBoard        MessageType = "board"
	MessageTypePrivate      MessageType = "private"
	MessageTypeSystem       MessageType = "system"
	MessageTypeNotification MessageType = "notification"
)

type Connection struct {
	// msgType     MessageType
	conn *websocket.Conn
	name string
	// workspaceID string
	userID string
}

type Message struct {
	ClientID    string      `json:"clientID"`
	ID          string      `json:"id"`
	Type        MessageType `json:"type"`
	SenderID    string      `json:"senderID"`
	ChannelID   string      `json:"channelID"`
	Content     string      `json:"content"`
	WorkspaceID string      `json:"workspaceID"`

	IsReply         bool   `json:"isReply"`
	ReplyToMessage  string `json:"replyToMessage"`
	ReplyToUserName string `json:"replyToUserName"`

	AttachmentTitle string `json:"attachmentTitle"`
	AttachmentLink  string `json:"attachmentLink"`

	Elements  []json.RawMessage `json:"elements"`
	Recievers []db.UserWorkspaceModel
}

var (
	connection = make(chan Connection)
	messages   = make(chan Message)
	closing    = make(chan string)
	users      = make(map[string]Connection)
	mu         sync.RWMutex
)

var (
	notifierOnce sync.Once
	notifier     *sse.Notifier
)

func getNotifier() *sse.Notifier {
	notifierOnce.Do(func() {
		notifier = sse.NewNotifier()
	})
	return notifier
}

func Hub() {
	for {
		select {
		case con := <-connection:
			fmt.Printf("user: %s connected\n", con.userID)
			mu.Lock()
			rdb := redis.GetClient()
			rdb.HSet(context.Background(), "connected", "user:"+con.userID, con.name)
			users[con.userID] = con
			mu.Unlock()

		case msg := <-messages:
			logger.LogDebug().Msg(fmt.Sprintf("Received message: %+v", msg))
			fmt.Println(msg.Content)
			switch msg.Type {
			case MessageTypeBroadcast:
				// Handle broadcasting messages to the entire channel
				err := broadcastMessageToChannel(msg)
				if err != nil {
					log.Printf("Error broadcasting message: %v\n", err)
				}

			case MessageTypeDelete:
				// Handle deleting messages
				err := broadcastDeleteMessage(msg)
				if err != nil {
					log.Printf("Error deleting message: %v\n", err)
				}

			case MessageTypePrivate:
				// Handle sending private messages to individual users
				for _, user := range msg.Recievers {
					err := sendPrivateMessage(user.UserID, msg)
					if err != nil {
						log.Printf("Error sending private message to user %s: %v\n", user.UserID, err)
					}
				}
			case MessageTypeBoard:
				workspace, err := wrkSrv.GetWorkspaceById(msg.WorkspaceID)
				if err != nil {
					log.Printf("Error getting workspace: %v\n", err)
				}
				for _, user := range workspace.Users() {
					con, ok := users[user.UserID]
					if !ok {
						continue
					}
					err := con.conn.WriteJSON(msg)
					if err != nil {
						continue
					}
				}

			case MessageTypeNotification:
				// Handle broadcasting notifications to specific recipients
				// Extract user IDs from the Recievers slice

				// Send the notification to all recipients
				err := sendNotification(msg.Recievers, msg)
				if err != nil {
					log.Printf("Error sending notifications: %v\n", err)
				}

			case MessageTypeSystem:
				// Handle system messages (log them or take other actions)
				log.Printf("System message received: %s", msg.Content)

			default:
				log.Printf("Unknown message type: %s", msg.Type)
			}

		case id := <-closing:
			fmt.Printf("user: %s disconnected\n", id)
			mu.Lock()
			delete(users, id)
			mu.Unlock()
		}
	}
}

// func sendRoomToken(conn *websocket.Conn, token string) error {
// 	err := conn.WriteJSON(map[string]string{
// 		"token": token,
// 	})
// 	if err != nil {
// 		log.Printf("Error sending token to user: %v\n", err)
// 		return err
// 	}
// 	return nil
// }

func sendPrivateMessage(userID string, msg Message) error {
	mu.RLock()
	defer mu.RUnlock()
	con, ok := users[userID]
	if !ok {
		return fmt.Errorf("user %s not found", userID)
	}
	err := con.conn.WriteJSON(msg)
	if err != nil {
		log.Printf("Error sending private message to user %s: %v\n", userID, err)
		return err
	}
	return nil
}

func broadcastMessageToChannel(msg Message) error {
	//this notifier needs to be initialized somewhere else asap
	n := getNotifier()
	if n == nil {
		return errors.New("SSE notifier not initialized")
	}

	// Saving msgs to the db
	savedMsg, err := msgSrv.SendMessage(types.MessageD{
		Content:         msg.Content,
		SenderID:        msg.SenderID,
		ChannelID:       msg.ChannelID,
		IsReply:         msg.IsReply,
		ReplyToMessage:  msg.ReplyToMessage,
		ReplyToUserName: msg.ReplyToUserName,
		AttachmentLink:  msg.AttachmentLink,
		AttachmentTitle: msg.AttachmentTitle,
	})
	if err != nil {
		return err
	}

	mu.RLock()
	defer mu.RUnlock()

	for _, user := range msg.Recievers {
		con, ok := users[user.UserID]
		if !ok {
			continue
		}
		err := con.conn.WriteJSON(map[string]any{
			"clientID":        msg.ClientID,
			"type":            string(msg.Type),
			"id":              savedMsg.ID,
			"content":         savedMsg.Content,
			"senderID":        savedMsg.SenderID,
			"channelID":       savedMsg.ChannelID,
			"isReply":         savedMsg.IsReply,
			"replyToMessage":  savedMsg.ReplyToMessage,
			"replyToUserName": savedMsg.ReplyToUserName,
			"attachmentTitle": savedMsg.AttachmentTitle,
			"attachmentLink":  savedMsg.AttachmentLink,
		})
		if err != nil {
			log.Printf("Error sending message to user %s: %v\n", user.UserID, err)
		}
		err = n.NotifyPing(user.UserID, types.PingNotification{
			Type:     types.MESSAGE_NOTIFICATION,
			Sender:   con.name,
			Content:  savedMsg.Content,
			Channel:  savedMsg.ChannelID,
			SenderID: savedMsg.SenderID,
		})
		if err != nil {
			log.Println(err)
		}

	}
	return nil
}

func broadcastDeleteMessage(msg Message) error {
	mu.RLock()
	defer mu.RUnlock()

	for _, user := range msg.Recievers {
		con, ok := users[user.UserID]
		if !ok {
			continue
		}
		err := con.conn.WriteJSON(map[string]any{
			"type":      MessageTypeDelete,
			"id":        msg.ID,
			"channelID": msg.ChannelID,
		})
		if err != nil {
			log.Printf("Error sending delete message to user %s: %v\n", user.UserID, err)
		}
	}

	err := msgSrv.DeleteMessage(msg.SenderID, msg.ID)
	if err != nil {
		return err
	}
	return nil
}

func sendNotification(recipients []db.UserWorkspaceModel, msg Message) error {
	mu.RLock()
	defer mu.RUnlock()
	for _, user := range recipients {
		con, ok := users[user.UserID]
		if !ok {
			continue // Skip users who are not connected
		}
		err := con.conn.WriteJSON(msg)
		if err != nil {
			log.Printf("Error sending notification to user %s: %v\n", user.UserID, err)
			return err
		}
	}
	return nil
}
