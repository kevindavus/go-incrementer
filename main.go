package main

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
	"gopkg.in/gorp.v2"
)

var dbmap *gorp.DbMap

//Number key and value pair
type Number struct {
	Key   string `db:"Key" form:"key"`
	Value int64  `db:"Value" form:"value"`
}

func main() {
	// initialize the DbMap
	dbmap = initDatabase()
	defer dbmap.Db.Close()

	app := gin.Default()
	app.GET("/increment", NumberList)      //Returns current contents of db
	app.POST("/increment", NumberUpdate)   //First tries to update, if the key is not present it posts
	app.PUT("/increment", NumberUpdate)    //Although the gist just mentioned POST I thought PUT would be a better fit
	app.DELETE("/increment", DeleteNumber) //Deletes entry with specified key

	app.Run(":3333")

}

/*
  initDatabase creates sqlite database that allows 1 open connection at a time to prevent
  databse lock issues. Allows 1000 idle connections as a fail safe in case there is some
  hold up. setting the lifetime to 0 means it won't time out.
*/
func initDatabase() *gorp.DbMap {
	db, err := sql.Open("sqlite3", "./numbers.db")
	checkErr(err, "sql.Open failed")

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1000)
	db.SetConnMaxLifetime(0)
	// construct a gorp DbMap
	dbmap = &gorp.DbMap{Db: db, Dialect: gorp.SqliteDialect{}}

	// add a table, setting the table name to 'numbers' and
	// specifying that the Key property is not an auto incrementing PK
	dbmap.AddTableWithName(Number{}, "numbers").SetKeys(false, "Key")

	// create the table. in a production system you'd generally
	// use a migration tool, or create the tables via scripts
	err = dbmap.CreateTablesIfNotExists()
	checkErr(err, "Create tables failed")
	return dbmap
}

//checkErr simple error handling
func checkErr(err error, msg string) {
	if err != nil {
		log.Fatalln(msg, err)
	}
}

/*
  NumberUpdate for incrementing values. If the item does not currently exist, the command
  passes through to NumberPost
*/
func NumberUpdate(c *gin.Context) {
	var number Number              //content from curl request
	var selected Number            //content from db
	number.Key = c.PostForm("key") //grab key value
	if number.Key != "" {

		var strVal = c.PostForm("value") //grab passed in value
		if strVal == "" {                //if value is not present default to 1
			strVal = "1"
		}
		var err error

		number.Value, err = strconv.ParseInt(strVal, 10, 64)

		if err != nil {
			checkErr(err, "error in converting string value to integer")
		}

		err = dbmap.SelectOne(&selected, "select * from numbers where Key=?", number.Key)
		if err == nil {
			log.Println("selected row:", selected)

			if number.Key != "" {
				selected.Value = selected.Value + number.Value //increment value

				_, err = dbmap.Update(&selected) //update the value in the db

				fmt.Printf("\nUpdate %#v %#v\n", number, selected)
				if err == nil {
					c.JSON(200, selected)
				} else {
					checkErr(err, "Updated failed")
				}
			}
		} else {
			NumberPost(c, &number)
		}
	} else {
		c.JSON(400, gin.H{"error": "Missing key value"})
	}
	//If key is not present in db calls NumberPost
	// curl -X POST http://localhost:3333/increment -d 'key=abcdef&value=1'
	// curl -X PUT http://localhost:3333/increment -d 'key=abcdef&value=1'
	// curl -X PUT http://localhost:3333/increment -d 'key=abcdef'

}

/*
  NumberPost Function for Posting new content to db
*/
func NumberPost(c *gin.Context, number *Number) {

	fmt.Printf("\nPost %#v \n", number)
	err := dbmap.Insert(number) //inserts the data gathered from NumberUpdate to the db
	log.Println(err)
	if err == nil {
		c.JSON(201, &number)
	} else {
		c.JSON(401, gin.H{"error": "Could not insert into numbers"})
	}
}

// curl -X POST http://localhost:3333/increment -d 'key=abcdef&value=1'

/*
  NumberList function for GET command
  returns full db contents
*/
func NumberList(c *gin.Context) {
	var number []Number
	_, err := dbmap.Select(&number, "SELECT * FROM numbers")

	if err == nil {
		c.JSON(200, number)
	} else {
		c.JSON(404, gin.H{"error": "no number(s) in the table"})
	}
	// curl -X GET http://localhost:3333/increment

}

/*
  DeleteNumber for removing entry from db
*/
func DeleteNumber(c *gin.Context) {
	var number Number
	var selected Number
	number.Key = c.PostForm("key")
	log.Println("selected row:", number.Key)
	if number.Key != "" {

		var err error

		err = dbmap.SelectOne(&selected, "select * from numbers where Key=?", number.Key)
		log.Println(err)
		if err == nil {
			selected.Value = selected.Value + number.Value

			_, err = dbmap.Delete(&selected)
			if err == nil {
				fmt.Printf("Delete %#v %#v\n", number, selected)
				c.JSON(200, selected)
			} else {
				checkErr(err, "Deletion failed")
			}
		}

	} else {
		c.JSON(404, gin.H{"error": "key could not be found"})
	}
}

// curl -X DELETE http://localhost:3333/increment
