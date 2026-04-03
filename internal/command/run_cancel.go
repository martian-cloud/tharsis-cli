package command

import (
	"errors"

	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
)

type runCancelCommand struct {
	*BaseCommand

	force *bool
}

var _ Command = (*runCancelCommand)(nil)

// NewRunCancelCommandFactory returns a runCancelCommand struct.
func NewRunCancelCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &runCancelCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *runCancelCommand) validate() error {
	if len(c.arguments) != 1 {
		return errors.New("expected exactly one argument: run id")
	}

	return nil
}

func (c *runCancelCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("run cancel"),
		WithInputValidator(c.validate),
		WithClient(true),
		WithWarningPrompt("This will forcefully cancel the run, which may leave resources in an inconsistent state."),
	); code != 0 {
		return code
	}

	runID := c.arguments[0]

	// Subscribe to run events.
	stream, err := c.grpcClient.RunsClient.SubscribeToRunEvents(c.Context, &pb.SubscribeToRunEventsRequest{
		RunId: &runID,
	})
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to subscribe to run events")
		return 1
	}

	input := &pb.CancelRunRequest{
		Id:    runID,
		Force: c.force,
	}

	if _, err = c.grpcClient.RunsClient.CancelRun(c.Context, input); err != nil {
		c.UI.ErrorWithSummary(err, "failed to cancel run")
		return 1
	}

	c.UI.Output("Run cancellation in progress...")

	// Wait for cancellation to complete.
	for {
		select {
		case <-c.Context.Done():
			c.UI.ErrorWithSummary(c.Context.Err(), "context canceled")
			return 1
		default:
			event, err := stream.Recv()
			if err != nil {
				c.UI.ErrorWithSummary(err, "failed to receive run event")
				return 1
			}

			switch event.Run.Status {
			case "canceled":
				c.UI.Successf("Run canceled successfully!")
				return 0
			case "applied", "planned", "planned_and_finished", "errored":
				c.UI.Errorf("Run completed with status: %s", event.Run.Status)
				return 1
			}
		}
	}
}

func (*runCancelCommand) Synopsis() string {
	return "Cancel a run."
}

func (*runCancelCommand) Description() string {
	return `
   Stops a running or pending run. Use -force when
   graceful cancellation is not sufficient.
`
}

func (*runCancelCommand) Usage() string {
	return "tharsis [global options] run cancel [options] <run-id>"
}

func (*runCancelCommand) Example() string {
	return `
tharsis run cancel -force <id>
`
}

func (c *runCancelCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")
	f.BoolVar(
		&c.force,
		"force",
		"Force the run to cancel.",
		flag.Aliases("f"),
	)

	return f
}
