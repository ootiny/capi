package main

import (
	"github.com/ootiny/capi/server/runtime"
	"github.com/ootiny/capi/server/runtime/api_system_city"
	"github.com/ootiny/capi/server/runtime/db_city"
)

func init() {
	api_system_city.OnAddCity(
		func(ctx *runtime.Context, city db_city.Default) (db_city.Default, *runtime.Error) {
			return city, nil
		})

	api_system_city.OnGetCityList(
		func(ctx *runtime.Context, country string) (api_system_city.CityList, *runtime.Error) {
			return db_city.QueryFull(ctx, runtime.SqlWhere{})
		})
}

func main() {
	if err := runtime.NewHttpServer("0.0.0.0:8080", "", "", true).Run(); err != nil {
		panic(err)
	}
}
