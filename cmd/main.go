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

	readParams := true
	if readParams {
		pm := luxtronik.NewParameterMap()

		err = c.ReadParameters(pm)
		if err != nil {
			panic(err)
		}

		if err := setLockTime(pm, c, luxtronik.ParamWaterLockTime1From, "06:00"); err != nil {
			panic(err)
		}
		if err := setLockTime(pm, c, luxtronik.ParamWaterLockTime1To, "00:00"); err != nil {
			panic(err)
		}
		if err := setLockTime(pm, c, luxtronik.ParamWaterLockTime2From, "00:00"); err != nil {
			panic(err)
		}
		if err := setLockTime(pm, c, luxtronik.ParamWaterLockTime2To, "01:00"); err != nil {
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

func setLockTime(pm luxtronik.ParameterMap, c luxtronik.Client, param int32, value string) error {
	wk := pm[param]
	hp, err := wk.ToHeatPump(value)
	if err != nil {
		return err
	}
	return c.WriteParameter(param, hp)
}
