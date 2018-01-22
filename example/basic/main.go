// A basic skill implementation for deployment as a lambda
package main

import (
	"os"

	awslambda "github.com/aws/aws-lambda-go/lambda"
	"github.com/mctofu/alexa-skill/lambda"
)

func main() {
	appID := os.Getenv("ALEXA_APP_ID")
	app := newApp(appID)
	handler := lambda.NewDebugHandler(lambda.NewAppHandler(app))
	awslambda.Start(handler)
}
