# gosensors
Go bindings for libsensors.so from the lm-sensors project via cgo.
### Example
``` go
package main

import (
	"fmt"

	"github.com/md14454/gosensors"
)

func main() {
	gosensors.Init()
	defer gosensors.Cleanup()

	chips := gosensors.GetDetectedChips()

	for i := 0; i < len(chips); i++ {
		chip := chips[i]

		fmt.Printf("%v\n", chip)
		fmt.Printf("Adapter: %v\n", chip.AdapterName())

		features := chip.GetFeatures()

		for j := 0; j < len(features); j++ {
			feature := features[j]

			fmt.Printf("%v ('%v'): %.1f\n", feature.Name, feature.GetLabel(), feature.GetValue())

			subfeatures := feature.GetSubFeatures()

			for k := 0; k < len(subfeatures); k++ {
				subfeature := subfeatures[k]

				fmt.Printf("  %v: %.1f\n", subfeature.Name, subfeature.GetValue())
			}
		}

		fmt.Printf("\n")
	}
}
```
