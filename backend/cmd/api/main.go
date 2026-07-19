package main

import (
	"log"
	"os"
	"os/signal"
	_ "server_nesting_optimizer/docs"
	"server_nesting_optimizer/internal/app"
	"syscall"
)

// @Version 1.0.0
// @Title Nesting Optimizer API
// @Description Backend API for nonlinear dense nesting of flat surfaces.
// @ContactName Diana Podshivalenko
// @SecurityDefinitions.apikey BearerAuth
// @In header
// @Name Authorization
// @Description Введите токен в формате: Bearer <access_token>
// @Server http://localhost:8082 Local server
func main() {
	application, err := app.New()
	if err != nil {
		log.Fatal(err)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(
		stop,
		os.Interrupt,
		syscall.SIGTERM,
	)

	go func() {
		if err := application.Run(); err != nil {
			log.Fatal(err)
		}
	}()

	<-stop

	application.Stop()
}
