package cmd

import (
	"context"

	"github.com/KonstantinGasser/sherlock/internal"
	"github.com/KonstantinGasser/sherlock/terminal"
	"github.com/spf13/cobra"
)

func cmdUpdate(ctx context.Context, sherlock *internal.Sherlock) *cobra.Command {
	update := &cobra.Command{
		Use:   "update",
		Short: "update an accounts password or name",
		Long:  "update an accounts password or name",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
	}
	update.AddCommand(cmdUpdateAccPassword(ctx, sherlock))
	update.AddCommand(cmdUpdateAccName(ctx, sherlock))
	return update
}

type passwordOptions struct {
	insecure bool
}

func cmdUpdateAccPassword(ctx context.Context, sherlock *internal.Sherlock) *cobra.Command {
	var opts passwordOptions
	password := &cobra.Command{
		Use:   "password",
		Short: "change account password",
		Long:  "allows to change/update the password of an existing account",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			groupKey, err := terminal.ReadPassword("(%s) password: ", args[0])
			if err != nil {
				terminal.Error(err.Error())
				return
			}
			password, err := terminal.ReadPassword("(%s) new password: ", args[0])
			if err != nil {
				terminal.Error(err.Error())
				return
			}
			if err := sherlock.UpdateState(ctx, args[0], groupKey, internal.OptAccPassword(password, opts.insecure)); err != nil {
				terminal.Error(err.Error())
				return
			}
			terminal.Info("account password updated")
		},
	}
	password.Flags().BoolVarP(&opts.insecure, "insecure", "i", false, "allow insecure password for account")
	return password
}

func cmdUpdateAccName(ctx context.Context, sherlock *internal.Sherlock) *cobra.Command {
	name := &cobra.Command{
		Use:   "name",
		Short: "change account name",
		Long:  "allows to change/update the account of an existing account",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			groupKey, err := terminal.ReadPassword("(%s) password: ", args[0])
			if err != nil {
				terminal.Error(err.Error())
				return
			}
			name, err := terminal.ReadLine("(%s) new account name: ", args[0])
			if err != nil {
				terminal.Error(err.Error())
				return
			}
			if err := sherlock.UpdateState(ctx, args[0], groupKey, internal.OptAccName(name)); err != nil {
				terminal.Error(err.Error())
				return
			}
			terminal.Info("account name updated")
		},
	}
	return name
}
