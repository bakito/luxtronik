package luxtronik

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"testing"
	"text/tabwriter"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration_Client(t *testing.T) {
	heatPumpIP := os.Getenv("HEATPUMP_IP")
	if heatPumpIP == "" {
		heatPumpIP = "192.168.2.250" + ":" + DefaultPort
	}

	runTest := func(pm DataTypeMap, readFromNet func(Client) error) func(t *testing.T) {
		return func(t *testing.T) {
			c := MustNewClient(heatPumpIP, Options{
				SafeMode: true,
			})

			require.NoError(t, c.Connect())
			defer func() {
				assert.NoError(t, c.Close())
			}()

			require.NoError(t, readFromNet(c))

			tw := tabwriter.NewWriter(os.Stdout, 12, 1, 1, ' ', 0)
			printFn := func(w io.Writer) func(i int32, p *Base) {
				return func(i int32, p *Base) {

					fmt.Fprintf(
						w,
						"Number: %d\tName: %s\tType: %s\tValue: %v\tUnit: %s\n",
						i,
						p.luxtronikName,
						p.class,
						checkStringer(p.FromHeatPump()),
						p.unit,
					)
				}
			}
			pm.IterateSorted(printFn(tw))
			require.NoError(t, tw.Flush())
		}
	}

	pm := NewParameterMap()
	t.Run("Parameter", runTest(DataTypeMap(pm), func(c Client) error {
		return c.ReadParameters(pm)
	}))
	vm := NewVisibilitiesMap()
	t.Run("Visibilities", runTest(DataTypeMap(vm), func(c Client) error {
		return c.ReadVisibilities(vm)
	}))
	cm := NewCalculationsMap()
	t.Run("Calculations", runTest(DataTypeMap(cm), func(c Client) error {
		return c.ReadCalculations(cm)
	}))
}

func TestIntegration_Refreshed_Calculations(t *testing.T) {
	c := MustNewClient("192.168.0.121:"+DefaultPort, Options{
		SafeMode: true,
	})

	require.NoError(t, c.Connect())
	defer func() {
		assert.NoError(t, c.Close())
	}()

	pm := NewCalculationsMap()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	tkr := time.NewTicker(3 * time.Second)

	for {

		require.NoError(t, c.ReadCalculations(pm))

		tw := tabwriter.NewWriter(os.Stdout, 12, 1, 1, ' ', 0)

		DataTypeMap(pm).IterateSorted(func(i int32, p *Base) {
			if p.rawValue == 0 || !p.HasChanges() {
				return
			}

			fmt.Fprintf(
				tw,
				"Number: %d\tName: %s\tType: %s\tValue: %v\tUnit: %s\n",
				i,
				p.luxtronikName,
				p.class,
				checkStringer(p.FromHeatPump()),
				p.unit,
			)
		})

		require.NoError(t, tw.Flush())
		select {
		case <-sigChan:
			return
		case tm := <-tkr.C:
			println(tm.Format(time.DateTime), strings.Repeat("=", 200))
			continue
		}
	}
}

func checkStringer(v any) any {
	if s, ok := v.(fmt.Stringer); ok {
		return s.String()
	}
	switch tv := v.(type) {
	case float32:
		return fmt.Sprintf("%.3f", tv)
	default:
		return v
	}
}
