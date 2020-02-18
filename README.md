# Friend Server in Golang

This is a toy TCP server written for a challenge by [daveagill](https://github.com/daveagill).

## Usage
Start the server with one of these examples:

```bash
# listen on all local IPs, port 1337 (default)
$ go run .
```
```bash
# listen on all local IPs, port 8080
$ go run . :8080
```
```bash
# listen on 127.0.0.1, port 9999
$ go run . 127.0.0.1:9999
```


Start a TCP client (here using netcat) and send the signin JSON payload for user1:
```bash
$ nc localhost 1337
{"user_id":1, "friends":[2, 3, 4]}
```

Start a 2nd TCP client and send the signin JSON payload for user2:
```bash
$ nc localhost 1337
{"user_id":2, "friends":[7, 8, 9]}
```

User1 will have received a signin notification _about_ user2 because user1 follows user2 as a "friend":
```bash
{"user_id":2, "online":true}
```

> ‼️ Note this is a deliberate deviation from the payload schema given in the code challenge spec which did not include the user_id to tell us who the notification was about: `{"online": true}`

Disconnect user1 by sending an `EOF` (press ctrl-d). This will close the netcat session.

User2 should receive a notification about user1 signing off:
```bash
{"user_id":1, "online":false}
```
> ‼️ User2 receives this notification because user1 lists them as a friend. User2 did not reciprocate that friendship however. This feels like a strange UX to me but it is difficult to interpret the code challenge spec in any other way 🙃🤷
> I.e. Steps 4 and 5 in the spec describe different (and possibly non-symmetric) approaches to determine the set users to notify about logins vs logoffs and this has been faithfully implemented here.


## Questions & Answers

### __Q1:__ What changes if we switch TCP to UDP?

UDP is a connectionless protocol. This has a some consequences for us:
1. Notifying clients of signons and signoffs requires the server to send datagram to clients. The server would need client network details to do that - One option is to read the source address and source port from the UDP/IP headers and send packets back to there.
2. There are no guarantees that packets are received. This could mean the server never sees clients signing in or out, it could also mean that users are not notified about their friends and followers.
3. There are no guarantees about exactly-once delivery. This means requests may be received multiple times. Our server and clients would need to be able to cope with this.
4. Messages may be receieved out-of-order. This could mean users receive a friend signoff notification _before_ they receive the signon notification.
5. No concept of KeepAlives in UDP, so we would have to implement heatbeat messages as part of an application-level protocol.

### __Q2:__ How would you detect the user (connection) is still online?

__In TCP:__ The solution implemented will `Read()` the connection waiting for users to hangup by sending an `EOF`. To detect a disorderly disconnect I configured KeepAlives to send repeat probe packets, once enough KA packets fail the connection gets dropped automatically, this causes the active `Read()` to time out and subsequent ones will return `EOF`.

__In UDP:__ I would implement a 'logoff' JSON payload so that users can purposefully log out. I would also have a custom KeepAlive-like system where clients would send periodic heartbeat messages to the server; if the client goes quiet for too long we consider them offline.

### __Q3:__ What happens if the user has a lot of friends?

They go to a lot of parties! 🤘

Also: More friends means more load on the notification system.

At the transport level this could mean we exhaust the OS's pool of ephemeral ports, plus they can take some time to return to active duty.

On the network this could mean sudden traffic spikes as the user logs in and out.

At the application level the server must determine 'who' to notify and we can consider how that scales...

* For sign-off notification this is simply linear, `O(f)`, for the number of friends `f`.
* For sign-on notifications this is linear in terms of friends `f` and total number of users `t`. This was achieved by building a hash-set (implemented as a `map[int]bool`) of friend IDs per user to get efficient `O(1)` tests of friendship between two users. Building the set requires `O(f)` time and determing which users to notify is achieved with a linear scan over the user-registry, costing `O(t)` time, which gives our total average cost: `O(f+t)`.
    > In theory this cost could be optimised further to be linear in terms of _just_ the number of followers and avoid the scan over the entire userbase.
    > This would require some fancier bookkeeping and would be easier to implement in a system that knew about users even before their first login (e.g. require an initial signup) then we could simply record 'our' user as a "follower" onto each of its friends on signup.
    >With the current approach a user's friend may not even be in the registry yet if they have never signed in, so that complicates things a bit more.

### __Q4:__ How design of your application will change if there will be more than one instance of the server.

The main change is we cannot store the state of active users in local memory. This is because each node in the cluster would have a different view of which users are on/offline based on the requests they happen to have serviced.

Two common solutions to this are:
* Replicate state across nodes in the cluster.
* Centralise state in a DB or cache that is accessed by all nodes in the cluster.

I think a good option would be to centralise the state into Redis. Use it store the registry of active users and replace the Golang channels (used in my solution) with Redis [Pub/Sub channels](https://redis.io/topics/pubsub). This refactoring could be done by introducing an `interface` type in [users.go](users.go) to abstract the `userRegistry` and then introduce a new Redis-based implementation: `redisUserRegistry`.

Finally, to direct and distribute traffic across the nodes in our cluster we would need a Load Balancer infront such as nginx or AWS ELB. This then becomes the real entry-point for clients to connect to.
