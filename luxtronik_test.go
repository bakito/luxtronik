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
)

func TestIntegration_Client(t *testing.T) {
	heatPumpIP := os.Getenv("HEATPUMP_IP")
	if heatPumpIP == "" {
		heatPumpIP = "192.168.2.250" + ":" + DefaultPort
	}

	runTest := func(pm DataTypeMap, readFromNet func(Client) error) func(t *testing.T) {
		return func(t *testing.T) {
			t.Helper()
			c := MustNewClient(heatPumpIP, Options{
				SafeMode: true,
			})

			if err := c.Connect(); err != nil {
				t.Fatalf("connect client: %v", err)
			}
			defer func() {
				if err := c.Close(); err != nil {
					t.Errorf("close client: %v", err)
				}
			}()

			if err := readFromNet(c); err != nil {
				t.Fatalf("read from network: %v", err)
			}

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
			if err := tw.Flush(); err != nil {
				t.Fatalf("flush tabwriter: %v", err)
			}
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

	if err := c.Connect(); err != nil {
		t.Fatalf("connect client: %v", err)
	}
	defer func() {
		if err := c.Close(); err != nil {
			t.Errorf("close client: %v", err)
		}
	}()

	pm := NewCalculationsMap()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	tkr := time.NewTicker(3 * time.Second)

	for {
		if err := c.ReadCalculations(pm); err != nil {
			t.Fatalf("read calculations: %v", err)
		}

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

		if err := tw.Flush(); err != nil {
			t.Fatalf("flush tabwriter: %v", err)
		}
		select {
		case <-sigChan:
			return
		case tm := <-tkr.C:
			fmt.Println(tm.Format(time.DateTime), strings.Repeat("=", 200))
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
