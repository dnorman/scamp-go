SCAMP Go Lang Edition
=====================

The `scamp` package provides all the facilities necessary for participating in a SCAMP environment:

  * Parsing the discovery cache and building a directory of available services
  * Parsing packets streams
  * Parsing and verifying messages

Architecture
--------

Remote services are invoked by first establishing a `Connection` (TLS under the hood). This `Connection` can then be used to spawn `Sessions` which send a `Request` and block until a `Reply` is provided.

Usage
-----

	func main() {
		scamp.Initialize()

		conn := new(scamp.Connection)
		err := conn.Connect("127.0.0.1:30100")
		defer conn.Close()

		if err != nil {
			scamp.Error.Printf("could not connect! `%s`\n", err)
			return
		}

		request := scamp.Request{
			Action:         "helloworld.hello",
			envelopeFormat: scamp.ENVELOPE_JSON,
			Version:        1,
		}
		conn.SendRequest(request)
		reply, err := conn.RecvReply()
		if err != nil {
			scamp.Error.Printf("error receving reply: `%s`", err)
		}
		scamp.Info.Printf("got reply: `%s`", reply)
	}

Running the test suite
----------------------

  export GOPATH=$PWD
  go test scamp

Documentation
-------------

	export GOPATH=$PWD
	godoc -http=:6060
	# open http://localhost:6060/

Tasking
-------

Features

 - [x] Setup Go project
 - [x] TLS session setup
 - [x] cert verification
 - [x] Parse packet
 - [x] Generate packet
 - [x] Generate request
   - [x] Generate request header JSON
 - [x] Parse reply
 - [x] Parse request
   - [x] Route to action based on header JSON
 - [x] Generate reply
 - [ ] Verify TLS certificate with `/etc/authorized_services`
 - [x] Manage connection msgno
 - [x] Parse service cache
 - [ ] Route RPC based on service cache
 - [x] Use go logging library
 - [] AuthZ service support
   - [x] Ticket parsing
   - [x] Ticket verification
 - [ ] Chunk body to 128k
 - [ ] Reconnect logic (`scamp.Connection` connects with exponential backoff)
  - [ ] What to do if connection goes down during `Session` exchange?

Important Restructuring

 - [ ] Stream messages bodies
   - [ ] Session stream interface? `Reader`/`Writer` for bytes? Benefit: integration with patterns/helpers in (io lib)[http://golang.org/pkg/io]
 - [ ] Unify concepts of `Request`/`Reply` with `Message` and move that distinction to the direction of the `Session`
   - [ ] Rewrite `Request`/`Reply` code to reuse `session` `Reader`/`Writer` under the hood
 
Rad/Cool Ideas

 - [ ] Ragel state machine specification to generate go code

Bugs

 - [ ] Fix bug where sending envelope type `JSON` fails silently (should at least emit 'unknown type' to STDERR)
 - [ ] Fix bug where header `"type": "request"` fails silently (should at least emit 'unknown type' to STDERR)
 - [ ] Fix reference to documentation `message_id` which should read `request_id`
 - [ ] Move to interface design. Message parts which implement `Packet` so we can specialize `Header` vs `Data` which have different bodies from different data types.