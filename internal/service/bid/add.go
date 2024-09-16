package bid

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	"net/http"
)

func (s *Service) Add(db *sql.DB, ctx *gin.Context) {
	ctx.Header("Content-Type", "application/json")

	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}

	var bid Bid
	if err = ctx.ShouldBindJSON(&bid); err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"reason": "Invalid request data"})
		return
	}

	bid, ok := insertBid(tx, ctx, bid)
	if !ok {
		return
	}

	if !insertBidDiff(tx, ctx, bid) {
		return
	}

	if err = tx.Commit(); err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}

	returningTender := Bid{
		Id:          bid.Id,
		Name:        bid.Name,
		Description: bid.Description,
		Status:      BidStatusCreated,
		TenderId:    bid.TenderId,
		AuthorType:  bid.AuthorType,
		AuthorId:    bid.AuthorId,
		Version:     1,
		CreatedAt:   bid.CreatedAt,
	}

	ctx.IndentedJSON(http.StatusCreated, returningTender)
}
