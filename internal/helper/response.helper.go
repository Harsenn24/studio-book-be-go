package helper

import "github.com/gin-gonic/gin"

type WebResponse[T any] struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Data    T      `json:"data,omitempty"` 
}

func SendSuccess(c *gin.Context, status int, message string, data any) {
	obj := WebResponse[any]{
        Status:  status,
        Message: message,
        Data:    data,
    }

	c.JSON(status, obj)
}

func SendError(c *gin.Context, status int, message string) {
	obj := WebResponse[any]{
        Status:  status,
        Message: message,
        Data:    nil,
    }

	c.JSON(status, obj)
}