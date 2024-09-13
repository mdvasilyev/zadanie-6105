package main

import (
	"git.codenrock.com/avito-testirovanie-na-backend-1270/cnrprod1725726738-team-78269/zadanie-6105/internal/app/commands"
	"git.codenrock.com/avito-testirovanie-na-backend-1270/cnrprod1725726738-team-78269/zadanie-6105/internal/app/database"
	"git.codenrock.com/avito-testirovanie-na-backend-1270/cnrprod1725726738-team-78269/zadanie-6105/internal/service/bid"
	"git.codenrock.com/avito-testirovanie-na-backend-1270/cnrprod1725726738-team-78269/zadanie-6105/internal/service/tender"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"log"
	"os"
)

func main() {
	db := database.ConnectDatabase()

	serverAddress := os.Getenv("SERVER_ADDRESS")
	if serverAddress == "" {
		log.Fatal("SERVER_ADDRESS not set")
	}

	tenderService := tender.NewService()
	bidService := bid.NewService()

	commander := commands.NewCommander(db, tenderService, bidService)

	router := gin.Default()

	apiGroup := router.Group("/api")

	apiGroup.GET("/ping", commander.Ping)

	apiGroup.GET("/tenders", commander.ListAllTenders)
	apiGroup.GET("/tenders/my", commander.ListMyTenders)
	apiGroup.POST("/tenders/new", commander.AddTender)
	apiGroup.GET("/tenders/:tenderId/status", commander.TenderStatus)
	apiGroup.PUT("/tenders/:tenderId/status", commander.PutTenderStatus)
	apiGroup.PATCH("/tenders/:tenderId/edit", commander.PatchTender)
	apiGroup.PUT("/tenders/:tenderId/rollback/:version", commander.TenderRollback)

	apiGroup.POST("/bids/new", commander.AddBid)
	apiGroup.GET("/bids/my", commander.ListMy)
	apiGroup.GET("/bids/tender/:tenderId/list", commander.TenderIdList)
	apiGroup.GET("/bids/:bidId/status", commander.BidStatus)
	apiGroup.PUT("/bids/:bidId/status", commander.PutBidStatus)
	apiGroup.PATCH("/bids/:bidId/edit", commander.PatchBid)
	apiGroup.PUT("/bids/:bidId/rollback/:version", commander.BidRollback)

	err := router.Run(serverAddress)
	if err != nil {
		log.Fatal("Error starting server:", err)
	}
}
