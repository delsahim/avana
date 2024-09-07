package main

import (
	"avana/internal/config"
	"avana/internal/events"
	"avana/internal/middlewares"
	"avana/internal/users"

	"github.com/gin-gonic/gin"
)

func init() {
	config.ConnectToDb()
}

func main() {
	r := gin.Default()

	usergroup := r.Group("/user")
	usergroup.POST("/create",users.CreateUser)
	usergroup.POST("/login",users.Login)
	usergroup.POST("/otp",users.GetOtp)
	usergroup.POST("/otp/verify",users.VerifyOtp)
	usergroup.POST("/password/change",users.ChangePassword)


	eventgroup := r.Group("/event")
	eventgroup.POST("/create",middlewares.RequireAuth,events.CreateEvent)
	eventgroup.GET("/:id",events.GetEventByID)
	eventgroup.GET("/all",events.GetAllEvent)
	eventgroup.GET("/:id/ticket/all",events.GetAllTickets)
	eventgroup.GET("/ticket/:id",events.GetTicketById)
	eventgroup.PATCH("/update/:id",middlewares.RequireAuth,events.UpdateEvent)
	eventgroup.POST("/:id/ticket/create",middlewares.RequireAuth,events.AddTicket)
	eventgroup.PATCH("/ticket/:id",middlewares.RequireAuth,events.UpdateTicket)
	eventgroup.DELETE("/ticket/:id",middlewares.RequireAuth,events.DeleteTicket)
	eventgroup.DELETE("/:id",middlewares.RequireAuth, events.DeleteEvent)
	eventgroup.POST("/ticket/:id/buy", middlewares.RequireAuth,events.BuyTicket)
	eventgroup.GET("/:id/attendees",middlewares.RequireAuth,events.GetTotalAttendees)
	eventgroup.GET("/:id/reviews", events.GetAllReviews)
	

	r.Run(":8000")
}