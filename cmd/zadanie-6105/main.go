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

	tenderGroup := router.Group("/api/tenders")
	bidGroup := router.Group("/api/bids")

	router.GET("/api/ping", commander.Ping)

	tenderGroup.GET("", commander.ListAllTenders)
	tenderGroup.GET("/my", commander.ListMyTenders)
	tenderGroup.POST("/new", commander.AddTender)
	tenderGroup.GET("/:tenderId/status", commander.TenderStatus)
	tenderGroup.PUT("/:tenderId/status", commander.PutTenderStatus)
	tenderGroup.PATCH("/:tenderId/edit", commander.PatchTender)
	tenderGroup.PUT("/:tenderId/rollback/:version", commander.TenderRollback)

	bidGroup.POST("/new", commander.AddBid)
	bidGroup.GET("/my", commander.ListMy)
	bidGroup.GET("/tender/:tenderId/list", commander.TenderIdList)
	bidGroup.GET("/:bidId/status", commander.BidStatus)
	bidGroup.PUT("/:bidId/status", commander.PutBidStatus)
	bidGroup.PATCH("/:bidId/edit", commander.PatchBid)
	bidGroup.PUT("/bids/:bidId/rollback/:version", commander.BidRollback)

	err := router.Run(serverAddress)
	if err != nil {
		log.Fatal("Error starting server:", err)
	}
}
