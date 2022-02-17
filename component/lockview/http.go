package lockview

import (
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func trxHandler(c *gin.Context) {
	startStr := c.DefaultQuery("start", "0")
	s := strconv.FormatUint(math.MaxUint64, 10)
	endStr := c.DefaultQuery("end", s)
	start, err := strconv.ParseUint(startStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}
	end, err := strconv.ParseUint(endStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}
	items := store.Select(start, end)
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data":   items,
	})
}

func HTTPService(g *gin.RouterGroup) {
	g.GET("/v1/trx", trxHandler)
}
