package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/yoheizuho/konatsu-mailer/internal/store"
)

type folderResp struct {
	Name   string `json:"name"`
	Role   string `json:"role"`
	Unread int    `json:"unread"`
}

// listFoldersHandler returns the user's IMAP mailboxes (discovered during sync)
// with per-folder unread counts. INBOX is always present so the sidebar has at
// least one folder before the first sync completes. Virtual views' unread
// counts (starred / important) are returned alongside.
func listFoldersHandler(db *store.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("userID")
		ctx := c.Request.Context()

		folders, err := db.UserFolders(ctx, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, errorResponse("internal_error", "failed to list folders"))
			return
		}

		counts, starredUnread, importantUnread, err := db.UnreadCounts(ctx, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, errorResponse("internal_error", "failed to count unread"))
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

		items := make([]folderResp, len(folders))
		for i, f := range folders {
			items[i] = folderResp{Name: f.Name, Role: f.Role, Unread: counts[f.Name]}
		}

		c.JSON(http.StatusOK, gin.H{
			"items":            items,
			"starred_unread":   starredUnread,
			"important_unread": importantUnread,
		})
	}
}
