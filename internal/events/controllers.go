package events

import (
	"avana/internal/config"
	"avana/internal/utils"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)


func CreateEvent(c *gin.Context) {
	// get the user id
	userId, err := getUserId(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": utils.ValidateTokenError,
		})
		return
	}
	
	// bind the request data
	var eventSchema CreateEventSchema
	if c.Bind(&eventSchema) != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": utils.ReadRequestError,
		})
		return
	}

	// validate all dates
	eventDate, err := utils.ValidateDate(eventSchema.EventDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": utils.ReadRequestError,
		})
		return
	}

	regExpDate, err := utils.ValidateDate(eventSchema.RegistrationExpirationDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": utils.ReadRequestError,
		})
		return
	}

	// create the object
	event := Event{
		Name: eventSchema.Name,
		Organiser: eventSchema.Organiser,
		Location: eventSchema.Location,
		IsPaidEvent: eventSchema.IsPaidEvent,
		Description: eventSchema.Description,
		IsLimited: eventSchema.IsLimitedEvent,
		MaxUnitReservation: eventSchema.MaxUnitReservation,
		EventDate: eventDate,
		RegistrationExpirationDate: regExpDate,
		UserID: userId,

	}

	// start the saving transaction
	tx := config.DB.Begin()
	
	result := tx.Create(&event)
	if result.Error != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": utils.CreateRecordError + "1",
		})
		return
	}

	// save all the tickets
	if len(eventSchema.Tickets) > 0{
		for _, ticketSchema := range eventSchema.Tickets {
			if err := createTicket(ticketSchema,event.ID,tx); err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{
				"message": utils.CreateRecordError+ "2",
			})
			return
			}
		}
	} else {
		ticket := TicketSchema{
			Name: "Regular",
			Price: 0,
			TotalAvailable: 0,
			SingleLimit: eventSchema.MaxUnitReservation,
			ExpiryTime: eventSchema.RegistrationExpirationDate,
		}

		if eventSchema.IsLimitedEvent {
			ticket.TotalAvailable  = eventSchema.TotalTicketLimit
		}

		if err := createTicket(ticket,event.ID,tx); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
			"message": utils.CreateRecordError+ "3",
		})
		return
		}

	}

	tx.Commit()

	// return success 
	c.JSON(http.StatusOK, gin.H{
		"message": utils.CreateRecordSuccess,
	})
}

func UpdateEvent(c *gin.Context) {
	// get the user id
	userId, err := getUserId(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": utils.ValidateTokenError,
		})
		return
	}

	// get the update object
	var updateSchema UpdateEventSchema
	if err := c.ShouldBind(&updateSchema); err != nil {
		c.JSON(http.StatusBadRequest,gin.H{
			"message": utils.ReadRequestError,
		})
		return
	}

	// get the id
	eventIdStr := c.Param("id")
	eventId, err := strconv.Atoi(eventIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest,gin.H{
			"message": utils.ReadRequestError,
		})
		return
	}

	// get the object from db
	var event  Event
	err = config.DB.First(&event, eventId).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError,gin.H{
			"message": utils.DatabaseCallError,
		})
		return
	}

	// check the permission 
	if err = canOperate(userId,event.UserID); err != nil {
		c.JSON(http.StatusUnauthorized,gin.H{
			"message": utils.IncorrecPermission,
		})
		return
	}


	//  use the update data to populate the model
	updateData, err := getEventUpdateData(updateSchema,event)
	if err != nil {
		c.JSON(http.StatusBadRequest,gin.H{
			"message": utils.ReadRequestError,
		})
		return
	}

	if err = config.DB.Save(&updateData).Error; err != nil {
		c.JSON(http.StatusInternalServerError,gin.H{
			"message": utils.UpdateRecordError,
		})
		return
	}

	//return success 
	c.JSON(http.StatusOK,gin.H{
		"message": utils.UpdateRecordSuccess,
	})
}


func GetAllEvent(c *gin.Context) {
	var events []Event
	if err := config.DB.Order("created_at DESC").Find(&events).Error ; err != nil {
		c.JSON(http.StatusInternalServerError,gin.H{
			"message": utils.DatabaseCallError,
		})
		return
	}

	c.JSON(http.StatusOK,gin.H{
		"event":events,
	})
}

func GetEventByID(c *gin.Context) {
	// get the event by id 
	eventIdStr := c.Param("id")
	eventId, err := strconv.Atoi(eventIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest,gin.H{
			"message": utils.ReadRequestError,
		})
		return
	}

	// query the database
	var event Event
	if err = config.DB.First(&event,eventId).Error; err != nil {
		c.JSON(http.StatusInternalServerError,gin.H{
			"message": utils.DatabaseCallError,
		})
		return
	}

	c.JSON(http.StatusOK,gin.H{
		"event":event,
	})
}

