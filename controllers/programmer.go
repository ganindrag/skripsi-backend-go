package controllers

import (
	"fmt"
	"os"
	"context"
	"strconv"
	"reflect"
	// "strings"
	"time"
	"crypto/md5"
    "encoding/hex"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/go-playground/validator/v10"
	"github.com/ganindrag/go-task-tracker/konst"
	"github.com/ganindrag/go-task-tracker/utils"
	jwt "github.com/form3tech-oss/jwt-go"
)

type Programmer struct {
	Id int `json:"id"`
	Name string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required,email"`
	Password string `json:"password,omitempty"`
	Role string `json:"role" validate:"required"`
}

func (prog Programmer) ValidateStruct() map[string]string {
    validate := validator.New()
    err := validate.Struct(prog)
    if err != nil {
    	return utils.ParseValidator(err.(validator.ValidationErrors))
    }
    return nil
}

func GetProgrammer(c *fiber.Ctx) error {
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Println(err.Error())
		return fiber.NewError(500, "Cannot connect to the database!")
	}
	defer conn.Close(context.Background());

	rows, err := conn.Query(context.Background(), "select id, name, email, role from programmer where id <> 1")
	if err != nil {
		fmt.Println(err.Error())
		return fiber.NewError(500, err.Error())
	}

	var result []Programmer
	for rows.Next() {
		var programmer Programmer
		if err := rows.Scan(&programmer.Id, &programmer.Name, &programmer.Email, &programmer.Role); err != nil {
			fmt.Println(err.Error())
			panic(err)
		} else {
			result = append(result, programmer)
		}
	}
	return c.JSON(result)
}

func GetSingleProgrammer(c *fiber.Ctx) error {
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Println(err.Error())
		return fiber.NewError(500, "Cannot connect to the database!")
	}
	defer conn.Close(context.Background());
	
	var prog Programmer
	prog.Id, err = strconv.Atoi(c.Params("id"))

	err = conn.QueryRow(context.Background(), "select id, name, email, role from programmer where id = $1", prog.Id).Scan(&prog.Id, &prog.Name, &prog.Email, &prog.Role)
	if err != nil {
		fmt.Println(err.Error())
		return fiber.NewError(500, "Data not found!")
	}

	return c.JSON(prog)
}

func CreateProgrammer(c *fiber.Ctx) error {
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Println(err.Error())
		return fiber.NewError(500, "Cannot connect to the database!")
	}
	defer conn.Close(context.Background())

	var prog Programmer
    if err := c.BodyParser(&prog); err != nil {
		return fiber.NewError(400, err.Error())
    }

    if errValidator := prog.ValidateStruct(); errValidator != nil {
		return c.Status(400).JSON(fiber.Map{
			"message": errValidator,
		})
    }

	err = conn.QueryRow(context.Background(), "select nextval('programmer_id_seq'::regclass);").Scan(&prog.Id)
	if err != nil {
		return err
	}

	commandTag, err := conn.Exec(context.Background(), "insert into programmer(name, email, password, role) values($1, $2, md5($3), $4);", prog.Name, prog.Email, prog.Password, prog.Role)
	if err != nil {
		fmt.Println(err.Error())
		if err, ok := err.(*pgconn.PgError); ok && err.Code == konst.UniqViolation {
			return fiber.NewError(500, "Data already been used!")
		}
  		return err
	}

	if commandTag.RowsAffected() > 0 {
		prog.Password = ""
		return c.Status(201).JSON(fiber.Map{
			"message": "Success",
			"data": prog,
		})
	}
	fmt.Println("prog not saved", prog)
	return c.JSON(fiber.Map{
		"message": "Success but not saved",
	})
}

func UpdateProgrammer(c *fiber.Ctx) error {
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Println(err.Error())
		return fiber.NewError(500, "Cannot connect to the database!")
	}
	defer conn.Close(context.Background())
	
	var prog Programmer
	prog.Id, err = strconv.Atoi(c.Params("id"))

	var idExists uint8
	err = conn.QueryRow(context.Background(), "select 1 from programmer where id = $1", prog.Id).Scan(&idExists)
	if err != nil {
		fmt.Println(err.Error())
		return fiber.NewError(404, "Data Not Found")
	}

    if err := c.BodyParser(&prog); err != nil {
		return fiber.NewError(400, err.Error())
    }

    if prog.Password != "" {
    	passByte := md5.Sum([]byte(prog.Password))
    	prog.Password = hex.EncodeToString(passByte[:])
    }
    
  //   if errValidator := ValidateStruct(prog); errValidator != nil {
		// return c.Status(400).JSON(fiber.Map{
		// 	"message": errValidator,
		// })
  //   }

    sqlUpdate, paramsUpdate := utils.ParseStructToUpdateSql(reflect.TypeOf(prog), reflect.ValueOf(prog))
    sql := fmt.Sprintf("update programmer set %s where id = $1;", sqlUpdate)

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

func DeleteProgrammer(c *fiber.Ctx) error {
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Println(err.Error())
		return fiber.NewError(500, "Cannot connect to the database!")
	}
	defer conn.Close(context.Background())
	
	var prog Programmer
	prog.Id, err = strconv.Atoi(c.Params("id"))

	commandTag, err := conn.Exec(context.Background(), "delete from feedback where programmer_id = $1;", prog.Id)
	commandTag, err = conn.Exec(context.Background(), "delete from programmer where id = $1;", prog.Id)
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

func Login(c *fiber.Ctx) error {
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Println(err.Error())
		return fiber.NewError(500, "Cannot connect to the database!")
	}
	defer conn.Close(context.Background())

	var prog Programmer
    if err := c.BodyParser(&prog); err != nil {
		return fiber.NewError(400, err.Error())
    }

	err = conn.QueryRow(context.Background(), "select id, name, email, password, role from programmer where email=$1 and password=md5($2)", prog.Email, prog.Password).Scan(&prog.Id, &prog.Name, &prog.Email, &prog.Password, &prog.Role)
	if err != nil {
		return fiber.NewError(401, "Email or Password incorrect!")
	}

	token := jwt.New(jwt.SigningMethodHS256)

	// Set claims
	claims := token.Claims.(jwt.MapClaims)
	claims["id"] = prog.Id
	claims["name"] = prog.Name
	claims["email"] = prog.Email
	claims["role"] = prog.Role
	claims["exp"] = time.Now().Add(time.Hour * 8).Unix()

	// Generate encoded token and send it as response.
	t, err := token.SignedString([]byte("mysecret"))
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.JSON(fiber.Map{"token": t})
}


func CheckAuth(c *fiber.Ctx) error {
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Println(err.Error())
		return fiber.NewError(500, "Cannot connect to the database!")
	}
	defer conn.Close(context.Background());
	
	user := c.Locals("user").(*jwt.Token).Claims.(jwt.MapClaims)
	var prog Programmer

	err = conn.QueryRow(context.Background(), "select id, name, email, role from programmer where id = $1", user["id"].(float64)).Scan(&prog.Id, &prog.Name, &prog.Email, &prog.Role)
	if err != nil {
		fmt.Println(err.Error())
		return fiber.NewError(500, "Data User not found!")
	}

	return c.JSON(prog)
}