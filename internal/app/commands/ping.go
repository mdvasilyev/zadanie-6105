package commands

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

func (cmd *Commander) Ping(ctx *gin.Context) {
	ctx.Header("Content-Type", "text/plain")
	defer func() {
		if panicValue := recover(); panicValue != nil {
			ctx.String(http.StatusInternalServerError, fmt.Sprintf("Recovered from panic: %v", panicValue))
			return
		}
	}()

	ctx.String(http.StatusOK, "ok")
}
