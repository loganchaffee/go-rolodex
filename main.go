package main

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
	_ "github.com/mattn/go-sqlite3"
)

type Contact struct {
	ID    int64  `json:"name"`
	Name  string `json:"name"`
	Phone string `json:"phone"`
}

// Util func for rendering templ components
func render(c echo.Context, component templ.Component) error {
	return component.Render(c.Request().Context(), c.Response())
}

func main() {
	// DB
	db, err := sql.Open("sqlite3", "./go-rolodex.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Router
	app := echo.New()

	// Middleware

	app.Static("/static", "static")

	// Get contacts
	app.GET("/", func(c echo.Context) error {
		// Query
		rows, err := db.Query("select * from contact", 1)
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		// Scan rows to structs
		var contacts []Contact
		for rows.Next() {
			var contact Contact
			err := rows.Scan(&contact.ID, &contact.Name, &contact.Phone)
			if err != nil {
				log.Fatal(err)
			}

			contacts = append(contacts, contact)
		}
		err = rows.Err()
		if err != nil {
			log.Fatal(err)
		}

		// Render template send reponse
		return render(c, index(contacts))
	})

	app.GET("/edit/:id", func(c echo.Context) error {
		id := c.Param("id")

		result := db.QueryRow("select * from contact where id = ?;", id)

		var contact Contact
		err := result.Scan(&contact.ID, &contact.Name, &contact.Phone)
		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound, "Something went wrong")
		}

		return render(c, contactListItem(contact, true))
	})

	app.POST("/", func(c echo.Context) error {
		err := c.Request().ParseForm()
		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound, "Something went wrong")
		}

		name := c.Request().FormValue("name")
		phone := c.Request().FormValue("phone")

		result, err := db.Exec("insert into contact (name, phone) values (?, ?);", name, phone)
		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound, "Something went wrong")
		}

		id, err := result.LastInsertId()
		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound, "Something went wrong")
		}

		newContact := Contact{id, name, phone}

		return render(c, contactListItem(newContact, false))
	})

	app.PUT("/:id", func(c echo.Context) error {
		id := c.Param("id")

		name := c.Request().FormValue("name")
		phone := c.Request().FormValue("phone")

		_, err := db.Exec("update contact set name = ?, phone = ? where id = ?;", name, phone, id)
		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound, "Something went wrong")
		}

		idInt, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound, "Something went wrong")
		}

		newContact := Contact{idInt, name, phone}

		return render(c, contactListItem(newContact, false))
	})

	app.DELETE("/:id", func(c echo.Context) error {
		id := c.Param("id")

		_, err := db.Exec("delete from contact where id = ?;", id)
		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound, "Something went wrong")
		}

		return nil
	})

	app.Logger.Fatal(app.Start(":3000"))
}
