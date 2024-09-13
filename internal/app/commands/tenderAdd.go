package commands

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

func (cmd *Commander) AddTender(ctx *gin.Context) {
	defer func() {
		if panicValue := recover(); panicValue != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"reason": fmt.Sprintf("Recovered from panic: %v", panicValue)})
			return
		}
	}()

	cmd.tenderService.Add(cmd.db, ctx)
}
