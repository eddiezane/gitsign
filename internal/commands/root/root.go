package root

import (
	"github.com/spf13/cobra"

	"github.com/sigstore/gitsign/internal/config"
	inIO "github.com/sigstore/gitsign/internal/io"
)

type RootOptions struct {
	FlagSign   bool
	FlagVerify bool
	Config     *config.Config

	*inIO.Streams
}

func (o *RootOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVarP(&o.FlagSign, "sign", "s", false, "make a signature")
	cmd.Flags().BoolVarP(&o.FlagVerify, "verify", "v", false, "verify a signature")
}

func New(cfg *config.Config, s *inIO.Streams) *cobra.Command {
	o := &RootOptions{Config: cfg, Streams: s}

	rootCmd := &cobra.Command{
		Use:              "gitsign",
		Short:            "Keyless Git signing with Sigstore!",
		TraverseChildren: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if o.FlagSign {
				for _, item := range cmd.Commands() {
					if item.Name() == "sign" {
						item.RunE(item, cmd.Flags().Args())
					}
				}
			} else if o.FlagVerify {
				for _, item := range cmd.Commands() {
					if item.Name() == "verify" {
						item.RunE(item, cmd.Flags().Args())
					}
				}
			}
			return nil
		},
	}

	o.AddFlags(rootCmd)

	return rootCmd
}
