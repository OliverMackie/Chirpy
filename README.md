# Chirpy

## Endpoints

GET /api/healthz takes no parameters, returns 200 OK when the server is running.  
GET /admin/metrics takes no parameters, returns a page showing how many times the site has been accessed since last reset.  
POST /admin/reset takes no parameters and resets both the metrics and database. (to be used for development).  
GET /api/chirps has two optional query parameters, (sort: default "asc" or "desc") and (author_id: a valid user_id), returns all chirps in newest first with no params. Returns all chirps for one user oldest first if sort=desc and author_id is set.  
POST /api/users takes a json object with two string fields: email and password, returns 201 and a json user object. Return fields: id, created_at, updated_at, email, is_chirpy_red.  
POST /api/login takes the same inputs as POST /api/users. Responds with the same information plus an access token and a refresh token as string.
POST /api/chirps takes a json object with two string parameters: body and user_id. The authorization header should be set to " Bearer {token}" where {token} is replaced with a valid acess token. Responds with 201 and a chirp object: containing string parameters: id, created_at, updated_at, body and user_id.  
GET /api/chirps/{chirpID} takes a valid chirp_id in the path. Returns a chirp object, same as POST /api/chirps.
POST /api/refresh takes no parameters but the Authorization header must be set to " Bearer {refresh token}" where {refresh token} is a valid refresh token. Returns 200 and a new access token if the refesh token is valid otherwise returns 401.  
POST /api/revoke takes the same input as POST /api/refresh but revokes the refresh token, so that it is no longer valid. Returns 204 if successful.  
PUT /api/users takes a json object containing two string fields: email and password. The Authorization header should be set to " Bearer {token}" where {token} is replaced with a valid acess token. This endpoint alters the saved email and password for the logged in user. Returns 200 and a user json object (see Post /api/users).  
DELETE /api/chirps/{chirpID} takes a valid chirp_id in the path and a valid access token in the Authorization header (see POST /api/chirps). Returns 204 and DELETES the chirp object associated wtih the chirp_id.  
POST /api/polka/webhooks takes a json object with a event field containing "user.upgraded" and a data field containing a struct with a user_id field. The Authorization header should be set to " APIKey {key}" where the key is a valid polka API key. Returns 204 and sets the users "is_chirpy_red" field to true.  

### Examples

Examples can be found in the requests.http file, the examples assume that the server is running on localhost:8080.  