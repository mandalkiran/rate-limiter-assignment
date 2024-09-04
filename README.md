An implementation of token bucket algorithm to handle rate limits for multiple endpoints and users.

### Running the application

```bash
./run_app.sh
```

### Update the rate-limiting rule
Pass the new limit as query param limit as show below
```bash
curl http://localhost:8080/update/rates?limit=60
```

### Run Test
```bash
go test -v
```

### Test Case

Testing the load for a ```user/:id/data``` which has limit of 5 request per minute.
```bash
for i in {1..6}; do curl http://localhost:8080/user/1/data; done
```
You can see the last response is 429 (too many request) 
```json
{"status":"Request Failed","body":"The API is at capacity, try again later."}
```