# PUBQZ, a simple embedded pub/sub queueing library for Go

## simple example

```go
buzz := pubqz.New()

buzz.Sub("client_id", "channel_id", func(msg pubqz.Msg) error {
    // do something
})

buzz.Pub("channel_id", pubqz.NewTextMsg("howdy"))
```
