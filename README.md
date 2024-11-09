# Txn-Parser

Implement Ethereum blockchain parser that will allow to query transactions for subscribed addresses.

## Installation

**With Binary**
* Git clone this [repo](git@github.com:rovilay/Txn-Parser.git)
* Navigate to the directory
    ```
    cd Txn-Parser
    ```
* Run
    ```
    ./cmd/main
    ```

**With Go**
* Install [Go](https://go.dev/doc/install) if you haven't
* Git clone this [repo](git@github.com:rovilay/Txn-Parser.git)
* Navigate to the directory
    ```
    cd Txn-Parser
    ```
* Run the server
    ```
    go run main.go
    ```

## API Endpoints

The api server runs on port 8080.

Note: block parsing starts once there is at least one subscription

* **POST v1/subscribe** 
   * adds address to observer.
   
   payload
   ```
   { "address": "" }
   ```
   response
   ```
    {
        "message": "subscription successful"
    }
   ```

* **GET v1/blocks/latest** 
   * returns the lastest parsed block number by the service.
   
   response
   ```
    {
        "data": 21151134, // block number
        "message": "success"
    }
   ```
 
* **GET v1/transactions/{address}**
   * list of inbound or outbound transactions for an address
    ```
    {
        "data": [], // transactions
        "message": "success"
    }
    ```

