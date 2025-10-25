# Interface Inc - Messaging App (by Wagner Domingos da Silva Santos)

## Project Description
This repo contains the development of the classic 1:1 chat as a challenger for Interface Inc process

---
## About the Author
Wagner Domingos, former senior Software Engineer at Amazon, currently working at Bitso as Senior Backend Engineer with more than 20 years of experience and knowledge about software analysis, development and architecture, involving in particular Java and Golang languages and related frameworks, acting also with coaching and leadership (technical and behavioral) of development teams.
Solid business knowledge in different industries like Aerospacial Program, Crypto, E-Commerce, Financial Investments Management, Fiscal/Billing (CTe issue, taxation IFRS), Capitalization, GED, Phone Billing, Asset/Equipment Management and Maintenance, Digital Products/Marketing, Life Insurance and some other business areas related to the main segments of country economy like: mining, railway, port, oil and gas, fertilizer, bank, telephony, energy, portage, digital marketing and BSS Telecom. 

* [Linkedin] https://www.linkedin.com/in/wagner-domingos-da-silva-santos-981278100/
* [Email] wagnerdomingos@gmail.com

Apart from the IT life, in my spare time I'm a Dj and music composer, mainly regarding Drum and Bass and Breakbeat genres.  

```
#####\             _             /#####
#( )# |          _( )__         | #( )#
##### |         /_    /         | #####
#" "# |     ___m/I_ //_____     | #" "#
# O # |____#-x.\ /++m\ /.x-#____| # O #
#m.m# |   /" \ ///###\\\ / "\   | #m.m#
#####/    ######/     \######    \#####

```


## Project Structure

```
messaging-app/
├── go.mod                          # Go module definition and dependencies
├── main.go                         # Application entry point
├── app/                            
│   ├── app.go                      # Main application setup and routing
│   └── handlers.go                 # HTTP request handlers
├── domain/                         
│   ├── models.go                   # Domain entities and data structures
│   └── errors.go                   # Domain-specific errors and error types
├── repositories/                   
│   ├── user_repository.go          # User data storage and operations
│   ├── chat_repository.go          # Chat and message data storage
│   └── interfaces.go               # Repository contracts (abstractions)
├── services/                      
│   └── message_service.go          # Core messaging business logic
├── sockets/                        
│   ├── hub.go                      # WebSocket connection management
│   └── client.go                   # WebSocket client handling
└── main_test.go                    # End-to-end integration tests
```

## Running the Application

```bash 
Build and run the application
go build
./messaging-app

Or run directly with go run
go run main.go

# Run on a specific port
PORT=8080 go run main.go
```
The server will start on http://localhost:8080 (or the port you specified).

## Testing with curl Commands

### Create Users

* Create first user

``` bash
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"username": "alice"}'
``` 

* Create second user  

``` bash
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"username": "bob"}'
```

* Create third user

``` bash
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"username": "charlie"}'
```

* Expected Response:
``` json
{"id":"{UUID}","username":"alice","created_at":"2023-10-01T10:00:00Z"}
```


### Get User Information

Replace USER_ID with the actual UUID from the create response

``` bash
curl http://localhost:8080/api/v1/users/{USER_ID}
```

### Send Messages

- Alice sends message to Bob
``` bash
curl -X POST http://localhost:8080/api/v1/messages \
  -H "Content-Type: application/json" \
  -d '{
    "sender_id": "{ALICE_USER_ID}",
    "recipient_id": "{BOB_USER_ID}", 
    "content": "Hello Bob!",
    "idempotency_key": "msg1"
  }'
```
-  Bob replies to Alice
``` bash
curl -X POST http://localhost:8080/api/v1/messages \
  -H "Content-Type: application/json" \
  -d '{
    "sender_id": "{BOB_USER_ID}",
    "recipient_id": "{ALICE_USER_ID}",
    "content": "Hi Alice! How are you?",
    "idempotency_key": "msg2"
  }'
```
- Test idempotency - same idempotency key should return existing message

``` bash
curl -X POST http://localhost:8080/api/v1/messages \
  -H "Content-Type: application/json" \
  -d '{
    "sender_id": "{ALICE_USER_ID}",
    "recipient_id": "{BOB_USER_ID}",
    "content": "This should not create a new message",
    "idempotency_key": "msg1"
  }'
```

### List User Chats

- Get Alice's chats
``` bash
curl "http://localhost:8080/api/v1/chats?user_id={ALICE_USER_ID}&page=1&page_size=10"
```

- With pagination
``` bash
curl "http://localhost:8080/api/v1/chats?user_id={ALICE_USER_ID}&page=1&page_size=5"
```

### List Chat Messages

Get messages from a specific chat.
First get the chat ID from the chats response, then:

``` bash
curl "http://localhost:8080/api/v1/chats/CHAT_ID/messages?page=1&page_size=50"
```

### Health Check

``` bash
curl http://localhost:8080/health
```

## Testing WebSocket Connections

You'll need websocat or a similar WebSocket client for terminal testing:

### Test WebSocket Connections

- Connect as Alice (replace with actual user ID)
``` bash
websocat "ws://localhost:8080/ws?user_id={ALICE_USER_ID}"
```

- Connect as Bob in another terminal
``` bash
websocat "ws://localhost:8080/ws?user_id={BOB_USER_ID}"
```

### Test Real-time Messaging

1. Start the server

2. In Terminal 1, connect as Alice:

``` bash
websocat "ws://localhost:8080/ws?user_id={ALICE_USER_ID}"
```

3. In Terminal 2, connect as Bob:

``` bash
websocat "ws://localhost:8080/ws?user_id={BOB_USER_ID}"
```

4. In Terminal 3, send a message from Alice to Bob:

``` bash
curl -X POST http://localhost:8080/api/v1/messages \
  -H "Content-Type: application/json" \
  -d '{
    "sender_id": "{ALICE_USER_ID}",
    "recipient_id": "{BOB_USER_ID}",
    "content": "Hello via WebSocket!"
  }'
```

5. Watch Bob's terminal - he should receive the message instantly!

### Mark Messages as Read via WebSocket
 In the WebSocket terminal, send:

``` json
{"type": "mark_read", "message_id": "MESSAGE_ID"}
```

## Testing Edge Cases
- Send empty message
``` bash
curl -X POST http://localhost:8080/api/v1/messages \
  -H "Content-Type: application/json" \
  -d '{"sender_id": "id1", "recipient_id": "id2", "content": ""}'
```


- Send message to self

``` bash
curl -X POST http://localhost:8080/api/v1/messages \
  -H "Content-Type: application/json" \
  -d '{"sender_id": "id1", "recipient_id": "id1", "content": "Hello me"}'
```  

- Invalid user ID

``` bash
curl -X POST http://localhost:8080/api/v1/messages \
  -H "Content-Type: application/json" \
  -d '{"sender_id": "invalid", "recipient_id": "invalid", "content": "Hello"}'
```  