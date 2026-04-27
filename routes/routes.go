package routes

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"echo/common"
	"echo/config"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	log "github.com/rs/zerolog/log"
)

const (
	echoDelayHeader  = "X-Echo-Delay"
	echoStatusHeader = "X-Echo-Status"
	echoCloseHeader  = "X-Echo-Close"
)

var GlobalStr = ""

func init() {
	count := 1024 * 8
	for i := 0; i < count; i++ {
		GlobalStr += "1"
	}
}

var resourceNotFound = common.ErrResponse{
	Code:    "R001",
	Message: "Resource not found",
}

func Start() {

	log.Info().Msg("Starting server")
	configs := config.GetConfig()

	fiberConfig := fiber.Config{
		Prefork:               configs.Server.PreFork,
		CaseSensitive:         true,
		StrictRouting:         true,
		ServerHeader:          "boddah",
		AppName:               *configs.Service,
		DisableStartupMessage: false,
		ReduceMemoryUsage:     true,
	}

	fiberApp := fiber.New(fiberConfig)
	reqLogger := logger.New(logger.Config{
		Format: "[${red}${time}] - ${cyan}${ip}:${port} ${status} - ${method} ${path} ${bytesSent} ${latency}\n",
	})

	fiberApp.Use(recover.New())
	fiberApp.Use(requestId)
	//fiberApp.Use(pprof.New())

	fiberApp.Use(compress.New(compress.Config{
		Level: compress.LevelDefault,
	}))

	fiberApp.All("/*", func(c *fiber.Ctx) error {

		log.Trace().Interface("request-context", c.Context()).Send()

		responseMap := make(map[string]any)
		responseMap["path"] = c.Path()
		responseMap["method"] = c.Method()
		responseMap["params"] = c.AllParams()
		responseMap["query"] = c.Queries()
		cookies := make(map[string]string)
		for k, v := range c.Request().Header.Cookies() {
			cookies[string(k)] = string(v)
		}
		responseMap["cookies"] = cookies
		//contentType := c.Request().Header.ContentType()

		if form, filesErr := c.MultipartForm(); filesErr != nil {
			log.Error().AnErr("Error gettinng files", filesErr).Send()
		} else {
			responseMap["files"] = form
			files := form.File["documents"]
			log.Trace().Interface("files", files).Send()

		}

		if c.Method() != "GET" {
			var body any
			if err := json.Unmarshal(c.Body(), &body); err != nil {
				responseMap["body"] = map[string]any{"error": err.Error()}
			} else {
				responseMap["body"] = body
			}
		}

		headers := c.GetReqHeaders()

		responseMap["headers"] = headers
		responseMap["hostname"] = c.Hostname()
		log.Trace().Interface("request-details", responseMap).Send()

		if delay, ok := getHeaderValue(headers, echoDelayHeader); ok && delay != "" {
			if d, err := strconv.Atoi(delay); err != nil {
				log.Error().Str("header", echoDelayHeader).AnErr("Unable to parse delay string to int", err).Send()
			} else {
				log.Debug().Int("Delay", d).Send()
				time.Sleep(time.Duration(d) * time.Second)
			}
		}

		statusCode := fiber.StatusOK
		if statusHeader, ok := getHeaderValue(headers, echoStatusHeader); ok {
			statusCode = resolveStatusCode(statusHeader)
		}

		responseBody, err := json.Marshal(responseMap)
		if err != nil {
			return err
		}

		if closeAfter, ok := getPositiveHeaderInt(headers, echoCloseHeader); ok && closeAfter < int64(len(responseBody)) {
			sendAbruptlyClosedResponse(c, statusCode, responseBody, closeAfter)
			return nil
		}

		return c.Status(statusCode).JSON(responseMap)
	})

	fiberApp.Use(reqLogger)

	err := fiberApp.Listen(":" + fmt.Sprint(configs.Server.Port))
	if err == nil {
		log.Error().AnErr("Error starting server", err)
	}
}

func getHeaderValue(headers map[string][]string, key string) (string, bool) {
	for headerKey, values := range headers {
		if !strings.EqualFold(headerKey, key) {
			continue
		}
		if len(values) == 0 {
			return "", true
		}
		return strings.TrimSpace(values[0]), true
	}

	return "", false
}

func resolveStatusCode(value string) int {
	statusCode, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || http.StatusText(statusCode) == "" {
		return fiber.StatusConflict
	}

	return statusCode
}

func getPositiveHeaderInt(headers map[string][]string, key string) (int64, bool) {
	value, ok := getHeaderValue(headers, key)
	if !ok || value == "" {
		return 0, false
	}

	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		log.Error().Str("header", key).AnErr("Unable to parse header value to int", err).Send()
		return 0, false
	}

	if parsed <= 0 {
		return 0, false
	}

	return parsed, true
}

func sendAbruptlyClosedResponse(c *fiber.Ctx, statusCode int, responseBody []byte, closeAfter int64) {
	c.Context().HijackSetNoResponse(true)
	c.Context().Hijack(func(conn net.Conn) {
		defer conn.Close()

		_, _ = fmt.Fprintf(
			conn,
			"HTTP/1.1 %d %s\r\nContent-Type: application/json; charset=utf-8\r\nContent-Length: %d\r\nConnection: close\r\n\r\n",
			statusCode,
			http.StatusText(statusCode),
			len(responseBody),
		)

		limit := len(responseBody)
		if closeAfter < int64(limit) {
			limit = int(closeAfter)
		}

		if limit > 0 {
			_, _ = conn.Write(responseBody[:limit])
		}
	})
}
