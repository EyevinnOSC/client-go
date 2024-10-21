<h1 align="center">
  OSAAS Go Client
</h1>


<div align="center">
Go client library for operating with
<a href="https://www.osaas.io">Eyevinn Open Source Cloud</a>
</div>

<div align="center">
<br />


[![github release](https://img.shields.io/github/v/release/Eyevinn/osaas-client-go?style=flat-square)](https://github.com/Eyevinn/osaas-client-go/releases)
[![license](https://img.shields.io/github/license/Eyevinn/osaas-client-go.svg?style=flat-square)](LICENSE)

[![PRs welcome](https://img.shields.io/badge/PRs-welcome-ff69b4.svg?style=flat-square)](https://github.com/eyevinn/osaas-client-go/issues?q=is%3Aissue+is%3Aopen+label%3A%22help+wanted%22)
[![made with hearth by Eyevinn](https://img.shields.io/badge/made%20with%20%E2%99%A5%20by-Eyevinn-59cbe8.svg?style=flat-square)](https://github.com/eyevinn)
[![Slack](http://slack.streamingtech.se/badge.svg)](http://slack.streamingtech.se)

</div>

<!-- Add a description of the project here -->


## Installation
Import library in your project
```
import "github.com/eyevinn/osaas-client-go"
```

## Usage

Example of creating, listing and removing an instance

```go
package main

import "github.com/eyevinn/osaas-client-go"
import "fmt"
import "os"

func main() {
  config := &osaasclient.ContextConfig{
	  PersonalAccessToken: os.Getenv("OSC_PAT"),
	  Environment:         "dev",
  }

  ctx, err := osaasclient.NewContext(config)
  if err != nil {
    fmt.Println("Error creating context:", err)
    return
  }

  token, err := ctx.GetServiceAccessToken("encore")
  if err != nil {
    fmt.Println("Error getting service access token:", err)
    return
  }

  instances, err := osaasclient.ListInstances(ctx, "encore", token)

  if err == nil {
    fmt.Printf("instances: %s\n", instances)
  } else {
    fmt.Println("Error listing instances:", err)
  }

  instanceName := "test-instance"

  instance, err := osaasclient.CreateInstance(ctx, "encore", token, map[string]interface{}{
    "name": instanceName,
  })

  if err != nil {
    fmt.Println("Error creating instance:", err)
  } else {
    fmt.Printf("Instance created: %s\n", instance)
  }
  
  err = osaasclient.RemoveInstance(ctx, "encore", instanceName, token)

  if err != nil {
    fmt.Println("Error removing instance:", err)
  } else {
    fmt.Printf("Instance %s removed\n", instanceName)
  }
}


```


## Contributing

See [CONTRIBUTING](CONTRIBUTING.md)

## License

This project is licensed under the MIT License, see [LICENSE](LICENSE).

# Support

Join our [community on Slack](http://slack.streamingtech.se) where you can post any questions regarding any of our open source projects. Eyevinn's consulting business can also offer you:

- Further development of this component
- Customization and integration of this component into your platform
- Support and maintenance agreement

Contact [sales@eyevinn.se](mailto:sales@eyevinn.se) if you are interested.

# About Eyevinn Technology

[Eyevinn Technology](https://www.eyevinntechnology.se) is an independent consultant firm specialized in video and streaming. Independent in a way that we are not commercially tied to any platform or technology vendor. As our way to innovate and push the industry forward we develop proof-of-concepts and tools. The things we learn and the code we write we share with the industry in [blogs](https://dev.to/video) and by open sourcing the code we have written.

Want to know more about Eyevinn and how it is to work here. Contact us at work@eyevinn.se!
