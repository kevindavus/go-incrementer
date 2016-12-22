package main

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"

	"github.com/coopernurse/gorp"
	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
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

func initDatabase() *gorp.DbMap {
	db, err := sql.Open("sqlite3", "db.sqlite3")
	checkErr(err, "sql.Open failed")

	// construct a gorp DbMap
	dbmap := &gorp.DbMap{Db: db, Dialect: gorp.SqliteDialect{}}

	// add a table, setting the table name to 'numbers' and
	// specifying that the keys_id property is not an auto incrementing PK
	dbmap.AddTableWithName(Number{}, "numbers").SetKeys(false, "Key")

	// create the table. in a production system you'd generally
	// use a migration tool, or create the tables via scripts
	err = dbmap.CreateTablesIfNotExists()
	checkErr(err, "Create tables failed")
	return dbmap
}

func checkErr(err error, msg string) {
	if err != nil {
		log.Fatalln(msg, err)
	}
}

//NumberUpdate for incrementing values
func NumberUpdate(c *gin.Context) {
	var number Number
	var selected Number
	number.Key = c.PostForm("key")
	if number.Key != "" {

		var strVal = c.PostForm("value")
		if strVal == "" {
			strVal = "1"
		}
		var err error

		number.Value, err = strconv.ParseInt(strVal, 10, 64)

		if err != nil {
			panic(err)
		}

		err = dbmap.SelectOne(&selected, "select * from numbers where Key=?", number.Key)
		if err == nil {
			log.Println("selected row:", selected)

			if number.Key != "" {
				selected.Value = selected.Value + number.Value

				_, err = dbmap.Update(&selected)
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

}

//NumberPost Function for Posting new content to db
func NumberPost(c *gin.Context, number *Number) {

	fmt.Printf("\nPost %#v \n", number)
	err := dbmap.Insert(number)
	log.Println(err)
	if err == nil {
		c.JSON(201, &number)
	} else {
		c.JSON(401, gin.H{"error": "Could not insert into numbers"})
	}
}

// curl -X POST http://localhost:3333/increment -d 'key=abcdef&value=1'

//NumberList function for GET command
//returns full db contents
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

//DeleteNumber for removing entry from db
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
