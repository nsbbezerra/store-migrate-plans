package main

import (
	"log"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Order struct {
	ID      uuid.UUID `db:"id"`
	Ccode   string    `db:"customer_code"`
	Onumber string    `db:"order_number"`
}

type OderProduct struct {
	ID        uuid.UUID `db:"id"`
	ProductID uuid.UUID `db:"product_id"`
	Cprofile  string    `db:"customer_profile"`
	OId       uuid.UUID `db:"order_id"`
	Status    string    `db:"acceptance_status"`
}

type OrderProductItem struct {
	ID         uuid.UUID  `db:"id"`
	Oproduct   string     `db:"order_product_id"`
	AircraftID string     `db:"aircraft_model_id"`
	ItemID     uuid.UUID  `db:"item_id"`
	StartDate  *time.Time `db:"start_date"`
	EndDate    *time.Time `db:"end_date"`
	ExcludedIn *time.Time `db:"excluded_in"`
	Status     *string    `db:"status"`
	Quantity   int        `db:"quantity"`
	SubsPeriod *string    `db:"subscription_period"`
}

type AircraftItems struct {
	ID        uuid.UUID `db:"id"`
	Caircraft uuid.UUID `db:"customer_aircraft_id"`
	ItemID    uuid.UUID `db:"order_products_item_id"`
}

type ManagePlans struct {
	ID            uuid.UUID `db:"id"`
	Ccode         string    `db:"customer_code"`
	OrderID       uuid.UUID `db:"order_id"`
	ProductID     uuid.UUID `db:"product_id"`
	ServiceCenter bool      `db:"service_center"`
	IsAdmin       string    `db:"is_admin"`
}

type ManagePlansItem struct {
	ID            uuid.UUID  `db:"id"`
	ManagePlansID uuid.UUID  `db:"manage_plans_id"`
	ProductItemID uuid.UUID  `db:"product_item_id"`
	AircraftID    string     `db:"aircraft_model_id"`
	StartDate     *time.Time `db:"start_date"`
	EndDate       *time.Time `db:"end_date"`
	BlockingDate  *time.Time `db:"blocking_date"`
	Status        *string    `db:"status"`
	SubsPeriod    int        `db:"subscription_period"`
	Quantity      int        `db:"quantity"`
	IsAdmin       string     `db:"is_admin"`
}

type ManagePlansAircraft struct {
	ID                uuid.UUID `db:"id"`
	Caircraft         uuid.UUID `db:"aircraft_id"`
	ManagePlansItemID uuid.UUID `db:"manage_plans_item_id"`
}

func main() {

	db, err := sqlx.Connect("postgres", "user=svq dbname=services_quotation sslmode=disable password=svq host=localhost")

	if err != nil {
		log.Fatalln(err)
	}

	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	} else {
		log.Println("Successfully Connected")
	}

	order := Order{}

	orders, oerror := db.Queryx("SELECT id, customer_code, order_number FROM orders")

	if oerror != nil {
		log.Fatalln(oerror)
	}

	/* ------------------------------------------------------------------- MAP ORDERS -------------------------------------------------------------- */

	log.Println("Save process is init!")

	for orders.Next() {
		err := orders.StructScan(&order)
		if err != nil {
			log.Fatalln(err)
		}
		oproduct := OderProduct{}

		products, perror := db.Queryx("SELECT id, product_id, customer_profile, order_id, acceptance_status FROM order_products WHERE order_id = $1", order.ID)

		if perror != nil {
			log.Fatalln(perror)
		}

		defer products.Close()

		/* ------------------------------------------------------------------- MAP PRODUCTS -------------------------------------------------------------- */

		for products.Next() {
			err := products.StructScan(&oproduct)
			if err != nil {
				log.Fatalln(err)
			}

			/* ------------------------------------------------------------------- INIT SAVE PLAN -------------------------------------------------------------- */

			var isServiceCenter bool
			var isAdmin string

			if oproduct.Cprofile == "easc" {
				isServiceCenter = true
			} else {
				isServiceCenter = false
			}

			if oproduct.Status == "Support Test" {
				isAdmin = "yes"
			} else {
				isAdmin = "no"
			}

			newPlan := ManagePlans{
				OrderID:       order.ID,
				Ccode:         order.Ccode,
				ProductID:     oproduct.ProductID,
				ServiceCenter: isServiceCenter,
				IsAdmin:       isAdmin,
			}

			queryPlans := `INSERT INTO manage_plans (customer_code, order_id, product_id, service_center, is_admin) VALUES ($1, $2, $3, $4, $5) RETURNING id`

			var id uuid.UUID

			mp_error := db.QueryRow(queryPlans, order.Ccode, order.ID, oproduct.ProductID, isServiceCenter, isAdmin).Scan(&id)

			if mp_error != nil {
				log.Fatalln(mp_error)
			}

			newPlan.ID = id

			/* ------------------------------------------------------------------------- FINISH SAVE PLAN -------------------------------------------------------------- */

			opitem := OrderProductItem{}

			items, ierror := db.Queryx("SELECT id, order_product_id, aircraft_model_id, item_id, quantity, subscription_period, start_date, end_date, status, excluded_in FROM order_products_items WHERE order_product_id = $1 AND id NOT IN (SELECT order_product_item_id FROM optionals_pro_rated_items opri)", oproduct.ID)

			if ierror != nil {
				log.Fatalln(ierror)
			}

			defer items.Close()

			/* ------------------------------------------------------------------- MAP PRODUCT ITEMS -------------------------------------------------------------- */

			for items.Next() {
				err := items.StructScan(&opitem)
				if err != nil {
					log.Fatalln(err)
				}

				/* ------------------------------------------------------------------- INIT SAVE PLAN ITEM -------------------------------------------------------------- */

				var id uuid.UUID

				sbsPeriod := 1

				if opitem.SubsPeriod != nil && *opitem.SubsPeriod != "" {
					intlVal, _ := strconv.Atoi(*opitem.SubsPeriod)
					sbsPeriod = intlVal
				}

				newItem := ManagePlansItem{
					ManagePlansID: newPlan.ID,
					ProductItemID: opitem.ItemID,
					AircraftID:    opitem.AircraftID,
					StartDate:     opitem.StartDate,
					EndDate:       opitem.EndDate,
					BlockingDate:  opitem.ExcludedIn,
					Status:        opitem.Status,
					SubsPeriod:    sbsPeriod,
					Quantity:      opitem.Quantity,
					IsAdmin:       isAdmin,
				}

				queryItems := `INSERT INTO manage_plans_items (manage_plans_id, product_item_id, aircraft_model_id, start_date, end_date, blocking_date, status, subscription_period, quantity, is_admin) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id`

				imp_error := db.QueryRow(queryItems, newPlan.ID, opitem.ItemID, opitem.AircraftID, opitem.StartDate, opitem.EndDate, opitem.ExcludedIn, opitem.Status, sbsPeriod, opitem.Quantity, isAdmin).Scan(&id)

				if imp_error != nil {
					log.Fatalln(imp_error.Error())
				}

				newItem.ID = id

				/* ----------------------------------------------------------------- FINISH SAVE PLAN ITEM -------------------------------------------------------------- */

				aircraft := AircraftItems{}

				models, aerror := db.Queryx("SELECT id, customer_aircraft_id FROM order_product_item_tail_numbers WHERE order_products_item_id = $1", opitem.ID)

				if aerror != nil {
					log.Fatalln(aerror)
				}

				defer models.Close()

				for models.Next() {
					err := models.StructScan(&aircraft)
					if err != nil {
						log.Fatalln(err)
					}

					var id uuid.UUID

					newAircraft := ManagePlansAircraft{
						Caircraft:         aircraft.Caircraft,
						ManagePlansItemID: newItem.ID,
					}

					queryAircraft := `INSERT INTO manage_plans_aircraft (manage_plans_item_id, aircraft_id) VALUES ($1, $2) RETURNING id`

					aim_error := db.QueryRow(queryAircraft, newItem.ID, aircraft.Caircraft).Scan(&id)

					if aim_error != nil {
						log.Fatalln(aim_error)
					}

					newAircraft.ID = id
				}
			}
		}

		log.Printf("Save successfuly order: %s", order.Onumber)
	}

	log.Println("Save process is finish!")
}
