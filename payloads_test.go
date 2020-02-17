package main

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

// failRW stubs a reader & writer that always return an err
type failRW struct{}

func (w *failRW) Read(p []byte) (n int, err error) {
	return 0, errors.New("failRW always errs on read")
}
func (w *failRW) Write(p []byte) (n int, err error) {
	return 0, errors.New("failRW always errs on write")
}

func TestReadSignInPayloadOnValidPayloads(t *testing.T) {
	tests := []struct {
		testname      string
		json          string
		expectUserID  int
		expectFriends []int
	}{
		{
			testname:      "with friends",
			json:          `{"user_id":1, "friends":[2, 3, 4]}`,
			expectUserID:  1,
			expectFriends: []int{2, 3, 4},
		},
		{
			testname:      "empty friends",
			json:          `{"user_id":1, "friends":[]}`,
			expectUserID:  1,
			expectFriends: []int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testname, func(t *testing.T) {
			r := strings.NewReader(tt.json)
			p, err := readSignInPayload(r)

			if err != nil {
				t.Fatal("Expected no error but got err: ", err)
			}

			if p == nil {
				t.Fatal("Expected populated payload but got nil")
			}

			if p.UserID == nil {
				t.Fatal("Expected user_id to be populated but it is nil")
			}

			if p.Friends == nil {
				t.Fatal("Expected friends to be populated but it is nil")
			}

			if *p.UserID != tt.expectUserID {
				t.Errorf("Expect UserID to be %v but got %v", tt.expectUserID, *p.UserID)
			}

			if !reflect.DeepEqual(p.Friends, tt.expectFriends) {
				t.Errorf("Expect Friends to be %v but got %v", tt.expectFriends, p.Friends)
			}
		})
	}
}

func TestReadSignInPayloadOnInvalidPayloads(t *testing.T) {
	tests := []struct {
		testname string
		json     string
	}{
		{
			testname: "null user",
			json:     `{"user_id":null, "friends":[2, 3, 4]}`,
		},
		{
			testname: "missing user",
			json:     `{"friends":[2, 3, 4]}`,
		},
		{
			testname: "null friends",
			json:     `{"user_id":1, "friends":null}`,
		},
		{
			testname: "missing friends",
			json:     `{"user_id":1}`,
		},
		{
			testname: "empty json",
			json:     `{}`,
		},
		{
			testname: "incompatible json",
			json:     `["a", "b"]`,
		},
		{
			testname: "invalid json",
			json:     `["a", b`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.testname, func(t *testing.T) {
			r := strings.NewReader(tt.json)
			p, err := readSignInPayload(r)

			if err == nil {
				t.Fatal("Expected an error but got payload:", p)
			}
		})
	}
}

func TestReadSignInPayloadOnFailingReader(t *testing.T) {
	p, err := readSignInPayload(&failRW{})
	if err == nil {
		t.Fatal("Expected an error but got payload:", p)
	}
}

func TestSendOnlineStatusPayload(t *testing.T) {
	tests := []struct {
		testname   string
		id         int
		online     bool
		expectJSON string
	}{
		{
			testname:   "is-online",
			id:         1,
			online:     true,
			expectJSON: `{"user_id":1,"online":true}`,
		},
		{
			testname:   "is-offline",
			id:         1,
			online:     false,
			expectJSON: `{"user_id":1,"online":false}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.testname, func(t *testing.T) {
			b := strings.Builder{}
			err := sendOnlineStatusPayload(&b, tt.id, tt.online)

			if err != nil {
				t.Fatal("Expected no error but got err: ", err)
			}

			// we don't care about the newlines
			got := strings.Trim(b.String(), "\n")

			if got != tt.expectJSON {
				t.Errorf("Expected JSON `%v` but got `%v`", tt.expectJSON, got)
			}
		})
	}
}

func TestSendOnlineStatusPayloadOnFailingWriter(t *testing.T) {
	err := sendOnlineStatusPayload(&failRW{}, 1, true)
	if err == nil {
		t.Fatal("Expected an error but none")
	}
}
