package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/cockroachdb/cockroach-go/v2/crdb/crdbgorm"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Reservation model having the rerqired data
type Reservation struct {
	Service   string
	UserID    uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4()"`
	StartTime time.Time
	EndTime   time.Time
}

var reservations []Reservation

var reserveID []uuid.UUID

func bookService(db *gorm.DB, userID int, service string, starttime, endtime time.Time) error {
	log.Printf("Booking a %s for you.....", service)
	for i := 0; i < userID; i = i + 1 {
		idnum := uuid.New()

		if err := db.Create(&Reservation{UserID: idnum}).Error; err != nil {
			return err
		}
		reserveID = append(reserveID, idnum)
	}
	log.Println("Reservation created ......")
	return nil
}

func available(service string, starttime, endtime time.Time) bool {
	for _, r := range reservations {
		if r.Service == service && !(endtime.Before(r.StartTime) || starttime.After(r.EndTime)) {
			return false
		}
	}
	return true
}
func reservation(db *gorm.DB, reserveID []uuid.UUID) error {
	log.Println("Deleting reserve created...")
	err := db.Where("user_id IN ?", reserveID).Delete(Reservation{}).Error
	if err != nil {
		return err
	}
	log.Println("Reservation deleted.")
	return nil
}

func printreserve(db *gorm.DB) {
	var reserve []Reservation
	db.Find(&reservations)
	fmt.Printf("Reserve at %s: \n", time.Now())

	for _, r := range reserve {
		fmt.Printf("%s %s %s :\n", r.Service, r.StartTime, r.EndTime)
	}
}

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Some error occured. Err: %s", err)
	}
	db, err := gorm.Open(postgres.Open(os.Getenv("DATABASE_URL")+"&application_name=$ docs_simplecrud_gorm"), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	// Automatically create the "accounts" table based on the `Account`
	// model.
	db.AutoMigrate(&Reservation{
		Service:   "",
		UserID:    [16]byte{},
		StartTime: time.Time{},
		EndTime:   time.Time{},
	})
	if err := crdbgorm.ExecuteTx(context.Background(), db, nil,
		func(tx *gorm.DB) error {
			return bookService(db, 56, "resort", time.Now(), time.Now().Add(time.Hour))
		},
	); err != nil {
		// For information and reference documentation, see:
		//   https://www.cockroachlabs.com/docs/stable/error-handling-and-troubleshooting.html
		fmt.Println(err)
	}
	printreserve(db)
	reservation(db, reserveID)
	available("resort", time.Now(), time.Now().Add(time.Hour))

	if err := crdbgorm.ExecuteTx(context.Background(), db, nil,
		func(db *gorm.DB) error {
			return reservation(db, reserveID)

		}); err != nil {
		fmt.Println(err)
	}
}
