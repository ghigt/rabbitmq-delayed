package main

// AMQP is the amqp configuration
type AMQP struct {
	URL string `default:"amqp://guest:guest@127.0.0.1:5672/guest"`
}

// Config is the application configuration
type Config struct {
	AppName    string `json:"app-name" default:"rabbitmq"`
	AppVersion string `json:"app-version" required:"true"`

	AMQP AMQP
}
