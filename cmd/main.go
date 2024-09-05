package main

import (
	"fmt"
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

	readParams := false
	if readParams {
		pm := luxtronik.NewParameterMap()

		err = c.ReadParameters(pm)
		if err != nil {
			panic(err)
		}

		wk := pm[luxtronik.ParamHeatingTargetCorrection]
		hp, err := wk.ToHeatPump(-1.)
		if err != nil {
			panic(err)
		}
		err = c.WriteParameter(luxtronik.ParamHeatingTargetCorrection, hp)
		if err != nil {
			panic(err)
		}
	} else {
		pm := luxtronik.NewCalculationsMap()
		err = c.ReadCalculations(pm)
		if err != nil {
			panic(err)
		}

		d1, d2, d3, t := pm.GetDisplay()
		fmt.Printf("%s %s %s\n%s", d1, d2, t, d3)
	}
}
