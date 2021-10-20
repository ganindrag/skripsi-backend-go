package controllers

import (
	"fmt"
	"os"
	"context"
	"strconv"
	"time"
	"errors"
	"reflect"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/go-playground/validator/v10"
	jwt "github.com/form3tech-oss/jwt-go"
	"github.com/ganindrag/go-task-tracker/utils"
	"github.com/ganindrag/go-task-tracker/konst"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Task struct {
	Id int `json:"id"`
	ProgrammerId *int `json:"programmer_id" pgxcolumn:"programmer_id"`
	Name string `json:"name" validate:"required"`
	Detail *string `json:"detail"`
	Status string `json:"status" validate:"required"`
	Weight *int `json:"weight"`
	CreatedAt *time.Time `json:"created_at" pgxcolumn:"created_at"`
	StartAt *time.Time `json:"start_at" pgxcolumn:"start_at"`
	EndAt *time.Time `json:"end_at" pgxcolumn:"end_at"`
	BugTolerance *int `json:"bug_tolerance" pgxcolumn:"bug_tolerance"`
	ActualBug *int `json:"actual_bug" pgxcolumn:"actual_bug"`
	Comprehension *int `json:"comprehension"`
	IsEvaluated bool `json:"is_evaluated" pgxcolumn:"is_evaluated"`
	ProgrammerName *string `json:"programmer_name" pgxignored:""`
}

func (task Task) ValidateStruct() map[string]string {
    validate := validator.New()
    err := validate.Struct(task)
    if err != nil {
    	return utils.ParseValidator(err.(validator.ValidationErrors))
    }
    return nil
}

func GetTask(c *fiber.Ctx) error {
	conn, err := gorm.Open(postgres.Open(os.Getenv("DB_DSN")), &gorm.Config{})
	if err != nil {
		fmt.Println(err.Error())
		return fiber.NewError(500, "Cannot connect to the database!")
	}

	user := c.Locals("user").(*jwt.Token).Claims.(jwt.MapClaims)
	fmt.Println(user)

	q := conn.Debug().Table("task").Joins("left join programmer on task.programmer_id=programmer.id").Select("task.*, programmer.name as programmer_name")

	if user["role"].(string) == "USER" {
		q.Where("(programmer.id = ? OR programmer_id is null)", user["id"].(float64))
	}

	start_at := c.Query("start_at")
	end_at := c.Query("end_at")
	if start_at != "" && end_at != "" {
		q.Where("(start_at::date between ?::date and ?::date and end_at::date <= ?) or (start_at is null or end_at is null)", start_at, end_at, end_at)
	}

	status := c.Query("status")
	if status != "" {
		q.Where("status = ?", status)
	}

	iseval := c.Query("iseval")
	if iseval != "" {
		q.Where("is_evaluated = ?", iseval)
	}

	prog_id := c.Query("prog_id")
	if user["role"].(string) != "USER" && prog_id != "" {
		q.Where("programmer_id = ?", prog_id)
	}

	var result []Task
	q.Find(&result)

	return c.JSON(result)
}

func GetSingleTask(c *fiber.Ctx) error {
	conn, err := gorm.Open(postgres.Open(os.Getenv("DB_DSN")), &gorm.Config{})
	if err != nil {
		fmt.Println(err.Error())
		return fiber.NewError(500, "Cannot connect to the database!")
	}
	
	var task Task
	task.Id, err = strconv.Atoi(c.Params("id"))
	err = conn.Table("task").Take(&task, task.Id).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		fmt.Println(err.Error())
		return fiber.NewError(404, "Data not found!")
	}

	return c.JSON(task)
}

func CreateTask(c *fiber.Ctx) error {
	conn, err := pgx.Connect(context.Background(), os.Getenv("DB_DSN"))
	if err != nil {
		fmt.Println(err.Error())
		return fiber.NewError(500, "Cannot connect to the database!")
	}
	defer conn.Close(context.Background())

	var task Task
    if err := c.BodyParser(&task); err != nil {
		return fiber.NewError(400, err.Error())
    }
    // task.CreatedAt.Set(task.EncodedCreatedAt)
    // task.StartAt.Set(task.EncodedStartAt)
    // task.EndAt.Set(task.EncodedEndAt)

    if errValidator := task.ValidateStruct(); errValidator != nil {
		return c.Status(400).JSON(fiber.Map{
			"message": errValidator,
		})
    }
// TODO: pengecekan programmer_id

	err = conn.QueryRow(context.Background(), "select nextval('task_id_seq'::regclass);").Scan(&task.Id)
	if err != nil {
		return err
	}
	
	columnsInsert, paramsInsert, valuesInsert := utils.ParseStructToInsertSql(reflect.TypeOf(task), reflect.ValueOf(task))
	sql := fmt.Sprintf("insert into task(%s) values(%s);", columnsInsert, paramsInsert)

	commandTag, err := conn.Exec(context.Background(), sql, valuesInsert...)
	if err != nil {
		fmt.Println(sql)
		if err, ok := err.(*pgconn.PgError); ok && err.Code == konst.FkViolation {
			return fiber.NewError(400, "Data programmer not found!")
		}
  		return err
	}

	if commandTag.RowsAffected() > 0 {
		return c.Status(201).JSON(fiber.Map{
			"message": "Success",
			"data": task,
		})
	}
	fmt.Println("task not saved", task)
	return c.JSON(fiber.Map{
		"message": "Success but not saved",
	})
}

