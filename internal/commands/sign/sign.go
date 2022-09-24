package sign

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/sigstore/gitsign/internal/config"
	"github.com/sigstore/gitsign/internal/fulcio"
	"github.com/sigstore/gitsign/internal/git"
	inIO "github.com/sigstore/gitsign/internal/io"
	"github.com/sigstore/gitsign/internal/signature"
)

type SignOptions struct {
	Config *config.Config

	*inIO.Streams
}

func New(cfg *config.Config, s *inIO.Streams) *cobra.Command {
	o := &SignOptions{Config: cfg, Streams: s}

	cmdSign := &cobra.Command{
		Use:   "sign",
		Short: "Make a signature",
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.Run()
		},
	}
	return cmdSign
}

func (o *SignOptions) Run() error {
	ctx := context.Background()
	userIdent, err := fulcio.NewIdentity(ctx, o.Config, o.TTYIn, o.TTYOut)
	if err != nil {
		return fmt.Errorf("failed to get identity: %w", err)
	}

	// Git is looking for "\n[GNUPG:] SIG_CREATED ", meaning we need to print a
	// line before SIG_CREATED. BEGIN_SIGNING seems appropriate. GPG emits this,
	// though GPGSM does not.
	sBeginSigning.emit()

	var f io.ReadCloser
	if len(fileArgs) == 1 {
		if f, err = os.Open(fileArgs[0]); err != nil {
			return fmt.Errorf("failed to open message file (%s): %w", fileArgs[0], err)
		}
		defer f.Close()
	} else {
		f = stdin
	}

	dataBuf := new(bytes.Buffer)
	if _, err = io.Copy(dataBuf, f); err != nil {
		return fmt.Errorf("failed to read message from stdin: %w", err)
	}

	rekor, err := newRekorClient(cfg.Rekor)
	if err != nil {
		return fmt.Errorf("failed to create rekor client: %w", err)
	}

	sig, cert, err := git.Sign(ctx, rekor, userIdent, dataBuf.Bytes(), signature.SignOptions{
		Detached:           *detachSignFlag,
		TimestampAuthority: *tsaOpt,
		Armor:              *armorFlag,
		IncludeCerts:       *includeCertsOpt,
	})
	if err != nil {
		return fmt.Errorf("failed to sign message: %w", err)
	}

	emitSigCreated(cert, *detachSignFlag)

	if _, err := stdout.Write(sig); err != nil {
		return errors.New("failed to write signature")
	}

	return nil
}
