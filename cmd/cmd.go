package cmd

import (
	"io"
	"sort"

	"github.com/cockroachdb/errors"
	"github.com/petems/tfvar-to-tfevar/pkg/tfvar"
	"github.com/spf13/cobra"
	"github.com/zclconf/go-cty/cty"
	"go.uber.org/zap"
)

const (
	flagAutoAssign = "auto-assign"
	flagDebug      = "debug"
	flagNoDefault  = "ignore-default"
	flagVar        = "var"
	flagVarFile    = "var-file"
	flagOrg        = "org"
	flagWorkspace  = "workspace"
)

// New returns a new instance of cobra.Command for tfvar. Usage:
//    c, sync := cmd.New(os.Stdout)
//    if err := c.Execute(); err != nil {
//    	log.Fatal(err)
//    }
//    sync()
func New(out io.Writer, version string) (*cobra.Command, func()) {
	r := &runner{
		out: out,
	}

	rootCmd := &cobra.Command{
		Use:   "tfvar-to-tfevar [DIR]",
		Short: "A CLI tool that helps export Terraform's variable definitions to Terraform Enterprise/Cloud Variables",
		Long: `Export variable definitions from variable definitions files (.tfvars) to Terraform Enterprise/Cloud.
`,
		PreRunE: r.preRootRunE,
		RunE:    r.rootRunE,
		Args:    cobra.ExactArgs(1),
		Version: version,
	}

	rootCmd.SetOut(out)

	rootCmd.PersistentFlags().BoolP(flagAutoAssign, "a", false, `Use values from environment variables TF_VAR_* and
variable definitions files e.g. terraform.tfvars[.json] *.auto.tfvars[.json]`)
	rootCmd.PersistentFlags().BoolP(flagDebug, "d", false, "Print debug log on stderr")
	rootCmd.PersistentFlags().Bool(flagNoDefault, false, "Do not use defined default values")
	rootCmd.PersistentFlags().StringArray(flagVar, []string{}, `Set a variable in the generated definitions.
This flag can be set multiple times.`)
	rootCmd.PersistentFlags().String(flagVarFile, "", `Set variables from a file.`)
	rootCmd.PersistentFlags().String(flagOrg, "example_organization", `Set the organisation for the generated terraform code.`)
	rootCmd.PersistentFlags().String(flagWorkspace, "example_workspace", `Set the workspace for the generated terraform code..`)

	return rootCmd, func() {
		if r.log != nil {
			_ = r.log.Sync()
		}
	}
}

type runner struct {
	out io.Writer
	log *zap.SugaredLogger
}

func (r *runner) preRootRunE(cmd *cobra.Command, args []string) error {
	// Setup logger
	logConfig := zap.NewDevelopmentConfig()

	isDebug, err := cmd.PersistentFlags().GetBool(flagDebug)
	if err != nil {
		return errors.Wrap(err, "cmd: get flag --debug")
	}

	if !isDebug {
		logConfig.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	logger, err := logConfig.Build()
	if err != nil {
		return errors.Wrap(err, "cmd: create new logger")
	}

	r.log = logger.Sugar()
	r.log.Debug("Logger initialized")

	return nil
}

func (r *runner) rootRunE(cmd *cobra.Command, args []string) error {
	dir := args[0]

	vars, err := tfvar.Load(dir)
	if err != nil {
		return err
	}

	sort.Slice(vars, func(i, j int) bool { return vars[i].Name < vars[j].Name })

	ignoreDefault, err := cmd.PersistentFlags().GetBool(flagNoDefault)
	if err != nil {
		return errors.Wrap(err, "cmd: get flag --ignore-default")
	}

	org, err := cmd.PersistentFlags().GetString(flagOrg)
	if err != nil {
		return errors.Wrap(err, "cmd: get flag --org")
	}

	workspace, err := cmd.PersistentFlags().GetString(flagWorkspace)
	if err != nil {
		return errors.Wrap(err, "cmd: get flag --workspace")
	}

	if ignoreDefault {
		r.log.Debug("Replacing values with null")
		for i, v := range vars {
			vars[i].Value = cty.NullVal(v.Value.Type())
		}
	}

	isAutoAssign, err := cmd.PersistentFlags().GetBool(flagAutoAssign)
	if err != nil {
		return errors.Wrap(err, "cmd: get flag --auto-assign")
	}

	unparseds := make(map[string]tfvar.UnparsedVariableValue)

	if isAutoAssign {
		r.log.Debug("Collecting values from environment variables")
		tfvar.CollectFromEnvVars(unparseds)

		autoFiles := tfvar.LookupTFVarsFiles(dir)

		for _, f := range autoFiles {
			if err := tfvar.CollectFromFile(f, unparseds); err != nil {
				return err
			}
		}
	}

	fvs, err := cmd.PersistentFlags().GetStringArray(flagVar)
	if err != nil {
		return errors.Wrap(err, "cmd: get flag --var")
	}

	for _, fv := range fvs {
		if err := tfvar.CollectFromString(fv, unparseds); err != nil {
			return err
		}
	}

	fromFile, err := cmd.PersistentFlags().GetString(flagVarFile)
	if err != nil {
		return errors.Wrap(err, "cmd: get flag --var-file")
	}

	if fromFile != "" {
		if err := tfvar.CollectFromFile(fromFile, unparseds); err != nil {
			return err
		}
	}

	vars, err = tfvar.ParseValues(unparseds, vars)
	if err != nil {
		return err
	}

	writer := tfvar.WriteAsTerraformCode

	return writer(r.out, vars, org, workspace)
}
