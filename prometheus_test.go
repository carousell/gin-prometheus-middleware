package gpmiddleware

import (
	"testing"
	"log"
    "github.com/gin-gonic/gin"
)

func TestNewPrometheus(t *testing.T){
    r := gin.New()
    p := NewPrometheus("gin")
    p.Use(r)
	r.GET("/", routeHandlerFn)
	r.GET("/health", routeHandlerHealthFn)
	log.Println("Listening to http://localhost:37321/metrics")
    r.Run(":37321")

}

func routeHandlerFn(c *gin.Context){
	c.JSON(200, "Hello world! visit /metrics for metrics")
}
func routeHandlerHealthFn(c *gin.Context){
	c.JSON(200, "Hello world! visit /metrics for metrics")
}