package services

import (
	"context"
	"errors"
	"fmt"
	"net/smtp"
	"time"

	"github.com/CollabTED/CollabTed-Backend/config"
	"github.com/CollabTED/CollabTed-Backend/pkg/redis"
	"github.com/CollabTED/CollabTed-Backend/pkg/types"
	"github.com/CollabTED/CollabTed-Backend/pkg/utils"
	"github.com/CollabTED/CollabTed-Backend/prisma"
	"github.com/CollabTED/CollabTed-Backend/prisma/db"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	BoardService *BoardService
}

func NewAuthService() *AuthService {
	return &AuthService{
		BoardService: NewBoardService(),
	}
}

func (s *AuthService) CreateUser(name string, email string, password string, profilePicture string, isOAuth bool) (*db.UserModel, error) {
	ctx := context.Background()

	encrypted_password, err := utils.Encrypt(password)
	if err != nil {
		return nil, err
	}
	// Check if user exists because prisma @unique does not work
	user, err := prisma.Client.User.FindFirst(
		db.User.Email.Equals(email),
	).Exec(context.Background())
	if err != nil {
		if !errors.Is(err, db.ErrNotFound) {
			return nil, err
		}
	} else {
		if user.IsOAuth {
			return nil, errors.New("user is registered with Google OAuth, please log in with Google")
		}
		return nil, errors.New("user with this email already exists")
	}

	// Create User
	result, err := prisma.Client.User.CreateOne(
		db.User.Email.Set(email),
		db.User.Name.Set(name),
		db.User.Password.Set(encrypted_password),
		db.User.ProfilePicture.Set(profilePicture),
		db.User.Active.Set(false),
		db.User.IsOAuth.Set(isOAuth),
	).Exec(ctx)

	if err != nil {
		return nil, err
	}

	// Create Personal Workspace
	personalWorkspace, err := NewWorkspaceService().CreateWorkspace(types.WorkspaceD{
		Name:    "Personal",
		OwnerID: result.ID,
	})

	if err != nil {
		return nil, err
	}

	// Create Personal UserWorkspace
	personalUserWorkspace, err := prisma.Client.UserWorkspace.CreateOne(
		db.UserWorkspace.User.Link(
			db.User.ID.Equals(result.ID),
		),
		db.UserWorkspace.Workspace.Link(
			db.Workspace.ID.Equals(personalWorkspace.ID),
		),
		db.UserWorkspace.Role.Set(db.UserRoleAdmin),
		db.UserWorkspace.JoinedAt.Set(time.Now()),
	).Exec(context.Background())
	if err != nil {
		return nil, err
	}

	// Create Personal AppState
	_, err = prisma.Client.AppState.CreateOne(
		db.AppState.UserWorkspaceID.Set(personalUserWorkspace.ID),
		db.AppState.UnreadChannels.Set([]string{}),
	).Exec(
		context.Background(),
	)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *AuthService) CheckUser(email string, password string) (*db.UserModel, error) {
	ctx := context.Background()
	client := db.NewClient()
	if err := client.Prisma.Connect(); err != nil {
		panic(err)
	}

	user, err := prisma.Client.User.FindUnique(
		db.User.Email.Equals(email),
	).Exec(ctx)
	if err != nil {
		return nil, errors.New("email not found")
	}

	if !user.Active {
		return nil, errors.New("user not activated")
	}

	if user.IsOAuth {
		return nil, errors.New("user is registered with google oauth, please log in with google")
	}

	enc_pass := user.Password
	err = utils.CheckPassword(enc_pass, password)
	if err != nil {
		return nil, errors.New("wrong password")
	}

	return user, nil
}

func (s *AuthService) GetUserByEmail(email string) (*db.UserModel, error) {
	ctx := context.Background()
	client := db.NewClient()
	if err := client.Prisma.Connect(); err != nil {
		return nil, err
	}

	user, err := client.User.FindUnique(
		db.User.Email.Equals(email),
	).Exec(ctx)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return nil, nil // User not found
		}
		return nil, err // Other errors
	}

	return user, nil // User found
}

func (s *AuthService) ActivateUser(userID string) error {
	ctx := context.Background()

	_, err := prisma.Client.User.FindUnique(
		db.User.ID.Equals(userID),
	).Update(
		db.User.Active.Set(true),
	).Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (s *AuthService) SendRessetLink(email string) error {
	//Check if user exists
	user, err := prisma.Client.User.FindFirst(
		db.User.Email.Equals(email),
	).Exec(context.Background())

	if user.IsOAuth {
		return errors.New("user is registered with google oauth, please log in with google")
	}
	if err != nil {
		return errors.New("no user found with this email")
	}

	// Generate a secure reset token
	token, err := utils.GenerateResetToken(20)
	if err != nil {
		return err
	}

	// Save token to Redis with 1-hour expiration
	r := redis.GetClient()
	err = r.Set(context.Background(), "reset:"+email, token, time.Hour*1).Err()
	if err != nil {
		return err
	}

	// Prepare the password reset link
	link := fmt.Sprintf("%s/auth/password-reset?token=%s", config.HOST_URL, token)

	// Prepare the email content
	subject := "Password Reset Request"
	body := fmt.Sprintf("Your password reset link is:\n\n%s\n\nThis link is valid for 1 hour.", link)
	message := fmt.Sprintf("Subject: %s\r\n\r\n%s", subject, body)

	// Email server configuration
	smtpHost := config.EMAIL_HOST
	smtpPort := config.EMAIL_PORT
	auth := smtp.PlainAuth("", config.EMAIL, config.EMAIL_PASSWORD, smtpHost)

	// Send the email
	err = smtp.SendMail(smtpHost+":"+smtpPort, auth, config.EMAIL, []string{email}, []byte(message))
	if err != nil {
		return err
	}

	fmt.Println("Password reset email sent to:", user.Email)
	return nil
}

func (s *AuthService) RessetPassword(email, token, new_password string) error {
	r := redis.GetClient()
	userToken, err := r.Get(context.Background(), "reset:"+email).Result()
	if err != nil {
		return errors.New("invalid token")
	}
	if userToken != token {
		return errors.New("invalid token")
	}

	encNew, err := bcrypt.GenerateFromPassword([]byte(new_password), 12)
	if err != nil {
		return err
	}

	_, err = prisma.Client.User.FindUnique(
		db.User.Email.Equals(email),
	).Update(
		db.User.Password.Set(string(encNew)),
	).Exec(context.Background())
	if err != nil {
		return err
	}

	return nil

}
