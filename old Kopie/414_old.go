package commands

import (
	sternCmd "github.com/wercker/stern/cmd"
	"github.com/wercker/stern/kubernetes"
	cmd.Use = "logs <appname>"
	cmd.Short = "Tail pods logs of an application"
	cmd.Long = "Tail pods logs of an application"
	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
			ioStreams.Errorf("please specify the application name")
		appName := args[0]
		cmd.SetArgs([]string{appName + ".*"})
		namespace := env.Namespace
		cmd.Flags().String("namespace", namespace, "")
		return nil
	}
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		config, err := sternCmd.ParseConfig(args)
		if err := Run(ctx, config, ioStreams); err != nil {
	"context"
	"fmt"
	"regexp"
	"text/template"
	"time"

	"github.com/oam-dev/kubevela/api/types"
	"github.com/oam-dev/kubevela/pkg/application"
	cmdutil "github.com/oam-dev/kubevela/pkg/commands/util"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/wercker/stern/stern"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

func NewLogsCommand(c types.Args, ioStreams cmdutil.IOStreams) *cobra.Command {
	largs := &Args{C: c}
	cmd := &cobra.Command{}
	cmd.Use = "logs"
	cmd.Short = "Tail logs for application"
	cmd.Long = "Tail logs for application"
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			ioStreams.Errorf("please specify app name")
			return nil
		}
		env, err := GetEnv(cmd)
		if err != nil {
			return err
		}
		app, err := application.Load(env.Name, args[0])
		if err != nil {
			return err
		}
		largs.App = app
		largs.Env = env
		ctx := context.Background()
		if err := largs.Run(ctx, ioStreams); err != nil {
			return err
		}
		return nil
	}
	cmd.Annotations = map[string]string{
		types.TagCommandType: types.TypeApp,
	}
	cmd.Flags().StringVarP(&largs.Output, "output", "o", "default", "output format for logs, support: [default, raw, json]")
	return cmd
}

type Args struct {
	Output string
	Env    *types.EnvMeta
	C      types.Args
	App    *application.Application
}

func (l *Args) AskComponent() (string, error) {
	comps := l.App.GetComponents()
	if len(comps) == 1 {
		return comps[0], nil
	}
	prompt := &survey.Select{
		Message: "You have multiple Component in your app. Choose one component for logs: ",
		Options: comps,
	}
	var compName string
	err := survey.AskOne(prompt, &compName)
	if err != nil {
		return "", fmt.Errorf("choosing component name err %v", err)
	}
	return compName, nil
}

// Run refer to the implementation at https://github.com/oam-dev/stern/blob/master/stern/main.go
func (l *Args) Run(ctx context.Context, ioStreams cmdutil.IOStreams) error {

	clientSet, err := kubernetes.NewForConfig(l.C.Config)
	if err != nil {
		return err
	}
	compName, err := l.AskComponent()
	if err != nil {
		return err
	}
	//TODO(wonderflow): we could get labels from service to narrow the pods scope selected
	labelSelector := labels.Everything()
	pod, err := regexp.Compile(compName + "-.*")
	if err != nil {
		return fmt.Errorf("fail to compile '%s' for logs query", compName+".*")
	}
	container := regexp.MustCompile(".*")
	namespace := l.Env.Namespace
	added, removed, err := stern.Watch(ctx, clientSet.CoreV1().Pods(namespace), pod, container, nil, stern.RUNNING, labelSelector)
	if err != nil {
		return err
	}
	tails := make(map[string]*stern.Tail)
	logC := make(chan string, 1024)

	go func() {
		for {
			select {
			case str := <-logC:
				ioStreams.Infonln(str)
			case <-ctx.Done():
				return
			}
		}
	}()

	var t string
	switch l.Output {
	case "default":
		if color.NoColor {
			t = "{{.PodName}} {{.ContainerName}} {{.Message}}"
		} else {
			t = "{{color .PodColor .PodName}} {{color .ContainerColor .ContainerName}} {{.Message}}"
		}
	case "raw":
		t = "{{.Message}}"
	case "json":
		t = "{{json .}}\n"
	}
	funs := map[string]interface{}{
		"json": func(in interface{}) (string, error) {
			b, err := json.Marshal(in)
			if err != nil {
				return "", err
			}
			return string(b), nil
		},
		"color": func(color color.Color, text string) string {
			return color.SprintFunc()(text)
		},
	}
	template, err := template.New("log").Funcs(funs).Parse(t)
	if err != nil {
		return errors.Wrap(err, "unable to parse template")
	}

	go func() {
		for p := range added {
			id := p.GetID()
			if tails[id] != nil {
				continue
			}
			//48h
			dur, _ := time.ParseDuration("48h")
			tail := stern.NewTail(p.Namespace, p.Pod, p.Container, template, &stern.TailOptions{
				Timestamps:   true,
				SinceSeconds: int64(dur.Seconds()),
				Exclude:      nil,
				Include:      nil,
				Namespace:    false,
				TailLines:    nil, //default for all logs
			})
			tails[id] = tail

			tail.Start(ctx, clientSet.CoreV1().Pods(p.Namespace), logC)
		}
	}()

	go func() {
		for p := range removed {
			id := p.GetID()
			if tails[id] == nil {
				continue
			}
			tails[id].Close()
			delete(tails, id)
		}
	}()

	<-ctx.Done()

	return nil
}
