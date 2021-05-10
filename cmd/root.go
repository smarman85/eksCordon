package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	//"github.com/spf13/viper"
)

var (
	client *kubernetes.Clientset
	zone   string
)

func init() {
	//config, err := rest.InClusterConfig()
	config, err := clientcmd.BuildConfigFromFlags("", os.Getenv("HOME")+"/.kube/config")
	if err != nil {
		//logging.LogErrorExitf("Error creating config: %v", err)
		fmt.Printf("Error creating config: %v", err)
	}

	client, err = kubernetes.NewForConfig(config)
	if err != nil {
		//logging.LogErrorExitf("Error creating config: %v", err)
		fmt.Printf("Error creating config: %v", err)
	}

	rootCmd.AddCommand(listAZs)
	rootCmd.AddCommand(cordonAZ)

	cordonAZ.Flags().StringVarP(&zone, "zone", "z", "", "sepcify an availability zone to cordon")
	cordonAZ.MarkFlagRequired("zone")
}

var rootCmd = &cobra.Command{
	Use:   "awsCordon",
	Short: "Helper script to cordon and drain a troublesome availability zone",
}

func Execute() error {
	return rootCmd.Execute()
}
