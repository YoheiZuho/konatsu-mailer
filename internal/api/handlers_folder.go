package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/yoheizuho/konatsu-mailer/internal/store"
)

// listFoldersHandler returns the user's IMAP mailboxes (discovered during sync).
// INBOX is always present so the sidebar has at least one folder before the
// first sync completes.
func listFoldersHandler(db *store.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("userID")
		folders, err := db.UserFolders(c.Request.Context(), userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, errorResponse("internal_error", "failed to list folders"))
			return
		}
		hasInbox := false
		for _, f := range folders {
			if f.Role == "inbox" || f.Name == "INBOX" {
				hasInbox = true
				break
			}
		}
		if !hasInbox {
			folders = append([]store.Folder{{Name: "INBOX", Role: "inbox"}}, folders...)
		}
		c.JSON(http.StatusOK, gin.H{"items": folders})
	}
}
