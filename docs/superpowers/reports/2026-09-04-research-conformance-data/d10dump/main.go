// D10 read-only harness: builds the 1D expectancy table through the PRODUCTION
// code path (expectancy.LoadAndBuildAt) against a read-only DSN. Writes nothing.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"nofx/expectancy"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gl "gorm.io/gorm/logger"
)

func main() {
	dsn := os.Args[1]
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{Logger: gl.Default.LogMode(gl.Silent)})
	if err != nil {
		fmt.Println("open:", err)
		os.Exit(1)
	}
	now := time.Now()
	t, err := expectancy.LoadAndBuildAt(db, now)
	if err != nil {
		fmt.Println("build:", err)
		os.Exit(1)
	}
	fmt.Println(t.BootLine())
	b, _ := json.MarshalIndent(t, "", " ")
	fmt.Println(string(b))
}
