package client

			for _, line := range event.Terminal.Lines {
				log.Trace("job terminal output", "line", line.Raw)
				ui.Output(line.Line)
import (
	"context"
	"fmt"
	"io"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/hashicorp/waypoint/internal/server/gen"
	"github.com/hashicorp/waypoint/sdk/terminal"
)

// job returns the basic job skeleton prepoulated with the correct
// defaults based on how the client is configured. For example, for local
// operations, this will already have the targeting for the local runner.
func (c *Project) job() *pb.Job {
	return &pb.Job{
		Workspace:    c.workspace,
		TargetRunner: c.runner,
		Labels:       c.labels,

		DataSource: &pb.Job_Local_{
			Local: &pb.Job_Local{},
		},

		Operation: &pb.Job_Noop_{
			Noop: &pb.Job_Noop{},
		},
	}
}

// doJob will queue and execute the job. If the client is configured for
// local mode, this will start and target the proper runner.
func (c *Project) doJob(ctx context.Context, job *pb.Job, ui terminal.UI) (*pb.Job_Result, error) {
	log := c.logger

	// In local mode we have to start a runner.
	if c.local {
		log.Info("local mode, starting local runner")
		r, err := c.startRunner()
		if err != nil {
			return nil, err
		}

		log.Info("runner started", "runner_id", r.Id())

		// We defer the close so that we clean up resources. Local mode
		// always blocks and streams the full output so when doJob exits
		// the job is complete.
		defer r.Close()

		// Accept a job. Our local runners execute exactly one job.
		go func() {
			if err := r.Accept(); err != nil {
				log.Error("runner job accept error", "err", err)
			}
		}()

		// Modify the job to target this runner and use the local data source.
		job.TargetRunner = &pb.Ref_Runner{
			Target: &pb.Ref_Runner_Id{
				Id: &pb.Ref_RunnerId{
					Id: r.Id(),
				},
			},
		}
	}

	return c.queueAndStreamJob(ctx, job, ui)
}

// queueAndStreamJob will queue the job. If the client is configured to watch the job,
// it'll also stream the output to the configured UI.
func (c *Project) queueAndStreamJob(
	ctx context.Context,
	job *pb.Job,
	ui terminal.UI,
) (*pb.Job_Result, error) {
	log := c.logger

	// Queue the job
	log.Debug("queueing job", "operation", fmt.Sprintf("%T", job.Operation))
	queueResp, err := c.client.QueueJob(ctx, &pb.QueueJobRequest{
		Job: job,
	})
	if err != nil {
		return nil, err
	}
	log = log.With("job_id", queueResp.JobId)

	// Get the stream
	log.Debug("opening job stream")
	stream, err := c.client.GetJobStream(ctx, &pb.GetJobStreamRequest{
		JobId: queueResp.JobId,
	})
	if err != nil {
		return nil, err
	}

	// Wait for open confirmation
	resp, err := stream.Recv()
	if err != nil {
		return nil, err
	}
	if _, ok := resp.Event.(*pb.GetJobStreamResponse_Open_); !ok {
		return nil, status.Errorf(codes.Aborted,
			"job stream failed to open, got unexpected message %T",
			resp.Event)
	}

	// Process events
	var (
		stateEventTimer *time.Timer
		tstatus         terminal.Status

		stdout, stderr io.Writer
	)

	for {
		resp, err := stream.Recv()
		if err != nil {
			return nil, err
		}
		if resp == nil {
			// This shouldn't happen, but if it does, just ignore it.
			log.Warn("nil response received, ignoring")
			continue
		}

		switch event := resp.Event.(type) {
		case *pb.GetJobStreamResponse_Complete_:
			if event.Complete.Error == nil {
				log.Info("job completed successfully")
				return event.Complete.Result, nil
			}

			st := status.FromProto(event.Complete.Error)
			log.Warn("job failed", "code", st.Code(), "message", st.Message())
			return nil, st.Err()

		case *pb.GetJobStreamResponse_Error_:
			st := status.FromProto(event.Error.Error)
			log.Warn("job stream failure", "code", st.Code(), "message", st.Message())
			return nil, st.Err()

		case *pb.GetJobStreamResponse_Terminal_:
			for _, ev := range event.Terminal.Events {
				log.Trace("job terminal output", "event", ev)

				switch ev := ev.Event.(type) {
				case *pb.GetJobStreamResponse_Terminal_Event_Line_:
					ui.Output(ev.Line.Msg, terminal.WithStyle(ev.Line.Style))
				case *pb.GetJobStreamResponse_Terminal_Event_NamedValues_:
					var values []terminal.NamedValue

					for _, tnv := range ev.NamedValues.Values {
						values = append(values, terminal.NamedValue{
							Name:  tnv.Name,
							Value: tnv.Value,
						})
					}

					ui.NamedValues(values)
				case *pb.GetJobStreamResponse_Terminal_Event_Status_:
					if tstatus == nil {
						tstatus = ui.Status()
						defer tstatus.Close()
					}

					if ev.Status.Msg == "" && !ev.Status.Step {
						tstatus.Close()
					} else if ev.Status.Step {
						tstatus.Step(ev.Status.Status, ev.Status.Msg)
					} else {
						tstatus.Update(ev.Status.Msg)
					}
				case *pb.GetJobStreamResponse_Terminal_Event_Raw_:
					if stdout == nil {
						stdout, stderr, err = ui.OutputWriters()
						if err != nil {
							return nil, err
						}
					}

					if ev.Raw.Stderr {
						stderr.Write(ev.Raw.Data)
					} else {
						stdout.Write(ev.Raw.Data)
					}
				case *pb.GetJobStreamResponse_Terminal_Event_Table_:
					tbl := terminal.NewTable(ev.Table.Headers...)

					for _, row := range ev.Table.Rows {
						var trow []terminal.TableEntry

						for _, ent := range row.Entries {
							trow = append(trow, terminal.TableEntry{
								Value: ent.Value,
								Color: ent.Color,
							})
						}
					}

					ui.Table(tbl)
				}
			}
		case *pb.GetJobStreamResponse_State_:
			// Stop any state event timers if we have any since the state
			// has changed and we don't want to output that information anymore.
			if stateEventTimer != nil {
				stateEventTimer.Stop()
				stateEventTimer = nil
			}

			// For certain states, we do a quality of life UI message if
			// the wait time ends up being long.
			switch event.State.Current {
			case pb.Job_QUEUED:
				stateEventTimer = time.AfterFunc(stateEventPause, func() {
					ui.Output("Operation is queued. Waiting for runner assignment...",
						terminal.WithHeaderStyle())
					ui.Output("If you interrupt this command, the job will still run in the background.",
						terminal.WithInfoStyle())
				})

			case pb.Job_WAITING:
				stateEventTimer = time.AfterFunc(stateEventPause, func() {
					ui.Output("Operation is assigned to a runner. Waiting for start...",
						terminal.WithHeaderStyle())
					ui.Output("If you interrupt this command, the job will still run in the background.",
						terminal.WithInfoStyle())
				})
			}

		default:
			log.Warn("unknown stream event", "event", resp.Event)
		}
	}
}

const stateEventPause = 1500 * time.Millisecond