func GetAllTickets(c *gin.Context) {
	// get the ticket id 
	eventIdStr := c.Param("id")
	eventId, err := strconv.Atoi(eventIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest,gin.H{
			"message": utils.ReadRequestError,
		})
		return
	}

	// query the database
	var tickets []Ticket
	if err := config.DB.Where("event_id = ?",eventId).Order("price").Find(&tickets).Error; err != nil {
		c.JSON(http.StatusInternalServerError,gin.H{
			"message": utils.DatabaseCallError,
		})
		return
	}

	// return the tickets
	c.JSON(http.StatusOK,gin.H{
		"tickets":tickets,
	})
}

func GetTicketById(c *gin.Context) {
	// get the ticket id 
	ticketIdStr := c.Param("id")
	ticketId, err := strconv.Atoi(ticketIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest,gin.H{
			"message": utils.ReadRequestError,
		})
		return
	}

	// query the database
	var ticket Ticket
	if err := config.DB.First(&ticket,ticketId).Error; err != nil {
		c.JSON(http.StatusInternalServerError,gin.H{
			"message": utils.DatabaseCallError,
		})
		return
	}

	// return the ticket
	c.JSON(http.StatusOK,gin.H{
		"ticket":ticket,
	})
}

func AddTicket(c *gin.Context) {
	// bind the ticket object
	var ticketSchema TicketSchema
	if c.Bind(&ticketSchema) !=nil {
		c.JSON(http.StatusBadRequest,gin.H{
			"message": utils.ReadRequestError,
		})
		return
	}

	// get the user id 
	userId, err := getUserId(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": utils.ValidateTokenError,
		})
		return
	}

	//get the event id
	eventIdStr := c.Param("id")
	eventId, err := strconv.Atoi(eventIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest,gin.H{
			"message": utils.ReadRequestError,
		})
		return
	}

	// get the event object
	var event Event
	if err = config.DB.First(&event,eventId).Error; err != nil {
		c.JSON(http.StatusInternalServerError,gin.H{
			"message": utils.DatabaseCallError,
		})
		return
	}

	//can operate 
	if err = canOperate(userId,event.UserID); err != nil {
		c.JSON(http.StatusUnauthorized,gin.H{
			"message": utils.IncorrecPermission,
		})
		return
	}

	// change the time
	expTime, err := utils.ValidateDate(ticketSchema.ExpiryTime)
	if err != nil {
		c.JSON(http.StatusBadRequest,gin.H{
			"message": utils.ReadRequestError,
		})
		return
	}

	if expTime.After(event.EventDate) {
		c.JSON(http.StatusBadRequest,gin.H{
			"message": utils.TicketTimeError,
		})
		return
	}

	// verify the price
	if !event.IsPaidEvent && ticketSchema.Price > 0{
		c.JSON(http.StatusBadRequest,gin.H{
			"message":utils.PriceError,
		})
		return
	}

	// create the ticket object 
	ticket := Ticket{
		Name: ticketSchema.Name,
		Price: ticketSchema.Price,
		TotalAvailable: ticketSchema.TotalAvailable,
		SingleLimit: ticketSchema.SingleLimit,
		ExpiryTime: expTime,
		EventID: event.ID,
	}

	// save the model
	if err = config.DB.Create(&ticket).Error; err != nil {
		c.JSON(http.StatusInternalServerError,gin.H{
			"message": utils.CreateRecordError,
		})
		return
	}

	// return success 
	c.JSON(http.StatusOK,gin.H{
		"message":utils.CreateRecordSuccess,
	})
}

