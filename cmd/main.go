package main

import (
	"github.com/bakito/luxtronik"
)

func main() {
	heatPumpIP := "192.168.2.250" + ":" + luxtronik.DefaultPort
	c := luxtronik.MustNewClient(heatPumpIP, luxtronik.Options{
		SafeMode: true,
	})

	err := c.Connect()
	if err != nil {
		panic(err)
	}
	defer c.Close()

	pm := luxtronik.NewParameterMap()

	err = c.ReadParameters(pm)
	if err != nil {
		panic(err)
	}

	wk := pm[luxtronik.IdEinstWkAkt]
	hp, err := wk.ToHeatPump(-1.)
	if err != nil {
		panic(err)
	}
	err = c.WriteParameter(luxtronik.IdEinstWkAkt, hp)
	if err != nil {
		panic(err)
	}

}
