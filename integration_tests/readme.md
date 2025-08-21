# Running the integration tests manually

1. Start the Keycloak server:
```bash
docker-compose up -d 

```
2. When the service is ready, obtain the external token:
```bash
export TF_VAR_external_token=$(curl \
  -d "client_id=test_client" \
  -d "client_secret=test_client_secret" \
  -d "username=test_user" \
  -d "password=test_user_password" \
  -d "grant_type=password" \
  "http://localhost:8080/realms/master/protocol/openid-connect/token" | jq -r '.access_token') && echo $TF_VAR_external_token
  
```  
3. Apply the terraform configuration
```bash
DEBUG=1 TF_LOG=DEBUG terraform apply --auto-approve

```

4. Get the internal token:
```bash
export INTERNAL_TOKEN=$(terraform output -json test | jq --raw-output .token) && echo $INTERNAL_TOKEN

```

5. Run the following curl command to check to token
```bash
curl -v -X 'GET' 'http://localhost:8081/token/review' \
  -H "X-Original-URL: http://www.example.com/api/v1/resources" -H "X-Original-Method: GET" \
  -H "Authorization: Bearer $INTERNAL_TOKEN"
  
```

