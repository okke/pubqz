# PUBQZ, a simple embedded pub/sub queing library

## simple example

```go
buzz := bus.New()

buzz.Sub("client_id", "channel_id", func(msg Msg) error {
    // do something
})

buzz.Pub("channel_id", bus.NewTextMsg("howdy"))
```
