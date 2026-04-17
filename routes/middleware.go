package routes

import (
	"time"

	"echo/common"
	"echo/config"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/rs/zerolog/log"
)

var conf *config.Configurations

func init() {
	log.Debug().Msg("Initializing middleware")
	conf = config.GetConfig()
}

var resInvalidClient = common.ErrResponse{
	Code:    "M001",
	Message: "Request failed. Invalid client id",
}

func getLimiter() func(c *fiber.Ctx) error {

	f := limiter.New(limiter.Config{
		Max:               conf.Server.RateLimit,
		Expiration:        time.Duration(conf.Server.ExpSecs) * time.Second,
		LimiterMiddleware: limiter.SlidingWindow{},
		KeyGenerator: func(c *fiber.Ctx) string {
			bodyStr := string(c.Body())
			ip := c.IP()
			key := bodyStr + ip
			log.Debug().Str("Limiter key", key).Send()
			return key
		},
	})
	return f
}

func logXff(c *fiber.Ctx) error {
	log.Debug().Str("X-FORWARDED-FOR", c.Get("X-Forwarded-For")).Send()
	return c.Next()
}

func requestId(c *fiber.Ctx) error {
	resp := c.Next()
	//log.Debug().Msg("In requsestId middleware")
	//c.Response().Header.Add("Test", "test")
	//GlobalStr += GlobalStr
	return resp
}
