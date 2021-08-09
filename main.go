package main

import (
	"fmt"
	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"github.com/satori/go.uuid"
)

const (
	sentryDSN = "http://88fa050b1c09432f8da6c99804eb90b0@192.168.159.180:9000/6"
	environmentDev = "Dev"
	userId = "USER_ID"
)
func main() {
	err := sentry.Init(sentry.ClientOptions{
		//TracesSampleRate: 0.2,
		TracesSampler: sentry.TracesSamplerFunc(func(ctx sentry.SamplingContext) sentry.Sampled {
			return sentry.UniformTracesSampler(0.2).Sample(ctx)

		}),
		Environment: environmentDev,
		Dsn: sentryDSN,
		BeforeSend: func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
			if hint.Context != nil {
				if req, ok := hint.Context.Value(sentry.RequestContextKey).(*http.Request); ok {
					// You have access to the original Request
					fmt.Println(req)
				}
			}
			return event
		},
		Debug:            true,
		AttachStacktrace: true,
	})
	if err != nil {
		log.Fatalf("sentry.Init: %s", err)
	}

	sentry.CaptureMessage("It works!")

	app := gin.Default()

	app.Use(sentrygin.New(sentrygin.Options{
		Repanic: true,
	}))

	app.Use(func(ctx *gin.Context) {
		if hub := sentrygin.GetHubFromContext(ctx); hub != nil {
			hub.Scope().SetTag("RANDOM_KEYS", "RANDOM_VALUES")
		}
		ctx.Next()
	})

	app.GET("/", func(ctx *gin.Context) {
		uid := uuid.NewV4()
		ctx.Set(userId,uid.String())
		if hub := sentrygin.GetHubFromContext(ctx); hub != nil {

			hub.WithScope(func(scope *sentry.Scope) {
				scope.SetExtra("unwantedQuery", "someQueryDataMaybe")
				scope.SetExtra("manual-key", "manual-value")
				// setting a userId to Tag
				scope.SetTag(userId,ctx.GetString(userId))
				hub.CaptureMessage("User provided unwanted query string, but we recovered just fine")

			})
		}
		ctx.Status(http.StatusOK)
	})

	app.GET("/foo", func(ctx *gin.Context) {
		uid := uuid.NewV4()
		ctx.Set(userId,uid.String())
		// sentrygin handler will catch it just fine, and because we attached "someRandomTag"
		// in the middleware before, it will be sent through as well
		sentry.WithScope(func(scope *sentry.Scope) {
			scope.SetTag(userId, ctx.GetString(userId))
			scope.SetLevel(sentry.LevelError)
			// will be tagged with my-tag="my value"
			sentry.CaptureException(fmt.Errorf("an error raised"))
		})
		panic("panic error")

	})

	_ = app.Run(":3000")
}