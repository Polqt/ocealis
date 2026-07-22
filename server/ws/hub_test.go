package ws

import "testing"

func TestBroadcastTopicOnlyReachesSubscribers(t *testing.T) {
	hub := NewHub()
	interested := NewClient(hub, nil, nil)
	other := NewClient(hub, nil, nil)

	hub.Register(interested)
	hub.Register(other)
	hub.Subscribe(interested, "ocean:all")

	msg := []byte(`{"type":"bottle_drift"}`)
	hub.BroadcastTopic("ocean:all", msg)

	select {
	case got := <-interested.send:
		if string(got) != string(msg) {
			t.Fatalf("unexpected payload: %s", got)
		}
	default:
		t.Fatal("expected ocean:all subscriber to receive message")
	}

	select {
	case <-other.send:
		t.Fatal("unsubscribed client should not receive topic message")
	default:
	}
}

func TestValidateTopicAllowsOceanAll(t *testing.T) {
	if err := validateTopic("ocean:all"); err != nil {
		t.Fatalf("ocean:all should be valid: %v", err)
	}
	if err := validateTopic("region:north_atlantic"); err != nil {
		t.Fatalf("region topic should be valid: %v", err)
	}
	if err := validateTopic("bottle:12"); err != nil {
		t.Fatalf("bottle topic should be valid: %v", err)
	}
	if err := validateTopic("nope"); err == nil {
		t.Fatal("expected invalid topic error")
	}
}
