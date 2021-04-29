package cmd

import (
        "fmt"
        "os"
        "github.com/spf13/cobra"
        "k8s.io/client-go/tools/clientcmd"
        "k8s.io/client-go/kubernetes"
	      //"github.com/spf13/viper"
)

var Clientset *kubernetes.Clientset

var zone string

func init() {
        //config, err := rest.InClusterConfig()
        config, err := clientcmd.BuildConfigFromFlags("", os.Getenv("HOME") + "/.kube/config")
        if err != nil {
          //logging.LogErrorExitf("Error creating config: %v", err)
          fmt.Println("Error creating config: %v", err)
        }

        Clientset, err = kubernetes.NewForConfig(config)
        if err != nil {
          //logging.LogErrorExitf("Error creating config: %v", err)
          fmt.Println("Error creating config: %v", err)
        }

        rootCmd.AddCommand(listAZs)
        rootCmd.AddCommand(cordonAZ)

        //listAZs.Flags().BoolP("listazs", "l", false, "list all availability zones")
        cordonAZ.Flags().StringVarP(&zone, "zone", "z", "", "sepcify an availability zone to cordon")
        cordonAZ.MarkFlagRequired("zone")
}

var rootCmd = &cobra.Command{
        Use: "awsCordon",
        Short: "Helper script to cordon and drain a troublesome availability zone",
        /*Run: func(cmd *cobra.Command, args []string) {
                fmt.Println("Hello from cobra")
        },*/
}

func Execute() error {
        return rootCmd.Execute()
}