func UpdateTicket(c *gin.Context) {
	// bind the data
	var updateSchema UpdateTicketSchema
	if err := c.ShouldBind(&updateSchema); err != nil {
		c.JSON(http.StatusBadRequest,gin.H{
			"message": utils.ReadRequestError,
		})
		return
	}

	// get the user id 
	userId, err := getUserId(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": utils.ValidateTokenError,
		})
		return
	}

	// get the ticket id
	ticketIdStr := c.Param("id")
	ticketId, err := strconv.Atoi(ticketIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest,gin.H{
			"message": utils.ReadRequestError,
		})
		return
	}

	//get the ticket
	var ticket Ticket
	if err = config.DB.First(&ticket,ticketId).Error; err != nil {
		c.JSON(http.StatusInternalServerError,gin.H{
			"message": utils.DatabaseCallError,
		})
		return
	}

	// get the event id
	var event Event
	if err = config.DB.First(&event,ticket.EventID).Error; err != nil {
							c.JSON(http.StatusInternalServerError,gin.H{
								"message": utils.DatabaseCallError,
							})
							return
						}

	//compare the event id
	if err = canOperate(userId, event.UserID); err != nil {
		c.JSON(http.StatusUnauthorized,gin.H{
			"message": utils.IncorrecPermission,
		})
		return
	}

	// build the new ticket
	newTicket, err := getTicketUpdateData(updateSchema,ticket,event.EventDate,event.IsPaidEvent)
	if err != nil {
		c.JSON(http.StatusBadRequest,gin.H{
			"message": utils.ReadRequestError,
		})
		return
	}

	// save the model
	if err = config.DB.Save(&newTicket).Error; err != nil {
		c.JSON(http.StatusInternalServerError,gin.H{
			"message": utils.UpdateRecordError,
		})
		return
	}

	//return success
	c.JSON(http.StatusOK,gin.H{
		"message": utils.UpdateRecordSuccess,
	})

}

func GetMyEvents(c *gin.Context) {
	// get the user id
	userId, err := getUserId(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": utils.ValidateTokenError,
		})
		return
	}

	// query the db
	var events []Event
	if err := config.DB.Where("user_id = ?", userId).Order("created_at DESC").Find(&events).Error ; err != nil {
		c.JSON(http.StatusInternalServerError,gin.H{
			"message": utils.DatabaseCallError,
		})
		return
	}

	c.JSON(http.StatusOK,gin.H{
		"event":events,
	})
}

func DeleteTicket(c *gin.Context) {
	// get the ticket id 
	ticketIdStr := c.Param("id")
	ticketId, err := strconv.Atoi(ticketIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest,gin.H{
			"message": utils.ReadRequestError,
		})
		return
	}

	// get the user id
	userId, err := getUserId(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": utils.ValidateTokenError,
		})
		return
	}


	// get the event id
	var eventUserId uint
	if err = config.DB.Table("tickets").
				Select("events.user_id").
				Joins("JOIN events ON tickets.event_id = events.id").
				Where("tickets.id = ?",ticketId).
				Find(&eventUserId).Error;
	 err != nil {
			c.JSON(http.StatusInternalServerError,gin.H{
				"message": utils.DatabaseCallError,
			})
			return
				}

	//compare the event id
	if err = canOperate(userId, eventUserId); err != nil {
		c.JSON(http.StatusUnauthorized,gin.H{
			"message": utils.IncorrecPermission,
		})
		return
	}

	// // delete the ticket 
	if err = config.DB.Delete(&Ticket{},ticketId).Error; err != nil {
		c.JSON(http.StatusInternalServerError,gin.H{
			"message": utils.DeleteRecordError,
		})
		return
	}

	//return success message
	c.JSON(http.StatusOK, gin.H{
		"message": utils.DeleteRecordSuccess,
	})
}

func DeleteEvent(c *gin.Context) {
	// get the event id 
	eventIdStr := c.Param("id")
	eventId, err := strconv.Atoi(eventIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest,gin.H{
			"message": utils.ReadRequestError,
		})
		return
	}

	// get the user id
	userId, err := getUserId(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": utils.ValidateTokenError,
		})
		return
	}

	// verify the userId
	var event Event
	if err = config.DB.First(&event, eventId).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": utils.DatabaseCallError,
		})
		return
	}

	// verify if the user has permission
	if err = canOperate(userId, event.UserID); err != nil {
		c.JSON(http.StatusUnauthorized,gin.H{
			"message": utils.IncorrecPermission,
		})
		return
	}

	// delete all related tickets
	if err = config.DB.Where("event_id = ?",event.ID).Delete(&Ticket{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": utils.DeleteRecordError,
		})
		return
	}
	// delete all events 
	if err = config.DB.Delete(&Event{},event.ID).Error;err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": utils.DeleteRecordError,
		})
		return
	}

	//return success message
	c.JSON(http.StatusOK, gin.H{
			"message": utils.DeleteRecordSuccess,
		})
}

