package main

type Config struct {
	Port int
	Domain string
}

var MyConfig = Config{
	Port: 40008,
	Domain: "auth.npool.com",

}
