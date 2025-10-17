package main

import (
	"github.com/ootiny/capi/server/runtime"
	"github.com/ootiny/capi/server/runtime/api_system_city"
	"github.com/ootiny/capi/server/runtime/db_city"
)

func init() {
	api_system_city.OnCreate(
		func(ctx *runtime.Context, city db_city.Create) (db_city.Default, *runtime.Error) {
			return db_city.Create(city)
		})

	api_system_city.OnDelete(
		func(ctx *runtime.Context, city db_city.Delete) (db_city.Default, *runtime.Error) {
			return db_city.Delete(id)
		})

	api_system_city.OnUpdate(
		func(ctx *runtime.Context, data db_city.Update) (db_city.Default, *runtime.Error) {
			return db_city.Update(ctx, data)
		})

	api_system_city.OnQuery(
		func(ctx *runtime.Context, query db_city.Query) (api_system_city.CityList, *runtime.Error) {
			return db_city.QueryFull(ctx, query)
		})
}

func main() {
	if err := runtime.NewHttpServer("0.0.0.0:8080", "", "", true).Run(); err != nil {
		panic(err)
	}
}
