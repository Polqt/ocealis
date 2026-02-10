package models

type Bottle struct {
	id 	   				int64
	sender_id   		int64
	message_test 		string
	bottle_style 		int
	start_lat  			float64
	start_lng  			float64
	hops 	   			int
	scheduled_release 	int64
	is_release			bool
	created_at   		int64
}
