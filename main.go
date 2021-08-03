package main

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
    jwtware "github.com/gofiber/jwt/v2"
	"github.com/joho/godotenv"
	"github.com/ganindrag/go-task-tracker/controllers"
    "github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	err := godotenv.Load()
    fmt.Println(err)
    app := fiber.New(fiber.Config{
	    ErrorHandler: func(ctx *fiber.Ctx, err error) error {
	        code := fiber.StatusInternalServerError
	        message := "Internal Server Error"
	        if e, ok := err.(*fiber.Error); ok {
	            code = e.Code
	            message = e.Message
	        }else {
	        	fmt.Println(err)
	        }
            return ctx.Status(code).JSON(fiber.Map{
            	"message": message,
            })
	    },
	})

    app.Use(cors.New())

    app.Post("/login", controllers.Login)
    
    app.Use(jwtware.New(jwtware.Config{
        SigningKey: []byte("mysecret"),
        ErrorHandler: func(ctx *fiber.Ctx, err error) error {
            return ctx.Status(401).JSON(fiber.Map{
                // "message": err.Error(),
                "message": "Autentikasi gagal!",
            })
        },
    }))
    
    app.Get("/auth", controllers.CheckAuth);

    app.Get("/programmer", controllers.GetProgrammer);
    app.Get("/programmer/:id", controllers.GetSingleProgrammer);
    app.Post("/programmer", controllers.CreateProgrammer);
    app.Put("/programmer/:id", controllers.UpdateProgrammer);
    app.Delete("/programmer/:id", controllers.DeleteProgrammer);

    app.Get("/task", controllers.GetTask);
    app.Get("/task/:id", controllers.GetSingleTask);
    app.Post("/task", controllers.CreateTask);
    app.Put("/task/:id", controllers.UpdateTask);
    app.Delete("/task/:id", controllers.DeleteTask);

    app.Get("/report", controllers.GetReport);

    app.Listen(":3000")
}