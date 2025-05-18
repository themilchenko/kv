package config

type Config struct {
	Cluster []*Node `mapstructure:"cluster"`
	DataDir string  `mapstructure:"data_dir"`
	Leader  string  `mapstructure:"leader"`
	BinPath string  `mapstructure:"bin_path"`
}

type Node struct {
	Alias       string `mapstructure:"alias"`
	HttpAddress string `mapstructure:"http_address"`
	RpcAddress  string `mapstructure:"rpc_address"`
}
