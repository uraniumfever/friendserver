package main

import (
	"io"
	"log"
	"net"
	"time"
)

func startServer(address string) {
	log.Println("Starting server...")

	// configure TCP KeepAlive period to something small so we get read timeouts
	// quickly for disorderly disconnects. This is obviously a tradeoff against
	// how robust the connection is too poor comms.
	// Note: MacOS attempts 8 KA packets (by default), other OS's may differ.
	conf := net.ListenConfig{
		KeepAlive: 5 * time.Second,
	}

	// [Step 1 from instructions]
	// Start listening for tcp connections
	ln, err := conf.Listen(nil, "tcp", address)
	if err != nil {
		log.Fatalln("Unable to start server, will shutdown:", err)
	}
	defer ln.Close()

	log.Println("Listening on address:", address)

	// create the user registry for storing active users and dispatching events
	users := userRegistry{}

	// [Step 2 from instructions]
	// Accept connections indefinitely
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("Error accepting connection:", err)
			continue
		}

		log.Println("Connection accepted")
		go handleConnection(conn, &users)
	}
}

// handleConnection manages a user's long-running connected session.
// User is considered 'online' while the connection is held.
func handleConnection(conn net.Conn, users *userRegistry) {
	defer conn.Close()

	// [Step 3a from instructions]
	// read the JSON payload
	p, err := readSignInPayload(conn)
	if err != nil {
		log.Println("Sign-in attempt failed")
		return
	}

	// [Step 3b from instructions]
	// add our user to the registry (and remove again on signoff)
	log.Println("User", *p.UserID, "signed in, claims friends:", p.Friends)
	u := users.add(*p.UserID, p.Friends)
	defer users.remove(u)

	// a signal channel, closing it tells our goroutine that the user signed off
	bye := make(chan struct{})
	defer close(bye)

	// [steps 4 & 5 from instructions]
	// watch asynchronously for signing in/out events
	go func() {
		for {
			select {
			case friendID := <-u.friendSignin:
				log.Printf("Notify %v that %v signed in", u.id, friendID)
				sendOnlineStatusPayload(conn, friendID, true)
			case followerID := <-u.followerSignoff:
				log.Printf("Notify %v that %v signed off", u.id, followerID)
				sendOnlineStatusPayload(conn, followerID, false)
			case <-bye:
				log.Printf("User %v signed out", u.id)
				return
			}
		}
	}()

	blockTilDisconnect(conn)
}

// blockTilDisconnect reads from the connection until an EOF is received.
// For an orderly disconnect an EOF is received immediately.
// For a disorderly disconnect an EOF is received subsequent to a read timeout
// the duration depends on the KeepAlive period and attempted retries.
func blockTilDisconnect(conn net.Conn) {
	for {
		_, err := conn.Read([]byte{0})
		if err != nil {
			if err == io.EOF {
				log.Println("Detected disconnect")
				return
			}

			log.Println("Error reading from stream", err)
		}
	}
}
