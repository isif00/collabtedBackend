package ws

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/CollabTED/CollabTed-Backend/config"
	"github.com/CollabTED/CollabTed-Backend/internal/services"
	"github.com/CollabTED/CollabTed-Backend/pkg/logger"
	"github.com/CollabTED/CollabTed-Backend/pkg/types"
	"github.com/CollabTED/CollabTed-Backend/prisma/db"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

var (
	readBufferSize  = 1024
	writeBufferSize = 1024
)

type WsChatHandler struct {
	srv services.ChannelService
}

func (ws WsChatHandler) Connections(c echo.Context) error {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  readBufferSize,
		WriteBufferSize: writeBufferSize,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}

	// Authentication handling
	cookie, err := c.Cookie("jwt")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if cookie.Value == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "token is required")
	}

	token, err := jwt.ParseWithClaims(cookie.Value, &types.Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.JWT_SECRET), nil
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	claims := token.Claims.(*types.Claims)
	conn, err := upgrader.Upgrade(c.Response().Writer, c.Request(), nil)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Get workspace ID and create user
	workspaceID := c.QueryParam("workspaceID")
	user := &User{
		UserID:      claims.ID,
		WorkspaceID: workspaceID,
		Conn:        conn,
	}

	// Register user
	online <- user

	// Setup connection lifecycle management
	done := make(chan struct{})
	defer close(done)

	// Reader goroutine (handles pongs and connection closure)
	go func() {
		defer conn.Close()
		conn.SetPongHandler(func(string) error {
			err := conn.SetReadDeadline(time.Now().Add(60 * time.Second))
			if err != nil {
				return err
			}
			return nil
		})

		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
					log.Printf("connection closed unexpectedly: %v", err)
				}
				disconnected <- user
				return
			}
		}
	}()

	// Writer goroutine (handles pings)
	go func() {
		ticker := time.NewTicker(25 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				user.WriteMu.Lock()
				err := conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(5*time.Second))
				user.WriteMu.Unlock()
				if err != nil {
					log.Println("ping failed:", err)
					return
				}
			case <-done:
				return
			}
		}
	}()

	<-done
	return nil
}

func (ws WsChatHandler) Chat(c echo.Context) error {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  readBufferSize,
		WriteBufferSize: writeBufferSize,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}

	cookie, err := c.Cookie("jwt")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if cookie.Value == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "token is required")
	}

	token, err := jwt.ParseWithClaims(cookie.Value, &types.Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.JWT_SECRET), nil
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	claims := token.Claims.(*types.Claims)

	conn, err := upgrader.Upgrade(c.Response().Writer, c.Request(), nil)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	defer conn.Close()

	// Set up ping/pong handlers
	conn.SetPongHandler(func(appData string) error {
		err := conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		if err != nil {
			return err
		}
		return nil
	})

	done := make(chan struct{})
	go func() {
		ticker := time.NewTicker(25 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(5*time.Second)); err != nil {
					log.Println("Ping failed:", err)
					return
				}
			case <-done:
				return
			}
		}
	}()

	conn.SetCloseHandler(func(code int, text string) error {
		close(done)
		closing <- claims.ID
		return nil
	})

	connection <- Connection{
		conn:   conn,
		name:   claims.Name,
		userID: claims.ID,
	}

	for {
		var data Message
		err := conn.ReadJSON(&data)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		// Handle ping messages
		if data.Type == "ping" {
			data.Type = "pong"
			if err := conn.WriteJSON(data); err != nil {
				log.Println("Failed to send pong:", err)
			}
			continue
		}

		logger.LogDebug().Msg(fmt.Sprintf("Received message: %+v", data))
		fmt.Println("received message", data)

		var channel *db.ChannelModel
		if data.ChannelID != "" {
			channel, err = ws.srv.GetChannelById(data.ChannelID)
			if err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, err.Error())
			}
			data.Recievers = channel.Participants()
			fmt.Println("the receivers are", data.Recievers)
		}
		data.SenderID = claims.ID
		messages <- data
	}
}
