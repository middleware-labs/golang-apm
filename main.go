package main

import (
	"github.com/middleware-labs/golang-apm/httptracer"
)

func main() {
     httptracer.Initialize(
		httptracer.WithConfigTag("service", "gin-go-k8s-demo"),
		httptracer.WithConfigTag("accessToken", "Adsadsa"),
		httptracer.WithConfigTag("target", "Adsadsa"),
	)
	


}