func BuyTicket(c * gin.Context) {
	// get the user id
	userId, err := getUserId(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": utils.ValidateTokenError,
		})
		return
	}

	// bind the request
	var ticketSchema BuyTicketScema
	if err = c.Bind(&ticketSchema); err != nil {
		c.JSON(http.StatusBadRequest,gin.H{
			"message": utils.ReadRequestError,
		})
		return
	}

	// get the ticket details
	ticketIdStr := c.Param("id")
	ticketId, err := strconv.Atoi(ticketIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest,gin.H{
			"message": utils.ReadRequestError,
		})
		return
	}

	var ticket Ticket
	if err = config.DB.First(&ticket,ticketId).Error; err != nil {
		c.JSON(http.StatusInternalServerError,gin.H{
			"message": utils.DatabaseCallError,
		})
		return
	}

	// verify the ticket units 
	if ticketSchema.Units > ticket.SingleLimit {
		c.JSON(http.StatusBadRequest,gin.H{
			"message": utils.TicketAmountError,
		})
		return
	}

	// verify if the ticket is still valid
	if time.Now().After(ticket.ExpiryTime) {
		c.JSON(http.StatusBadRequest,gin.H{
			"message": utils.TicketExpiredError,
		})
		return
	}
	
	var attendeeID uint
	if err = config.DB.Table("attendees").Select("id").
					Where("user_id = ? AND ticket_id = ?",userId,ticketId).Scan(&attendeeID).Error;
					err != nil {
						c.JSON(http.StatusInternalServerError,gin.H{
							"message": utils.DatabaseCallError,
						})
						return
					}
	if attendeeID != 0 {
		c.JSON(http.StatusInternalServerError,gin.H{
			"message": utils.ExistingDataError,
		})
		return
	}
	
	attendee := Attendee{
			UserID: userId,
			Units: ticketSchema.Units,
			TicketID: ticket.ID,
		}

	if ticket.Price == 0 {		// free ticket case
		// save the atendee model
		if err = config.DB.Create(&attendee).Error; err != nil {
			c.JSON(http.StatusInternalServerError,gin.H{
				"message": utils.CreateRecordError,
			})
			return
		}

	} else if ticket.Price >0 {
		// case for paid ticket
	} else {
		// bad request
		c.JSON(http.StatusBadRequest,gin.H{
			"message":utils.ReadRequestError,
		})
		return
	}

	c.JSON(http.StatusOK,gin.H{
		"message": utils.CreateRecordSuccess,
	})
}

func GetTotalAttendees(c *gin.Context) {
	// get the event id 
	eventIdStr := c.Param("id")
	eventId, err := strconv.Atoi(eventIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest,gin.H{
			"message": utils.ReadRequestError,
		})
		return
	}

	// get the user id 
	userId, err := getUserId(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": utils.ValidateTokenError,
		})
		return
	}

	// get the event id
	var event Event
	if err = config.DB.First(&event,eventId).Error; err != nil {
							c.JSON(http.StatusInternalServerError,gin.H{
								"message": utils.DatabaseCallError,
							})
							return
						}

	//compare the event id
	if err = canOperate(userId, event.UserID); err != nil {
		c.JSON(http.StatusUnauthorized,gin.H{
			"message": utils.IncorrecPermission,
		})
		return
	}

	// get all tickets using the id 
	var attendees []GetAllAttendees
	query := config.DB.Table("attendees").
					Joins("LEFT JOIN tickets ON tickets.id = attendees.ticket_id").
					Joins("LEFT JOIN events ON events.id = tickets.event_id").
					Joins("LEFT JOIN users ON users.id = attendees.user_id").
					Select("users.email AS email,tickets.name AS ticket_type, attendees.units AS amount").
					Where("events.id = ?", eventId).Scan(&attendees)
	
	
	 if err = query.Error; err != nil {
		c.JSON(http.StatusInternalServerError,gin.H{
			"message": utils.DatabaseCallError,
		})
		return
	 }
	

	 // return success
	c.JSON(http.StatusOK,gin.H{
		"attendees": attendees,
		"attendeeCount":query.RowsAffected,
	})
}

func GetAllReviews(c *gin.Context) {
	// get the event id 
	eventIdStr := c.Param("id")
	eventId, err := strconv.Atoi(eventIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest,gin.H{
			"message": utils.ReadRequestError,
		})
		return
	}
	// verify if the event has held
	var eventHeld time.Time 
	var reviews []GetAllReviewSchema
	err = config.DB.Table("events").Select("event_date").Where("id = ?",eventId).Scan(&eventHeld).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError,gin.H{
			"message": utils.DatabaseCallError,
		})
		return
	}

	// return empty string for 
	if !time.Now().After(eventHeld) {
		c.JSON(http.StatusOK,gin.H{
			"reviews": reviews,
			"message": utils.EventHeldError,
		})
		return
	}

	// use the string
	err = config.DB.Table("attendees").
					Joins("LEFT JOIN tickets ON tickets.id = attendees.ticket_id").
					Joins("LEFT JOIN events ON events.id = tickets.event_id").
					Joins("LEFT JOIN users ON users.id = attendees.user_id").
					Select("users.first_name, users.last_name, attendees.rating, attendees.review").
					Where("events.id = ?", eventId).
					Scan(&reviews).
					Error
	if err != nil {
		c.JSON(http.StatusInternalServerError,gin.H{
			"message": utils.DatabaseCallError,
		})
		return
	}

	c.JSON(http.StatusOK,gin.H{
		"reviews": reviews,
	})
	
}

