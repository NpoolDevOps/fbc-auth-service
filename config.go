package main

type Config struct {
	Port    int
	Targets []string
}

var MyConfig = Config{
	Port:    40008,
	Targets: []string{"auth.npool.com"},
}
