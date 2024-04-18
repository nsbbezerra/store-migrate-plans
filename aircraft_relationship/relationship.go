package main

import (
	"log"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type CustomerAircraft struct {
	ID                uuid.UUID `db:"id"`
	AircraftName      *string   `db:"aircraft_name"`
	OwnerCode         *string   `db:"owner_code"`
	OperatorCode      *string   `db:"operator_code"`
	ModelCode         *string   `db:"model_code"`
	SerialNumber      string    `db:"serial_number"`
	Status            *string   `db:"status"`
	TailNumber        *string   `db:"tail_number"`
	LifeCycle         *string   `db:"life_cycle"`
	OtherOperator     string    `db:"other_operator"`
	OtherOperatorType string    `db:"other_operator_type"`
}

type AircraftRelationship struct {
	ID               uuid.UUID `db:"id"`
	SerialNumber     string    `db:"serial_number"`
	CustomerCode     string    `db:"customer_code"`
	SalesforceId     *string   `db:"salesforce_id"`
	RelationshipType string    `db:"relationship_type"`
}

func main() {
	db, err := sqlx.Connect("postgres", "host=localhost port=5432 user=svq password=svq dbname=services_quotation sslmode=disable")

	if err != nil {
		log.Fatalln(err)
	}

	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	} else {
		log.Println("Successfully Connected")
	}

	customerAircraft := CustomerAircraft{}

	customersAircraft, caerror := db.Queryx("SELECT ca.id, ca.aircraft_name, ca.owner_code, ca.operator_code, ca.other_operator, ca.other_operator_type, ca.model_code, ca.tail_number, ca.serial_number, ca.life_cycle, ca.status FROM customers_aircrafts ca WHERE ca.other_operator IS NOT NULL")

	if caerror != nil {
		log.Fatalln(caerror)
	}

	defer customersAircraft.Close()

	log.Println("Save process is init!")

	for customersAircraft.Next() {
		err := customersAircraft.StructScan(&customerAircraft)
		if err != nil {
			log.Fatalln(err)
		}

		newRelationshipt := AircraftRelationship{
			SerialNumber:     customerAircraft.SerialNumber,
			CustomerCode:     customerAircraft.OtherOperator,
			RelationshipType: customerAircraft.OtherOperatorType,
		}

		var id uuid.UUID

		queryRelationship := `INSERT INTO aircraft_relationship (serial_number, customer_code, relationship_type) VALUES ($1, $2, $3) RETURNING id`

		rlerror := db.QueryRow(queryRelationship, customerAircraft.SerialNumber, customerAircraft.OtherOperator, customerAircraft.OtherOperatorType).Scan(&id)

		if rlerror != nil {
			log.Fatalln(rlerror)
		}

		newRelationshipt.ID = id

		log.Printf("Save successfuly relationship of aircraft: %s", customerAircraft.SerialNumber)
	}

	log.Println("Save process is finish!")
}
