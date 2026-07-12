// Command ai_simulation runs Panda's AI Digital Twin Simulation — see
// backend/tools/ai_simulation's package docs for the full architecture.
// This binary never touches production Rider/Driver apps, protobuf, or any
// production database; it is a standalone research tool.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fairride/ai_simulation/simulation"
)

func main() {
	drivers := flag.Int("drivers", 500, "number of simulated drivers")
	riders := flag.Int("riders", 5000, "number of simulated riders")
	days := flag.Int("days", 1, "number of simulated days (1 tick = 1 simulated minute)")
	model := flag.String("model", "phi4:14b", "Ollama model name for the 5% AI decision engine")
	ollamaURL := flag.String("ollama-url", "http://localhost:11434", "Ollama server base URL")
	seed := flag.Int64("seed", 0, "random seed (0 = derived from current time)")
	outDir := flag.String("out", "output", "output directory for JSON exports + dashboard.html (relative to the current working directory)")
	flag.Parse()

	fmt.Printf("Panda AI Digital Twin Simulation\n")
	fmt.Printf("  drivers=%d riders=%d days=%d model=%s ollama=%s\n", *drivers, *riders, *days, *model, *ollamaURL)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	engine := simulation.NewEngine(simulation.Config{
		Drivers: *drivers, Riders: *riders, Days: *days,
		Model: *model, OllamaURL: *ollamaURL, Seed: *seed,
	})

	if engine.AIEnabled() {
		fmt.Printf("  Ollama: reachable — AI decisions enabled for the ambiguous ~5%%\n")
	} else {
		fmt.Printf("  Ollama: NOT reachable at %s — running 100%% Rule Engine (per spec, this is a supported mode, not an error)\n", *ollamaURL)
	}

	start := time.Now()
	if err := engine.Run(ctx); err != nil && err != context.Canceled {
		fmt.Fprintf(os.Stderr, "simulation error: %v\n", err)
		os.Exit(1)
	}
	elapsed := time.Since(start)
	fmt.Printf("  Simulation finished in %s\n", elapsed.Round(time.Millisecond))

	if err := engine.Export(*outDir); err != nil {
		fmt.Fprintf(os.Stderr, "export error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("  Exported statistics + dashboard.html to %s\n", *outDir)
}
