package main

import cli "github.com/ElrondNetwork/arwen-wasm-vm/v1_5/cmd/mandostestcli"

/// Legacy executor, still used by the older versions of erdpy.
func main() {
	cli.MandosTestCLI()
}
