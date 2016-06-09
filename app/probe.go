package app

import (
  "github.com/urfave/cli"
  "net/http"
  "net/url"
  "errors"
  "time"
  "math"
  "net"
  "fmt"
  "os"
)

func ProbeCommand() cli.Command {

  return cli.Command{
    Name:  "probe",
    Usage: "Probe a TCP/HTTP(S) endpoint to determine if it is healthy",
    ArgsUsage: "<endpoint>",
    Action: probe,
    Flags: []cli.Flag{
      cli.DurationFlag{
        Name: "timeout, t",
        Usage: "Connection timeout in seconds",
        Value: 5 * time.Second,
      },
      cli.BoolFlag{
        Name:  "loop",
        Usage: "Continuously probe endpoint until it is healthy",
      },
      cli.Float64Flag{
        Name:  "backoff, b",
        Usage: "(Loop) Rate at which to back off from retries, must be >= 1",
        Value: 1.0,
      },
      cli.DurationFlag{
        Name:  "min, m",
        Usage: "(Loop) Minimum time to wait before retrying",
        Value: 1 * time.Second,
      },
      cli.DurationFlag{
        Name:  "max, x",
        Usage: "(Loop) Maximum time to wait before retrying",
        Value: 120 * time.Second,
      },
    },
  }
}

func probe(c *cli.Context) error {
  if c.Args().First() == "" {
    cli.ShowCommandHelp(c, "probe")
    os.Exit(1)
  }

  if c.Bool("loop") {
    min := c.Duration("min")
    max := c.Duration("max")
    backoff := c.Float64("backoff")
    loops := 0
    delay := min

    for err := healthCheck(c); err != nil; err = healthCheck(c) {
      fmt.Println(err)
      time.Sleep(delay)
      loops += 1
      
      if delay < max {
        delay = time.Duration(float64(min) * math.Pow(backoff, float64(loops)))
      }
      if delay > max {
        delay = max
      }
    }
    fmt.Println("OK")

  } else {
    if err := healthCheck(c); err != nil {
      fmt.Println(err)
      os.Exit(1)
    }
    fmt.Println("OK")
  }
  return nil
}

func healthCheck(c *cli.Context) error {
  endpoint := c.Args().First()
  timeout := c.Duration("timeout")

  url, err := url.Parse(endpoint)
  if err != nil {
    return err
  }

  switch url.Scheme {
  case "tcp":
    var conn net.Conn
    if conn, err = net.DialTimeout(url.Scheme, url.Host, timeout); err != nil {
      return err
    }
    conn.Close()
  case "http", "https":
    client := &http.Client{
      Timeout: timeout,
    }
    var resp *http.Response
    resp, err = client.Get(endpoint)

    switch {
    case err != nil:
      return err
    case resp.StatusCode >= 200 && resp.StatusCode <= 299:
      return nil
    default:
      return errors.New(fmt.Sprintf("HTTP %d\n", resp.StatusCode))
    }
  default:
    return errors.New(fmt.Sprintf("Unsupported URL scheme: %s\n", url.Scheme))
  }
  return nil
}
