package main

import (
	"fmt"
	"log"

	omnicli "github.com/omnicli/sdk-go"
)

//go:generate go run ../cmd/omni-metagen-go -struct=AppConfig -output=../example-dist/sdkgoexample2.metadata.yaml

// AppConfig demonstrates various field types and tags
// @help This is a sample application configuration
// with various field types and tags.
//
// @category Main Category, Sub Category
type AppConfig struct {
	// Basic types with short flags
	Name     string   // -n,--name
	Debug    bool     // -d,--debug
	Port     int      // -p,--port
	Timeout  float64  // -t,--timeout
	LogFile  *string  // -l,--log-file
	Verbose  *bool    // -v,--verbose
	Workers  *int     // -w,--workers
	Throttle *float64 // --throttle

	// Array types
	Host     []string  // -h,--host
	Features []bool    // --features
	Weights  []float64 // --weights

	// Unknown value
	Unknown *string `omniarg:"-"` // --unknown, not parsed into config struct

	// Non-exported fields are ignored
	ignored string // nolint:all
}

// DatabaseConfig demonstrates database-specific configuration
type DatabaseConfig struct {
	// All fields need tags as they have db_ prefix
	Host     string   `omniarg:"db_host"`    // --db-host
	Port     int      `omniarg:"db_port"`    // --db-port
	User     string   `omniarg:"db_user"`    // --db-user
	Password *string  `omniarg:"db_pass"`    // --db-pass
	Replicas []string `omniarg:"db_replica"` // --db-replica
}

func main() {
	// Parse arguments into both config structs at once
	var appCfg AppConfig
	var dbCfg DatabaseConfig

	_, err := omnicli.ParseArgs(&appCfg, &dbCfg)
	if err != nil {
		log.Fatalf("Failed to parse args: %v", err)
	}

	// Print the parsed configurations
	fmt.Println("Application Config:")
	fmt.Printf("  Name: %s\n", appCfg.Name)
	fmt.Printf("  Debug: %v\n", appCfg.Debug)
	fmt.Printf("  Port: %d\n", appCfg.Port)
	fmt.Printf("  Timeout: %f\n", appCfg.Timeout)

	fmt.Printf("  LogFile: %v\n", stringPtrValue(appCfg.LogFile))
	fmt.Printf("  Verbose: %v\n", boolPtrValue(appCfg.Verbose))
	fmt.Printf("  Workers: %v\n", intPtrValue(appCfg.Workers))
	fmt.Printf("  Throttle: %v\n", floatPtrValue(appCfg.Throttle))

	fmt.Printf("  Host: %v\n", appCfg.Host)
	fmt.Printf("  Features: %v\n", appCfg.Features)
	fmt.Printf("  Weights: %v\n", appCfg.Weights)

	fmt.Printf("\n")
	fmt.Println("Database Config:")
	fmt.Printf("  Host: %s\n", dbCfg.Host)
	fmt.Printf("  Port: %d\n", dbCfg.Port)
	fmt.Printf("  User: %s\n", dbCfg.User)
	fmt.Printf("  Password: %v\n", stringPtrValue(dbCfg.Password))
	fmt.Printf("  Replicas: %v\n", dbCfg.Replicas)
}

// Helper functions to safely print pointer values
func stringPtrValue(p *string) string {
	if p == nil {
		return "<not set>"
	}
	return *p
}

func boolPtrValue(p *bool) string {
	if p == nil {
		return "<not set>"
	}
	return fmt.Sprintf("%v", *p)
}

func intPtrValue(p *int) string {
	if p == nil {
		return "<not set>"
	}
	return fmt.Sprintf("%d", *p)
}

func floatPtrValue(p *float64) string {
	if p == nil {
		return "<not set>"
	}
	return fmt.Sprintf("%f", *p)
}
