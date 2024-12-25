package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var (
	rootCommand = &cobra.Command{
		Use:   "etcd-k8s-extract",
		Short: "etcd-k8s-extract reads an etcd db and writes all the kubernetes resources to file.",
		Run:   etcdKubernetesExtract,
	}
)

var (
	flockTimeout       time.Duration
	iterateBucketLimit uint64
	outputPath         string
)

func init() {
	rootCommand.PersistentFlags().DurationVar(&flockTimeout, "timeout", 10*time.Second, "time to wait to obtain a file lock on db file, 0 to block indefinitely")
	rootCommand.PersistentFlags().StringVar(&outputPath, "output-path", "", "path where the kuberegistry will be written")
}

func main() {
	if err := rootCommand.Execute(); err != nil {
		fmt.Fprintln(os.Stdout, err)
		os.Exit(1)
	}
}

func etcdKubernetesExtract(_ *cobra.Command, args []string) {
	if len(args) != 1 {
		log.Fatalf("Must provide 1 argument pointing at the etcd data directory or db file (got %d)", len(args))
	}
	dp := args[0]
	if !strings.HasSuffix(dp, "db") {
		dp = filepath.Join(snapDir(dp), "db")
	}
	if !existFileOrDir(dp) {
		log.Fatalf("%q does not exist", dp)
	}

	err := iterateEtcdKeys(dp, outputPath, iterateBucketLimit)
	if err != nil {
		log.Fatal(err)
	}
}

func existFileOrDir(name string) bool {
	_, err := os.Stat(name)
	return err == nil
}
