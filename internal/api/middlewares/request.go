package middlewares

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"subscription-aggregator-service/internal/utils/request"
)

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := strings.TrimSpace(c.GetHeader(request.HeaderName))
		if id == "" {
			id = uuid.NewString()
		}

		ctx := request.WithContext(c.Request.Context(), id)
		c.Request = c.Request.WithContext(ctx)
		c.Header(request.HeaderName, id)

		c.Next()
	}
}
