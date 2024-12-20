## **Architecture Overview**

1. **API Endpoint Calls** (`/deposit` or `/withdrawal`):

   - When someone calls `/deposit` or `/withdrawal`, it create a **transaction** in the database with a **`Pending`** status.

2. **Webhook Callback**:

   - The **gateway provider** processes the request and sends a webhook to our system with a **transaction ID** (`txId`) and the updated status (e.g., `Completed`, `Failed`, etc.).
   - This webhook updates the transaction status in our database.

3. **Circuit Breaker for Reliability**:
   - To ensure reliable processing of messages from Kafka, we use a **circuit breaker** around `PublishWithCircuitBreaker` for `StartWebhookProcessing`.
   - This helps prevent cascading failures by halting processing when repeated errors occur and retrying later when the system stabilizes.

---

## **Proposed Architecture Flow**

### **Endpoints**

Currently we support both XML and JSON as requested by the readme.
Api doc is generated using SWAG.

Swagger URL: http://localhost:8080/api/v1/swagger/index.html

Whenever server is started, swagger docs will start with it. Port if not changed, will be `8080`

- Payments Routes ->

  ```sh
  curl --location 'localhost:8080/api/v1/payments/Deposit' \
  --header 'Content-Type: application/json' \
  --data '{
      "amount": 100.00,
      "user_id": 1,
      "currency": "USD",
      "country_id": 3
  }'
  ```

  ```sh
  curl --location 'localhost:8080/api/v1/payments/withdraw' \
  --header 'Content-Type: application/json' \
  --data '{
      "amount": 100.00,
      "user_id": 1,
      "currency": "USD",
      "country_id": 3
  }'
  ```

- Webhooks Route ->
  ```sh
  curl --location 'localhost:8080/api/v1/webhooks/' \
  --header 'Content-Type: application/json' \
  --data '{
      "txn_id": 35,
      "status": "SUCCESS",
      "updated_at": "2024-12-18T11:58:52.283721968Z"
  }'
  ```

---

## **Best Practices**

1. **Database Consistency**:

   - Always **persist transactions in a `Pending` state** before publishing to Kafka.
   - Ensure **idempotent updates** when webhook calls arrive (e.g., don't double-update the status using TX id, we send this tx id to the gateway provider which helps us with the idempotency).
   - For more robustness, we can even use a `clientID` generated by Fe and share that with our gateway providers.

2. **Circuit Breaker Settings**:

   - Configured the circuit breaker to trip after a certain number of consecutive failures (`MaxRequests`).
   - Used appropriate `Timeout` and `Interval` values based on our system's tolerance for downtime.

3. **Logging & Monitoring**:

   - Log key events and failures for **troubleshooting and monitoring**.
   - Suggestion: Integration with a monitoring tool (e.g., Prometheus, ELK) to track circuit breaker states and Kafka message delivery rates will be required in production.

4. **Retry Logic**:

   - Implemented **retries** with exponential backoff for publishing failures.
   - We might want to use a **dead-letter queue (DLQ)** for failed Kafka messages that need manual intervention.

5. **Graceful Shutdown**:
   - Server is handled gracefully to ensure smooth processing and fair resource usage and free up after they're consumed.

---

## **How to run the server**

1. The server can be run easily with the help of `docker`.
2. `make all` This will generate swagger files, build the code, and run the server
3. `make test` will run all tests
4. To generate only swagger test files, `make swagger`.
   Remember: `go install github.com/swaggo/swag/cmd/swag@latest` will be needed to install swaggo if not already installed. After installing, run `make swagger`. If `swag` `command not found` error is visible on the terminal, run `export PATH=$PATH:$(go env GOPATH)/bin`. Swagger files are already committed, running the server should render the swagger UI.

## **QnA**

### **Gateway**:

Current Code supports gateway fetching based on Currency and CountryID. There are countries that allow multiple currencies, for now, we don't support that as the provided model already had a unique constraint, in future if needed, we can simply remove the unique constraint and we will be able to support multiple currencies for a single country.

Also, If multiple gateways exist for a currency and country pair, we pick the latest gateway. We can customize it however we like. We might be able to assign priority in case one of the gateways is having an issue, so we can route traffic to other gateways! For the scope of this project, it's simply getting the latest gateway.

We've assumed, user_id is coming from request body, in production env, user_id will be coming from JWT headers.

Usage of mask and unmask data is limited, as there is `no PII` data of users that is being logged.
