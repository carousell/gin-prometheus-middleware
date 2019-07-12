# gin-prometheus-middleware
Go Gin Prometheus middleware to export metrics 

## Example 

    import (
    "github.com/701search/gin-prometheus-middleware"
    "github.com/gin-gonic/gin"
    )

    func main(){
        r := gin.New()
        
        p := gpmiddleware.NewPrometheus("gin")
        p.Use(r)
        
        r.GET("/", func(c *gin.Context) {
		        c.JSON(200, "Hello world! visit /metrics for metrics")
	      })

        r.Run(":37321")
    }
