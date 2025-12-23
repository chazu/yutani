package cli

import (
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/spf13/cobra"
)

var (
	profileDuration int
	profileOutput   string
	profileType     string
)

var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Profile a Yutani application",
	Long: `Profile a Yutani application for performance analysis.

Supported profile types:
  - cpu: CPU profiling
  - mem: Memory profiling
  - goroutine: Goroutine profiling
  - block: Block profiling
  - mutex: Mutex profiling`,
	Example: `  # CPU profile for 30 seconds
  yutani profile --type cpu --duration 30

  # Memory profile
  yutani profile --type mem --output mem.prof

  # Goroutine profile
  yutani profile --type goroutine`,
	RunE: runProfile,
}

func init() {
	profileCmd.Flags().IntVarP(&profileDuration, "duration", "d", 30, "Profile duration in seconds (for CPU profiling)")
	profileCmd.Flags().StringVarP(&profileOutput, "output", "o", "", "Output file (default: <type>.prof)")
	profileCmd.Flags().StringVarP(&profileType, "type", "t", "cpu", "Profile type (cpu, mem, goroutine, block, mutex)")
}

func runProfile(cmd *cobra.Command, args []string) error {
	// Set default output file
	if profileOutput == "" {
		profileOutput = fmt.Sprintf("%s.prof", profileType)
	}

	fmt.Printf("Starting %s profiling...\n", profileType)
	fmt.Printf("Output file: %s\n", profileOutput)

	switch profileType {
	case "cpu":
		return profileCPU()
	case "mem":
		return profileMemory()
	case "goroutine":
		return profileGoroutine()
	case "block":
		return profileBlock()
	case "mutex":
		return profileMutex()
	default:
		return fmt.Errorf("unknown profile type: %s", profileType)
	}
}

func profileCPU() error {
	f, err := os.Create(profileOutput)
	if err != nil {
		return fmt.Errorf("failed to create profile file: %w", err)
	}
	defer f.Close()

	if err := pprof.StartCPUProfile(f); err != nil {
		return fmt.Errorf("failed to start CPU profile: %w", err)
	}
	defer pprof.StopCPUProfile()

	fmt.Printf("Profiling for %d seconds...\n", profileDuration)
	time.Sleep(time.Duration(profileDuration) * time.Second)

	fmt.Printf("Profile saved to %s\n", profileOutput)
	fmt.Printf("\nAnalyze with: go tool pprof %s\n", profileOutput)
	return nil
}

func profileMemory() error {
	f, err := os.Create(profileOutput)
	if err != nil {
		return fmt.Errorf("failed to create profile file: %w", err)
	}
	defer f.Close()

	runtime.GC() // Get up-to-date statistics

	if err := pprof.WriteHeapProfile(f); err != nil {
		return fmt.Errorf("failed to write heap profile: %w", err)
	}

	fmt.Printf("Memory profile saved to %s\n", profileOutput)
	fmt.Printf("\nAnalyze with: go tool pprof %s\n", profileOutput)
	return nil
}

func profileGoroutine() error {
	f, err := os.Create(profileOutput)
	if err != nil {
		return fmt.Errorf("failed to create profile file: %w", err)
	}
	defer f.Close()

	profile := pprof.Lookup("goroutine")
	if profile == nil {
		return fmt.Errorf("goroutine profile not available")
	}

	if err := profile.WriteTo(f, 0); err != nil {
		return fmt.Errorf("failed to write goroutine profile: %w", err)
	}

	fmt.Printf("Goroutine profile saved to %s\n", profileOutput)
	fmt.Printf("\nAnalyze with: go tool pprof %s\n", profileOutput)
	return nil
}

func profileBlock() error {
	runtime.SetBlockProfileRate(1)
	defer runtime.SetBlockProfileRate(0)

	f, err := os.Create(profileOutput)
	if err != nil {
		return fmt.Errorf("failed to create profile file: %w", err)
	}
	defer f.Close()

	fmt.Printf("Profiling for %d seconds...\n", profileDuration)
	time.Sleep(time.Duration(profileDuration) * time.Second)

	profile := pprof.Lookup("block")
	if profile == nil {
		return fmt.Errorf("block profile not available")
	}

	if err := profile.WriteTo(f, 0); err != nil {
		return fmt.Errorf("failed to write block profile: %w", err)
	}

	fmt.Printf("Block profile saved to %s\n", profileOutput)
	fmt.Printf("\nAnalyze with: go tool pprof %s\n", profileOutput)
	return nil
}

func profileMutex() error {
	runtime.SetMutexProfileFraction(1)
	defer runtime.SetMutexProfileFraction(0)

	f, err := os.Create(profileOutput)
	if err != nil {
		return fmt.Errorf("failed to create profile file: %w", err)
	}
	defer f.Close()

	fmt.Printf("Profiling for %d seconds...\n", profileDuration)
	time.Sleep(time.Duration(profileDuration) * time.Second)

	profile := pprof.Lookup("mutex")
	if profile == nil {
		return fmt.Errorf("mutex profile not available")
	}

	if err := profile.WriteTo(f, 0); err != nil {
		return fmt.Errorf("failed to write mutex profile: %w", err)
	}

	fmt.Printf("Mutex profile saved to %s\n", profileOutput)
	fmt.Printf("\nAnalyze with: go tool pprof %s\n", profileOutput)
	return nil
}

