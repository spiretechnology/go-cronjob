# go-cronjob

A simple library for scheduling automated tasks within your Go program.

### Creating an automated task

Below is an example of a task that simply prints "Hello world" every 15 seconds:

```go
type HelloTask struct {}

func (t *HelloTask) Type() string {
	return "helloworld"
}

func (t *HelloTask) ScheduleFirstRun() time.Time {
	return time.Now()
}

func (t *HelloTask) DefaultRunInterval() time.Duration {
	return time.Second * 15
}

func (t *HelloTask) Run() (*cronjob.Result, error) {
	fmt.Println("Hello world")
	return nil, nil
}

```

The above `HelloTask` struct implements the `cronjob.CronJob` interface, which allows it to be registered to a `cronjob.Manager` (see below).

### Launching a task manager

A `Manager` handles the scheduling and execution of `CronJob`s that are registered to it. Here's a simple example, using the above `HelloTask`:

```go
manager := cronjob.NewManager(myDb)
manager.Register(
    &HelloTask{},
)
go manager.Run(nil)
```

### Considerations

If you're running your server as a cluster, each instance in the cluster will attempt to independently manage the cronjobs. This will likely result in duplicated work and undefined behavior. This library does not (yet) handle this situation.
