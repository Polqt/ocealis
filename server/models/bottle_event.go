package models

type BottleEvent struct {
	id         	int64
	bottle_id  	int64
	event_type 	string
	lat        	float64
	lgn        	float64
	created_at 	int64
}
