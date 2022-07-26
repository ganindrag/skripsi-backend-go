package controllers

import (
	"fmt"
	"os"
	"context"
	"strconv"
	// "strings"
	"time"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v4"
	"github.com/go-playground/validator/v10"
	"github.com/ganindrag/go-task-tracker/utils"
	jwt "github.com/form3tech-oss/jwt-go"
)

type Feedback struct {
	Id int `json:"id"`
	ProgrammerId *int `json:"programmer_id" pgxcolumn:"programmer_id"`
	TaskId *int `json:"task_id" pgxcolumn:"task_id" validate:"required"`
	Feedback string `json:"feedback" validate:"required"`
	CreatedAt *time.Time `json:"created_at" pgxcolumn:"created_at"`
	ProgrammerName *string `json:"programmer_name" pgxignored:""`
}

func (feedback Feedback) ValidateStruct() map[string]string {
    validate := validator.New()
    err := validate.Struct(feedback)
    if err != nil {
    	return utils.ParseValidator(err.(validator.ValidationErrors))
    }
    return nil
}

func GetFeedback(c *fiber.Ctx) error {
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Println(err.Error())
		return fiber.NewError(500, "Cannot connect to the database!")
	}
	defer conn.Close(context.Background());

	taskId, _ := strconv.Atoi(c.Query("task_id"))

	rows, err := conn.Query(context.Background(), "select feedback.id, feedback.feedback, created_at, programmer.name from feedback join programmer on programmer_id=programmer.id where task_id = $1 order by created_at", taskId)
	if err != nil {
		fmt.Println(err.Error())
		return fiber.NewError(500, err.Error())
	}
	fmt.Println(taskId)

	var result []Feedback
	for rows.Next() {
		var feedback Feedback
		if err := rows.Scan(&feedback.Id, &feedback.Feedback, &feedback.CreatedAt, &feedback.ProgrammerName); err != nil {
			fmt.Println(err.Error())
			panic(err)
		} else {
			result = append(result, feedback)
		}
	}
	return c.JSON(result)
}

func CreateFeedback(c *fiber.Ctx) error {
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Println(err.Error())
		return fiber.NewError(500, "Cannot connect to the database!")
	}
	defer conn.Close(context.Background())

	user := c.Locals("user").(*jwt.Token).Claims.(jwt.MapClaims)

	var feedback Feedback
  if err := c.BodyParser(&feedback); err != nil {
		return fiber.NewError(400, err.Error())
  }

  if errValidator := feedback.ValidateStruct(); errValidator != nil {
		return c.Status(400).JSON(fiber.Map{
			"message": errValidator,
		})
  }

	err = conn.QueryRow(context.Background(), "select nextval('feedback_id_seq'::regclass);").Scan(&feedback.Id)
	if err != nil {
		return err
	}

	commandTag, err := conn.Exec(context.Background(), "insert into feedback(programmer_id,task_id,feedback) values($1, $2, $3);", user["id"], feedback.TaskId, feedback.Feedback)
	if err != nil {
		return err
	}

	if commandTag.RowsAffected() > 0 {
		return c.Status(201).JSON(fiber.Map{
			"message": "Success",
			"data": feedback,
		})
	}

	return c.JSON(fiber.Map{
		"message": "Success but not saved",
	})
}
