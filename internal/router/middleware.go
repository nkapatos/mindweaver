package router

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// TODO: move them to the package middleware under the router package

// Custom middleware functions can be added here
// Example:
// func ValidateJSON() echo.MiddlewareFunc {
//     return func(next echo.HandlerFunc) echo.HandlerFunc {
//         return func(c echo.Context) error {
//             // Validation logic here
//             return next(c)
//         }
//     }
// }

// RateLimit middleware example
func RateLimit() echo.MiddlewareFunc {
	return middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(20))
}

// CORS middleware
func CORS() echo.MiddlewareFunc {
	return middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.GET, echo.HEAD, echo.PUT, echo.PATCH, echo.POST, echo.DELETE},
	})
}
