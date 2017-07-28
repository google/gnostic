# urlshortener sample client

## Steps to run:

1. Generate the OpenAPI 3.0 description using `disco2oas`.

        disco2oas --api=urlshortener --version=v1 --v3
	
2. (optional) View the JSON OpenAPI 3.0 description.

        gnostic urlshortener-v1.pb --json-out=-
	
3. Generate the urlshortener client.

        gnostic urlshortener-v1.pb --go-client-out=urlshortener
	
4. Build the client.

        go install 
	
5. Download `client_secrets.json` from the Google Cloud Developer Console.

6. Run the client

        urlshortener
	
