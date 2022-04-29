package main

import (
	"flag"
	"github.com/naoina/toml"
	"github.com/orznewbie/gotmpl/pkg/log"
	"os"
)

func main() {
	flag.Parse()

	cfg, err := LoadConfigFile(*configFile)
	if err != nil {
		panic(err)
	}
	log.Info(cfg)
}

var (
	configFile = flag.String("config", "../../../configs/config.toml", "Configuration file to use")
)

// LoadConfigFile parses the specified file into a Config object
func LoadConfigFile(filename string) (cfg Config, err error) {
	f, err := os.Open(filename)
	if err != nil {
		return cfg, err
	}
	defer f.Close()

	return cfg, toml.NewDecoder(f).Decode(&cfg)
}

type Config struct {
	RPCRelays []RPCConfig `toml:"rpc"`
}

type RPCConfig struct {
	// Name identifies the HTTP relay
	Name string `toml:"name"`

	// Addr should be set to the desired listening host:port
	Addr string `toml:"bind-addr"`

	// Persistent buffer path
	BufferDir string `toml:"buffer-dir"`
}
