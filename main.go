package main

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"time"

	"go.uber.org/zap"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

func main() {
	e := echo.New()
	e.Use(traceMiddleware)
	e.HTTPErrorHandler = errorHandler
	e.GET("/standard", standard)
	e.GET("/structured", structured)
	e.GET("/ungrouped", ungrouped)
	e.GET("/grouped", grouped)
	e.GET("/uncolored", uncolored)
	e.GET("/colored", colored)
	e.GET("/unstacktraced", unstacktraced)
	e.GET("/stacktraced", stacktraced)
	log.Fatal(e.Start(":8080"))
}

func standard(c echo.Context) error {
	log.Print("hogehoge")
	return nil
}

func structured(c echo.Context) error {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	logger.Info("ほげほげ",
		zap.String("ふがふが", "fugafuga"),
		zap.String("ぴよぴよ", "piyopiyo"),
	)

	return nil
}

func ungrouped(c echo.Context) error {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	time.Sleep(time.Second * 1)
	logger.Info("hogehoge")
	time.Sleep(time.Second * 1)
	logger.Info("fugafuga")
	time.Sleep(time.Second * 1)
	logger.Info("piyopiyo")

	return nil
}

type ctxKey struct{}

func traceMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		match := regexp.MustCompile(`([a-f\d]+)/([a-f\d]+)`).
			FindAllStringSubmatch(
				c.Request().Header.Get("X-Cloud-Trace-Context"),
				-1,
			)

		trace := match[0][1]

		ctx := context.WithValue(c.Request().Context(), ctxKey{}, trace)
		c.SetRequest(c.Request().WithContext(ctx))

		return next(c)
	}
}

func grouped(c echo.Context) error {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	trace := c.Request().Context().Value(ctxKey{}).(string)

	traceField := zap.String(
		"logging.googleapis.com/trace",
		fmt.Sprintf("projects/%s/traces/%s", "exs-development", trace),
	)
	time.Sleep(time.Second * 1)
	logger.Info("hogehoge", traceField)
	time.Sleep(time.Second * 1)
	logger.Info("fugafuga", traceField)
	time.Sleep(time.Second * 1)
	logger.Info("piyopiyo", traceField)

	return nil
}

func uncolored(c echo.Context) error {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	logger.Info("hogehoge")
	logger.Error("fugafuga")
	logger.Fatal("piyopiyo")

	return nil
}

func colored(c echo.Context) error {
	config := zap.NewProductionConfig()
	config.EncoderConfig.LevelKey = "severity"

	logger, _ := config.Build()
	defer logger.Sync()

	logger.Info("hogehoge")
	logger.Error("fugafuga")
	logger.Fatal("piyopiyo")

	return nil
}

func unstacktraced(c echo.Context) error {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	err := fmt.Errorf("error occured")
	if err != nil {
		logger.Error(err.Error())
		return err
	}

	return nil
}

func errorHandler(err error, c echo.Context) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	if er, ok := err.(interface{ StackTrace() errors.StackTrace }); ok {
		logger.Error(
			err.Error(),
			zap.String(
				"stacktrace",
				fmt.Sprintf("%+v\n\n", er.StackTrace()),
			),
		)
	}
}

func stacktraced(c echo.Context) error {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	err := fmt.Errorf("error occured")
	if err != nil {
		return errors.Wrap(err, err.Error())
	}

	return nil
}
