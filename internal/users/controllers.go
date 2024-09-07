package users

import (
	"avana/internal/config"
	"avana/internal/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

func CreateUser(c *gin.Context) {
	var userSchema CreateUserSchema
	// bind the schema
	if c.Bind(&userSchema) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": utils.ReadRequestError,
		})
		return
	}

	// check if the user exists
	var userCheck User
	if (config.DB.Where("email = ?", userSchema.Email).
	First(&userCheck).RowsAffected > 0) {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": utils.ExistingDataError,
		})
		return
	}

	// hash the password
	hash, err := bcrypt.GenerateFromPassword([]byte(userSchema.Password),10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": utils.HashingError,
		})
		return
	}

	//create the model instance
	user := User{
		FirstName: userSchema.FirstName,
		LastName: userSchema.LastName,
		Email: userSchema.Email,
		Password: string(hash),

	}

	// save the model
	err = config.DB.Create(&user).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": utils.CreateRecordError,
		})
		return
	}

	// return the success message
	c.JSON(http.StatusCreated,gin.H{
		"message":utils.CreateRecordSuccess,
	})

}

func Login(c *gin.Context) {
	// receive the request body
	var loginSchema UserCredentials
	if c.Bind(&loginSchema) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": utils.ReadRequestError,
		})
		return
	}

	// get the user details
	var user User
	if err := config.DB.Table("users").
					Where("email= ?",loginSchema.Email).
					First(&user).Error; err !=nil {
						c.JSON(http.StatusInternalServerError, gin.H{
							"message": utils.DatabaseCallError,
						})
						return
					}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password),[]byte(loginSchema.Password));
			err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"message": utils.CredentialsError,
				})
				return
				}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256,jwt.MapClaims{
		"sub": user.Email,
		"exp": time.Now().Add(time.Hour * 24 * 30).Unix(),
	})

	tokenString, err := token.SignedString([]byte("chgjfskiuyfgshdigjhv"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": utils.TokenError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":tokenString,
		"type":"Bearer",
		"expiresIn": 86400,
	})
}


func GetOtp(c *gin.Context) {
	// get the request body
	var getOtpSchema UserEmail
	if c.Bind(&getOtpSchema) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message":utils.ReadRequestError,
		})
		return
	}

	// get the user
	var user User
	if err := config.DB.Table("users").
					Where("email= ?",getOtpSchema.Email).
					First(&user).Error; err !=nil {
						c.JSON(http.StatusInternalServerError, gin.H{
							"message": utils.DatabaseCallError,
						})
						return
					}
	
	// generate otp 
	user.Otp = utils.GenerateOtp()
	user.OtpExpires = time.Now().Add(time.Minute*10)

	//save the otp 
	if err := config.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": utils.UpdateRecordError,
		})
		return
	}

	// send out the otp (to be changed to mailing)
	c.JSON(http.StatusOK, gin.H{
		"message": utils.OperationSucess,
		"otp": user.Otp,
	})

}

func VerifyOtp(c *gin.Context) {
	// bind the body 
	var verifyOtpSchema OtpCredentials
	if c.Bind(&verifyOtpSchema) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message":utils.ReadRequestError,
		})
		return
	}

	// fetch the user data
	var user User
	if err := config.DB.Table("users").
					Where("email= ?",verifyOtpSchema.Email).
					First(&user).Error; err !=nil {
						c.JSON(http.StatusInternalServerError, gin.H{
							"message": utils.DatabaseCallError,
						})
						return
					}
	
	// verify and save the user
	user.OtpVerified = true
	if err := config.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": utils.UpdateRecordError,
		})
		return
	}
	
	// send success message
	c.JSON(http.StatusOK, gin.H{
		"message": utils.OperationSucess,
	})

}

func ChangePassword(c *gin.Context) {
	// bind the request data
	var changePasswordSchema UserCredentials
	if c.Bind(&changePasswordSchema) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message":utils.ReadRequestError,
		})
		return
	}

	// get the user
	var user User
	if err := config.DB.Table("users").
					Where("email= ?",changePasswordSchema.Email).
					First(&user).Error; err !=nil {
						c.JSON(http.StatusInternalServerError, gin.H{
							"message": utils.DatabaseCallError,
						})
						return
					}

	// check if the user verification is valid
	if !user.OtpVerified || time.Now().After(user.OtpExpires) {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": utils.ExpiresVerificationError,
		})
		return
	}

	// hash the password
	hash, err := bcrypt.GenerateFromPassword([]byte(changePasswordSchema.Password),10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": utils.HashingError,
		})
		return
	}

	// change the new password and save it
	user.Password = string(hash)
	user.OtpVerified = false
	user.OtpExpires = time.Now()
	user.Otp = ""

	if err := config.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": utils.UpdateRecordError,
		})
		return
	}

	// send success message
	c.JSON(http.StatusOK, gin.H{
		"message": utils.OperationSucess,
	})
}

