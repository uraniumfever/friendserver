package main

import (
	"sync"
)

// user models an active user in the registry.
type user struct {
	id              int
	friends         map[int]bool
	friendSignin    chan int
	followerSignoff chan int
}

// userRegistry models a thread-safe registry of active users.
// The zero'd object is ready to use.
type userRegistry struct {
	byID map[int]*user
	mux  sync.Mutex
}

// add will inserts a new user into the registry and notify other users.
func (users *userRegistry) add(id int, friends []int) *user {

	u := &user{
		id:              id,
		friends:         make(map[int]bool),
		friendSignin:    make(chan int),
		followerSignoff: make(chan int),
	}

	// populate friend 'set' for efficient O(1) membership tests
	// (don't befriend oneself or our channels will deadlock)
	for _, f := range friends {
		if f != u.id {
			u.friends[f] = true
		}
	}

	// begin critical section to read/write the registry
	users.mux.Lock()
	defer users.mux.Unlock()

	// init map if not yet done so
	// means no need for constructor and zero'd struct is ready to rock 'n' roll
	if users.byID == nil {
		users.byID = make(map[int]*user)
	}

	// add to lookup
	users.byID[u.id] = u

	// [step 4 from instructions]
	// notify users that their friend has signed in.
	// Note1: Our user may not reciprocate the friendship so this is not as
	//        simple as just looking at *our* user's friends list!
	// Note2: This is a linear search for followers, we could maintain a list of
	//        followers per user but that would mean a lot more bookkeeping.

	for _, follower := range users.byID {
		if follower.friends[u.id] {
			follower.friendSignin <- u.id
		}
	}

	return u
}

// remove will delete the user from the registry and notify other users.
func (users *userRegistry) remove(u *user) {
	// begin critical section to read/write the registry
	users.mux.Lock()
	defer users.mux.Unlock()

	delete(users.byID, u.id)

	// [step 5 from instructions]
	// notify this user's friends about the sign-off of their follower.
	// Note: This is a different set of users than the ones we notified about
	//       the sign-in event, it is strange, but does seem to be what the code
	//       test is specifically asking for ¯\_(ツ)_/¯

	for fID := range u.friends {
		friend, exists := users.byID[fID]
		if exists {
			friend.followerSignoff <- u.id
		}
	}
}