func VerifyAttendance(c * gin.Context) {

}

func ReviewEvent(c * gin.Context) {
	
}




// internal functions
func createTicket(ticket TicketSchema, eventId uint, tx *gorm.DB) error{
	//validate the date 
	date, err := utils.ValidateDate(ticket.ExpiryTime)
	if err != nil {
		return errors.New("incorrect date time format")
	}
	// create the ticket model
	ticketModel := Ticket{
		Name: ticket.Name,
		Price: ticket.Price,
		TotalAvailable: ticket.TotalAvailable,
		SingleLimit: ticket.SingleLimit,
		ExpiryTime: date,
		EventID: eventId,
	}

	if err = tx.Create(&ticketModel).Error; err != nil {
		return err
	}

	return nil
}

func getUserId(c *gin.Context) (uint, error){
	//get the auth token 
	userIdStr, exist := c.Get("userID")
	if !exist {
		return 0, errors.New("wrong user Id")
	}

	userId, ok := userIdStr.(uint)
	if !ok {
		return 0, errors.New("wrong user Id")
	}
	return userId, nil
}

func canOperate(userId, jobUserId uint) error{
	if userId != jobUserId {
		return errors.New("incorrect permission")
	}
	return nil
}

func getEventUpdateData(updateData UpdateEventSchema, event Event) (Event,error) {
	if updateData.Name != "" {
		event.Name = updateData.Name
	}

	if updateData.Location != "" {
		event.Location = updateData.Location
	}

	if updateData.Organiser != "" {
		event.Organiser = updateData.Organiser
	}

	if updateData.IsPaidEvent != nil {
		isPaidEvent := utils.BoolValue(updateData.IsPaidEvent, false)
		event.IsPaidEvent = isPaidEvent
	}

	if updateData.IsLimitedEvent != nil{
		isLimitedEvent := utils.BoolValue(updateData.IsLimitedEvent, false)
		event.IsLimited = isLimitedEvent
	}

	if updateData.Description != "" {
		event.Description = updateData.Description
	}

	if updateData.MaxUnitReservation != nil {
		maxUnitReservation := utils.UintValue(updateData.MaxUnitReservation,1)
		event.MaxUnitReservation = maxUnitReservation
	}

	if updateData.EventDate != "" {
		eventDate, err := utils.ValidateDate(updateData.EventDate)
		if err != nil {
			return Event{}, errors.New("invalid date")
		}
		event.EventDate = eventDate
	}

	if updateData.RegistrationExpirationDate != "" {
		regDate, err := utils.ValidateDate(updateData.RegistrationExpirationDate)
		if err != nil {
			return Event{}, errors.New("invalid date")
		}
		event.EventDate = regDate
	}

	return event, nil 
}

func getTicketUpdateData(updateData UpdateTicketSchema, ticket Ticket, eventDate time.Time, isPaid bool) (Ticket, error) {
	if updateData.Name != "" {
		ticket.Name = updateData.Name
	}

	if updateData.Price != nil {
		actualPrice := utils.FloatValue(updateData.Price, 0 )
		if !isPaid && actualPrice > 0 {
			return Ticket{}, errors.New("the event is free")
		}
		ticket.Price = actualPrice
	}

	if updateData.TotalAvailable != nil {
		actualAvailable := utils.UintValue(updateData.TotalAvailable,0)
		ticket.TotalAvailable = actualAvailable
	}

	if updateData.SingleLimit != nil {
		actualLimit := utils.UintValue(updateData.SingleLimit,0)
		ticket.TotalAvailable = actualLimit
	}

	if updateData.ExpiryTime != "" {
		expTime, err := utils.ValidateDate(updateData.ExpiryTime)
		if err != nil {
			return Ticket{}, errors.New("wrong time")
		}
		if expTime.After(eventDate) {
			return Ticket{},errors.New("wrong time")
		}
	}

	return ticket, nil
}




