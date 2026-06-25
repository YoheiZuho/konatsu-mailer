package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/coder/websocket"
	"github.com/gin-gonic/gin"

	"github.com/yoheizuho/konatsu-mailer/internal/config"
	"github.com/yoheizuho/konatsu-mailer/internal/ws"
)

// wsUpgradeHandler upgrades GET /api/ws to a WebSocket and streams the user's
// realtime events. The bearer token is passed via the Sec-WebSocket-Protocol
// header as "bearer, <token>" (design §8), since browsers cannot set arbitrary
// headers on a WebSocket handshake.
func wsUpgradeHandler(cfg *config.Config, hub *ws.Hub) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := tokenFromProtocolHeader(c.Request.Header.Get("Sec-WebSocket-Protocol"))
		userID := userIDFromToken(token, cfg.JWTSecret)
		if userID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse("invalid_token", "missing or invalid token"))
			return
		}

		conn, err := websocket.Accept(c.Writer, c.Request, &websocket.AcceptOptions{
			Subprotocols: []string{"bearer"},
			// Auth is via the token, and the app is served same-origin behind
			// nginx; skip strict origin checking so the dev proxy also works.
			InsecureSkipVerify: true,
		})
		if err != nil {
			return
		}
		defer conn.CloseNow()

		sub := hub.Register(userID)
		defer sub.Close()

		ctx, cancel := context.WithCancel(c.Request.Context())
		defer cancel()

		// Detect client disconnect: a failed read cancels the context.
		go func() {
			for {
				if _, _, err := conn.Read(ctx); err != nil {
					cancel()
					return
				}
			}
		}()

		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-sub.C:
				if err := conn.Write(ctx, websocket.MessageText, msg); err != nil {
					return
				}
			}
		}
	}
}

// tokenFromProtocolHeader extracts the JWT from a "bearer, <token>" subprotocol
// header value.
func tokenFromProtocolHeader(header string) string {
	for _, part := range strings.Split(header, ",") {
		p := strings.TrimSpace(part)
		if p != "" && p != "bearer" {
			return p
		}
	}
	return ""
}

// userIDFromToken validates a JWT and returns its subject, or "" if invalid.
func userIDFromToken(token, secret string) string {
	if token == "" {
		return ""
	}
	parsed, err := jwtParse(token, secret)
	if err != nil || !parsed.Valid {
		return ""
	}
	claims, ok := parsed.Claims.(jwtMap)
	if !ok {
		return ""
	}
	sub, _ := claims["sub"].(string)
	return sub
}
