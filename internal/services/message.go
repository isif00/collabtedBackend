package services

import (
	"context"
	"fmt"

	"github.com/CollabTED/CollabTed-Backend/pkg/types"
	"github.com/CollabTED/CollabTed-Backend/prisma"
	"github.com/CollabTED/CollabTed-Backend/prisma/db"
)

type MessageService struct{}

func NewMessageService() *MessageService {
	return &MessageService{}
}

func (s *MessageService) SendMessage(data types.MessageD) (*db.MessageModel, error) {
	message, err := prisma.Client.Message.CreateOne(
		db.Message.Content.Set(data.Content),
		db.Message.Channel.Link(
			db.Channel.ID.Equals(data.ChannelID),
		),

		db.Message.Sender.Link(
			db.User.ID.Equals(data.SenderID),
		),

		db.Message.ReplyToUserName.Set(data.ReplyToUserName),
		db.Message.ReplyToMessage.Set(data.ReplyToMessage),

		db.Message.AttachmentLink.Set(data.AttachmentLink),
		db.Message.AttachmentTitle.Set(data.AttachmentTitle),
		db.Message.IsReply.Set(data.IsReply),
	).Exec(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %v", err)
	}

	return message, nil
}

func (s *MessageService) GetMessagesByChannel(channelID string, page int) ([]db.MessageModel, error) {
	messages, err := prisma.Client.Message.FindMany(
		db.Message.ChannelID.Equals(channelID),
	).Skip((page - 1) * 10000).Take(10000).Exec(context.Background())
	if err != nil {
		return nil, err
	}
	return messages, nil
}
func (s *MessageService) GetAttachmentsByChannel(channelID string) ([]db.AttachmentModel, error) {
	attachments, err := prisma.Client.Attachment.FindMany(
		db.Attachment.ChannelID.Equals(channelID),
	).Exec(context.Background())
	if err != nil {
		return nil, err
	}

	return attachments, nil
}
func (s *MessageService) GetMessageById(messageID string) (*db.MessageModel, error) {
	message, err := prisma.Client.Message.FindUnique(
		db.Message.ID.Equals(messageID),
	).Exec(context.Background())
	if err != nil {
		return nil, err
	}

	return message, nil
}

func (s *MessageService) DeleteMessage(userId, messageID string) error {
	_, err := prisma.Client.Message.FindMany(
		db.Message.ID.Equals(messageID),
		db.Message.SenderID.Equals(userId),
	).Delete().Exec(context.Background())
	if err != nil {
		return err
	}

	return nil
}

func (s *MessageService) PingMessage(messageID string) error {
	_, err := prisma.Client.Message.FindUnique(
		db.Message.ID.Equals(messageID),
	).Update(
		db.Message.IsPined.Set(true),
	).Exec(context.Background())

	if err != nil {
		return err
	}

	return nil

}

func (s *MessageService) GetPinnedMessages(channelID string, page int) ([]db.MessageModel, error) {
	result, err := prisma.Client.Message.FindMany(
		db.Message.ChannelID.Equals(channelID),
		db.Message.IsPined.Equals(true),
	).Skip((page - 1) * 10).Take(10).Exec(context.Background())

	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *MessageService) CreateAttachment(attachment types.AttachmentD) (*db.AttachmentModel, error) {
	user, err := prisma.Client.UserWorkspace.FindFirst(
		db.UserWorkspace.UserID.Equals(attachment.SenderID),
	).Exec(context.Background())
	if err != nil {
		return nil, err
	}
	result, err := prisma.Client.Attachment.CreateOne(db.Attachment.Channel.Link(
		db.Channel.ID.Equals(attachment.ChannelID)),
		db.Attachment.UserID.Set(user.UserID),
		db.Attachment.File.Set(attachment.File),
		db.Attachment.Title.Set(attachment.Title),
	).Exec(context.Background())
	return result, err
}
func (s *MessageService) DeleteAttachment(userId, attachmentID string) error {
	_, err := prisma.Client.Attachment.FindMany(
		db.Attachment.ID.Equals(attachmentID),
		db.Attachment.UserID.Equals(userId),
	).Delete().Exec(context.Background())
	if err != nil {
		return err
	}

	return nil
}
