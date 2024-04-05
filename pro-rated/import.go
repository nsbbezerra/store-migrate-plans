package main

import (
	"log"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type ProRated struct {
	ID             uuid.UUID `db:"pro_rated_id"`
	OrderProductID uuid.UUID `db:"order_product_id"`
	OrderID        uuid.UUID `db:"order_id"`
	ProductID      uuid.UUID `db:"product_id"`
	IsAdmin        string    `db:"acceptance_status"`
}

type ProRatedItem struct {
	ID          uuid.UUID `db:"pro_rated_item_id"`
	ProRatedID  uuid.UUID `db:"pro_rated_id"`
	OrderItemID uuid.UUID `db:"order_product_item_id"`
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
	ProRated      bool       `db:"pro_rated"`
	ProRatedID    uuid.UUID  `db:"pro_rated_id"`
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
		log.Println("Database Connected")
	}

	pro_rated := ProRated{}

	pro_rateds, prerror := db.Queryx("select opr.id  as pro_rated_id, op.id as order_product_id, o.id as order_id, op.product_id, op.acceptance_status from optionals_pro_rated opr join order_products op on op.id = opr.order_product_id join orders o on o.id = op.order_id")

	if prerror != nil {
		log.Fatalln(prerror)
	}

	defer pro_rateds.Close()

	log.Println("Migration process init!")

	var isAdmin string

	if pro_rated.IsAdmin == "Support Test" {
		isAdmin = "yes"
	} else {
		isAdmin = "no"
	}

	for pro_rateds.Next() {
		err := pro_rateds.StructScan(&pro_rated)

		if err != nil {
			log.Fatalln(err)
		}

		pro_rated_item := ProRatedItem{}
		manage_plan := ManagePlans{}

		pro_rated_items, prierror := db.Queryx("select opri.id as pro_rated_item_id, opri.pro_rated_id, opri.order_product_item_id from optionals_pro_rated_items opri where opri.pro_rated_id = $1", pro_rated.ID)

		manage_plans, mperror := db.Queryx("select mp.id, mp.customer_code, mp.order_id, mp.product_id, mp.service_center from manage_plans mp where mp.order_id = $1 and mp.product_id = $2", pro_rated.OrderID, pro_rated.ProductID)

		if prierror != nil {
			log.Fatalln(prierror)
		}

		if mperror != nil {
			log.Fatalln(mperror)
		}

		defer pro_rated_items.Close()
		defer manage_plans.Close()

		var manage_plans_id uuid.UUID

		for manage_plans.Next() {
			err := manage_plans.StructScan(&manage_plan)
			if err != nil {
				log.Fatalln(err)
			}
			manage_plans_id = manage_plan.ID
		}

		for pro_rated_items.Next() {
			err := pro_rated_items.StructScan(&pro_rated_item)

			if err != nil {
				log.Fatalln(err)
			}

			order_product_item := OrderProductItem{}

			order_product_items, opierror := db.Queryx("select opi.id, opi.order_product_id, opi.aircraft_model_id, opi.item_id, opi.start_date, opi.end_date, opi.subscription_period, opi.excluded_in, opi.status, opi.quantity from order_products_items opi where opi.id = $1", pro_rated_item.OrderItemID)

			if opierror != nil {
				log.Fatalln(opierror)
			}

			defer order_product_items.Close()

			for order_product_items.Next() {
				err := order_product_items.StructScan(&order_product_item)

				if err != nil {
					log.Fatalln(err)
				}

				var id uuid.UUID

				sbsPeriod := 1

				if order_product_item.SubsPeriod != nil && *order_product_item.SubsPeriod != "" {
					intlVal, _ := strconv.Atoi(*order_product_item.SubsPeriod)
					sbsPeriod = intlVal
				}

				newItem := ManagePlansItem{
					ManagePlansID: manage_plans_id,
					ProductItemID: order_product_item.ItemID,
					AircraftID:    order_product_item.AircraftID,
					StartDate:     order_product_item.StartDate,
					EndDate:       order_product_item.EndDate,
					BlockingDate:  order_product_item.ExcludedIn,
					Status:        order_product_item.Status,
					SubsPeriod:    sbsPeriod,
					Quantity:      order_product_item.Quantity,
					ProRated:      true,
					ProRatedID:    pro_rated.ID,
					IsAdmin:       isAdmin,
				}

				queryItems := `INSERT INTO manage_plans_items (manage_plans_id, product_item_id, aircraft_model_id, start_date, end_date, blocking_date, status, subscription_period, quantity, pro_rated, pro_rated_id, is_admin) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12) RETURNING id`

				imp_error := db.QueryRow(queryItems, manage_plans_id, order_product_item.ItemID, order_product_item.AircraftID, order_product_item.StartDate, order_product_item.EndDate, order_product_item.ExcludedIn, order_product_item.Status, sbsPeriod, order_product_item.Quantity, true, pro_rated.ID, isAdmin).Scan(&id)

				if imp_error != nil {
					log.Fatalln(imp_error.Error())
				}

				newItem.ID = id

				aircraft := AircraftItems{}

				all_aircraft, aerror := db.Queryx("select opitn.id, opitn.customer_aircraft_id, opitn.order_products_item_id from order_product_item_tail_numbers opitn where opitn.order_products_item_id = $1", order_product_item.ID)

				if aerror != nil {
					log.Fatalln(aerror)
				}

				defer all_aircraft.Close()

				for all_aircraft.Next() {
					err := all_aircraft.StructScan(&aircraft)

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

		log.Printf("Save successfuly order: %s", pro_rated.ID)
	}

	log.Println("Save process is finish!")
}
