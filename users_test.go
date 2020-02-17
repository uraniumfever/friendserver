package main

import (
	"testing"
)

func TestAddCreatesValidUser(t *testing.T) {
	id := 123
	friends := []int{2, 3, 4}

	users := userRegistry{}
	u := users.add(id, friends)

	if len(users.byID) != 1 {
		t.Fatal("Expected one item in registry, len=", len(users.byID))
	}

	if u == nil {
		t.Fatal("User was nil")
	}

	if u.id != id {
		t.Errorf("Expected ID to be %v but got %v", id, u.id)
	}

	if len(u.friends) != len(friends) {
		t.Errorf("Expected %v friends but got %v", len(friends), len(u.friends))
	}

	for _, fID := range friends {
		if !u.friends[fID] {
			t.Errorf("Expected friend ID %v is absent from friends", fID)
		}
	}
}

func TestRemoveWhenEmptyDoesNothing(t *testing.T) {
	users := userRegistry{}
	users.remove(&user{})

	if len(users.byID) != 0 {
		t.Fatal("Empty registry should stay empty, len=", len(users.byID))
	}
}

func TestAddThenRemoveLeavesRegistryEmpty(t *testing.T) {
	users := userRegistry{}
	u := users.add(1, []int{})
	users.remove(u)

	if len(users.byID) != 0 {
		t.Fatal("Expected to leave registry empty, len=", len(users.byID))
	}
}

func TestAddNotifiesFollowersOfSignon(t *testing.T) {
	users := userRegistry{}

	// user1 follows user2
	u1 := users.add(1, []int{2})

	// adding user2 will block so run it async
	go func() {
		users.add(2, []int{})

		// close user1's channel to unblock the test thread
		close(u1.friendSignin)
	}()

	// user1 should be notified about user2 coming online
	// (this pertains to step 4 in the code test specification)
	count := 0
	for friendID := range u1.friendSignin {
		count++
		if friendID != 2 {
			t.Fatalf("Expected to be notified of user 2 but got %v", friendID)
		}
	}

	if count != 1 {
		t.Errorf("Expected 1 signon notification but got %v", count)
	}
}

func TestRemoveNotifiesFriendsOfSignoff(t *testing.T) {
	users := userRegistry{}

	// user1 follows user2
	u1 := users.add(1, []int{2})

	// adding user2 will block so run it async
	var u2 *user
	go func() {
		u2 = users.add(2, []int{})

		// close user1's signin channel to unblock the test thread
		close(u1.friendSignin)
	}()

	// wait for channel to close so we know user2 is fully added
	for range u1.friendSignin {
	}

	// removing user1 will block so run it async
	go func() {
		users.remove(u1)

		// close user2's follower signout channel to unblock the test thread
		close(u2.followerSignoff)
	}()

	// user2 should be notified of user1 going offline because user2 is listed
	// as a "friend" by user1
	// (this pertains to step 5 in the code test specification)
	count := 0
	for friendID := range u2.followerSignoff {
		count++
		if friendID != 1 {
			t.Fatalf("Expected to be notified of user 1 but got %v", friendID)
		}
	}

	if count != 1 {
		t.Errorf("Expected 1 signoff notification but got %v", count)
	}
}
