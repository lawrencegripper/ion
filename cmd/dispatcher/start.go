package main

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewCmdStart() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Instanciate the dispatcher to process events",
		Run: func(cmd *cobra.Command, args []string) {
			//TODO Prepare all CLI flags & configs here and then start call Run()
			log.Println("start run")

			log.Println(cmd.Flags().GetString("loglevel"))
			log.Println(viper.GetString("loglevel"))
			//dispatcher.Run()
		},
	}

	//cmd.PersistentFlags().

	return cmd
}
