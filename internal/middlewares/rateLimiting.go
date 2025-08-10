package middlewares

import (
	"github.com/gin-gonic/gin"
	"github.com/ulule/limiter/v3"
	ginLimiter "github.com/ulule/limiter/v3/drivers/middleware/gin"
	memstore "github.com/ulule/limiter/v3/drivers/store/memory"
)

func RateLimiterMiddleware() gin.HandlerFunc {
	rate, _ := limiter.NewRateFromFormatted("50-M")
	store := memstore.NewStore()
	return ginLimiter.NewMiddleware(limiter.New(store, rate))
}
