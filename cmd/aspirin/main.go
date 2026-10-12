package main

import "aspirin_back/internal/api"

// @title Aspirin Production API
// @version 1.0
// @description Управление этапами синтеза аспирина. Для защищённых методов укажите Bearer JWT в Authorize.
// @BasePath /api
// @securityDefinitions.apikey AspirinProductionBearer
// @in header
// @name Authorization
// @description Введите Bearer и JWT, полученный через POST /api/authentication.
func main() {
	api.StartServer()
}
