# Chirpy

## Endpoints

GET /api/healthz
GET /admin/metrics
POST /admin/reset
	GET /api/chirps
	POST /api/users", config.handlerCreateUser)
	POST /api/chirps", config.handlerValidate)
	GET /api/chirps/{chirpID}", config.handlerGetChirp)
	POST /api/login", config.handlerLogin)
	POST /api/refresh", config.handlerRefresh)
	POST /api/revoke", config.handlerRevoke)
	PUT /api/users", config.handlerUpdateUser)
	DELETE /api/chirps/{chirpID}", config.handlerDeleteChirp)
	POST /api/polka/webhooks"