package routers

import (
	"net/http"
	"sync"
	"time"

	"github.com/gokul-ms4/sop-manager/config"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"
)

type ipLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.Mutex
}

var limiter = &ipLimiter{
	limiters: make(map[string]*rate.Limiter),
}

func (l *ipLimiter) get(ip string) *rate.Limiter {
	l.mu.Lock()
	defer l.mu.Unlock()
	if _, ok := l.limiters[ip]; !ok {
		l.limiters[ip] = rate.NewLimiter(rate.Every(time.Minute), 1000)
	}
	return l.limiters[ip]
}

func RateLimitMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		ip := c.RealIP()
		if !limiter.get(ip).Allow() {
			return c.JSON(http.StatusTooManyRequests, map[string]interface{}{
				"success": false,
				"error":   "Too many requests. Please slow down.",
			})
		}
		return next(c)
	}
}

func InitRoutes(e *echo.Echo) {
    e.Use(middleware.Recover())
    e.Use(middleware.BodyLimit("100M"))

    e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
        AllowOrigins: []string{"http://localhost:5173", "*"},
        AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
        AllowHeaders: []string{"Content-Type", "Authorization"},
    }))

    e.Use(RateLimitMiddleware)

    // Static files inside InitRoutes so CORS covers them
    e.Static("/uploads", "uploads")

    AuthRoutes(e)

    protected := e.Group("/api/v1")
    protected.Use(config.JWTMiddleware())


    SopRoutes(protected)
    // ChatRoutes(protected)
    // GroupRoutes(protected)
    // NotificationRoutes(protected)
}