func UpdateTask(c *fiber.Ctx) error {
	conn, err := pgx.Connect(context.Background(), os.Getenv("DB_DSN"))
	if err != nil {
		fmt.Println(err.Error())
		return fiber.NewError(500, "Cannot connect to the database!")
	}
	defer conn.Close(context.Background())
	
	var task Task
	task.Id, err = strconv.Atoi(c.Params("id"))

	var idExists uint8
	err = conn.QueryRow(context.Background(), "select 1 from task where id = $1", task.Id).Scan(&idExists)
	if err != nil {
		fmt.Println(err.Error())
		return fiber.NewError(404, "Data Not Found")
	}

    if err := c.BodyParser(&task); err != nil {
		return fiber.NewError(400, err.Error())
    }
    // task.CreatedAt.Set(task.EncodedCreatedAt)
    // task.StartAt.Set(task.EncodedStartAt)
    // task.EndAt.Set(task.EncodedEndAt)

    sqlUpdate, paramsUpdate := utils.ParseStructToUpdateSql(reflect.TypeOf(task), reflect.ValueOf(task))
    sql := fmt.Sprintf("update task set %s where id = $1;", sqlUpdate)

	_, err = conn.Exec(context.Background(), sql, paramsUpdate...)
	if err != nil {
		fmt.Println(err.Error())
		fmt.Println(sql)
		fmt.Println(sqlUpdate)
  		return err
	}

	return c.JSON(fiber.Map{
		"message": "Success",
	})
}

func DeleteTask(c *fiber.Ctx) error {
	conn, err := pgx.Connect(context.Background(), os.Getenv("DB_DSN"))
	if err != nil {
		fmt.Println(err.Error())
		return fiber.NewError(500, "Cannot connect to the database!")
	}
	defer conn.Close(context.Background())
	
	var task Task
	task.Id, err = strconv.Atoi(c.Params("id"))

	commandTag, err := conn.Exec(context.Background(), "delete from task where id = $1;", task.Id)
	if err != nil {
		fmt.Println(err.Error())
  		return err
	}

	if commandTag.RowsAffected() > 0 {
		return c.JSON(fiber.Map{
			"message": "Success",
		})
	}

	return fiber.NewError(404, "Data not found!")
}

func GetReport(c *fiber.Ctx) error {
	conn, err := gorm.Open(postgres.Open(os.Getenv("DB_DSN")), &gorm.Config{})
	if err != nil {
		fmt.Println(err.Error())
		return fiber.NewError(500, "Cannot connect to the database!")
	}

	user := c.Locals("user").(*jwt.Token).Claims.(jwt.MapClaims)
	fmt.Println(user)

	q := conn.Debug().Table("task").Joins("left join programmer on task.programmer_id=programmer.id").Select("task.*, programmer.name as programmer_name").Order("start_at")

	if user["role"].(string) == "USER" {
		q.Where("(programmer.id = ? OR programmer_id is null)", user["id"].(float64))
	}

	start_at := c.Query("start_at")
	end_at := c.Query("end_at")
	if start_at != "" && end_at != "" {
		q.Where("(start_at::date between ?::date and ?::date and (end_at::date <= ? or end_at is null))", start_at, end_at, end_at)
	}
	q.Where("(end_at is null or start_at <= end_at)")

	iseval := c.Query("iseval")
	if iseval != "" {
		q.Where("is_evaluated = ?", iseval)
	}

	prog_id := c.Query("prog_id")
	if user["role"].(string) != "USER" && prog_id != "" {
		q.Where("programmer_id = ?", prog_id)
	}

	var result []Task
	q.Find(&result)

	var total float64
	var countEvaluated float64
	var hasUnEval bool = false
	var totalGrade float64 = 0.0
	grade := ""
	if len(result) > 0 {
		for _, task := range result {
			if task.IsEvaluated {
				startAt := *task.StartAt
				endAt := *task.EndAt
				bugTolerance := *task.BugTolerance
				actualBug := *task.ActualBug
				weight := *task.Weight
				comprehension := *task.Comprehension

				bugPercentage := float64(bugTolerance) / float64(actualBug)
				bugGrade := 50.0
				if bugPercentage < 1 {
					bugGrade = bugGrade * bugPercentage
				}

				actualDate := endAt.Sub(startAt).Hours() / 24
				datePercentage := actualDate / float64(weight)
				dateGrade := 30.0
				if datePercentage < 1 {
					dateGrade = dateGrade * datePercentage
				}

				comprehensionGrade := float64(comprehension) / 100 * 20

				total += bugGrade + dateGrade + comprehensionGrade
				countEvaluated += 1
			} else {
				hasUnEval = true
			}
		}

		if countEvaluated > 0 {
			grade = "A"
			totalGrade = total / countEvaluated
			if totalGrade >= 75 && totalGrade <= 84 {
				grade = "B"
			} else if totalGrade >= 60 && totalGrade <= 74 {
				grade = "C"
			} else if totalGrade >= 50 && totalGrade <= 59 {
				grade = "D"
			} else if totalGrade < 50 {
				grade = "E"
			}
		}
	}

	fmt.Println(result, grade, totalGrade, hasUnEval)

	return c.JSON(fiber.Map{
		"data": result,
		"grade": grade,
		"numberGrade": totalGrade,
		"hasUnEvaluated": hasUnEval,
	})
}
