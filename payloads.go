package main

import (
	"encoding/json"
	"errors"
	"io"
	"log"
)

// userSignInPayload models the JSON sign-in request.
type userSignInPayload struct {
	UserID  *int  `json:"user_id"`
	Friends []int `json:"friends"`
}

// onlineStatusPayload models the JSON online status notification response.
type onlineStatusPayload struct {
	UserID int  `json:"user_id"`
	Online bool `json:"online"`
}

// readSignInPayload reads signin JSON from the reader.
func readSignInPayload(r io.Reader) (*userSignInPayload, error) {
	p := &userSignInPayload{}
	err := json.NewDecoder(r).Decode(p)

	if err != nil {
		log.Println("Unable to decode signin payload: ", err)
		return nil, err
	}

	if p.UserID == nil {
		return nil, errors.New("Required field 'user_id' not found")
	}

	if p.Friends == nil {
		return nil, errors.New("Required field 'friends' not found")
	}

	return p, nil
}

// sendOnlineStatusPayload writes status notifiation JSON to the writer.
func sendOnlineStatusPayload(w io.Writer, id int, online bool) error {
	p := &onlineStatusPayload{id, online}
	err := json.NewEncoder(w).Encode(p)

	if err != nil {
		log.Println("Unable to encode online status payload:", err)
		return err
	}

	return nil
}
