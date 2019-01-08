===
cli
===


Usage
=====

.. code-block:: go

   func ExampleCommand_ExecuteWithArgs() {
   	opts := struct {
   		Export, Verbose, Recursive bool
   		Label, Selector, Sort      string
   		Max, Min                   int
   		Interval                   time.Duration
   		Outs                       []string
   	}{}

   	app := &cli.Command{
   		Name: "app",
   	}

   	sub := &cli.Command{
   		Name: "sub",
   	}

   	subsub := &cli.Command{
   		Name: "subsub",
   		Run: func(_ context.Context, _ *cli.Command, _ []string) {
   			err := template.Must(template.New("").Parse(`
   export:    {{.Export}}
   verbose:   {{.Verbose}}
   recursive: {{.Recursive}}
   label:     {{.Label}}
   selector:  {{.Selector}}
   sort:      {{.Sort}}
   max:       {{.Max}}
   min:       {{.Min}}
   interval:  {{.Interval}}
   outs:      {{.Outs}}`)).Execute(os.Stdout, &opts)
   			if err != nil {
   				panic(err)
   			}
   		},
   	}

   	err := subsub.Options().
   		Add(&cli.BoolOpt{Var: &opts.Export, Long: "export"}).
   		Add(&cli.BoolOpt{Var: &opts.Verbose, Long: "verbose", Short: "v"}).
   		Add(&cli.BoolOpt{Var: &opts.Recursive, Long: "recursive", Short: "R"}).
   		Add(&cli.StringOpt{Var: &opts.Label, Long: "label"}).
   		Add(&cli.StringOpt{Var: &opts.Selector, Long: "selector", Short: "s"}).
   		Add(&cli.StringOpt{Var: &opts.Sort, Long: "sort", Default: "desc"}).
   		Add(&cli.IntOpt{Var: &opts.Max, Long: "max", Aliases: []string{"limit"}}).
   		Add(&cli.IntOpt{Var: &opts.Min, Long: "min", Default: 10}).
   		Add(&cli.DurationOpt{Var: &opts.Interval, Long: "interval"}).
   		Add(&cli.StringsOpt{Var: &opts.Outs, Long: "outs"}).Err

   	if err != nil {
   		panic(err)
   	}

   	app.
   		AddCommand(sub.AddCommand(subsub)).
   		ExecuteWithArgs(context.Background(), []string{
   			"--export", "--label=app", "--outs=aaa.yml", "sub", "-vR", "-s", "match", "subsub", "--limit", "100", "--interval=10s", "--outs=bbb.yml"})

   	// Output:
   	// export:    true
   	// verbose:   true
   	// recursive: true
   	// label:     app
   	// selector:  match
   	// sort:      desc
   	// max:       100
   	// min:       10
   	// interval:  10s
   	// outs:      [aaa.yml bbb.yml]
   }
