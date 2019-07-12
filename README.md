# gin-prometheus-middleware
Go [Gin](https://github.com/gin-gonic/gin) middleware for Prometheus

Export metrics for request duration and request count

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
