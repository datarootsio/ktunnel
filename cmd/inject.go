package cmd

import (
	"context"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"

	"github.com/omrikiei/ktunnel/pkg/client"
	"github.com/omrikiei/ktunnel/pkg/k8s"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var Namespace string
var ServerImage string

var injectCmd = &cobra.Command{
	Use:   "inject",
	Short: "Inject server sidecar to the cluster and run the ktunnel client to establish a connection",
	Long: `This command accepts a pod/deployment and injects the tunnel sidecar to that artifact, 
			it then establishes a reverse tunnel`,
}

var injectDeploymentCmd = &cobra.Command{
	Use:   "deployment [flags] DEPLOYMENT_NAME [ports]",
	Short: "Inject server sidecar to a deployment and run the ktunnel client to establish a connection",
	Args:  cobra.MinimumNArgs(2),
	Example: `
# Inject a back tunnel from a running deployment to local mysql and redis 
ktunnel inject deploymeny mydeployment 3306 6379
`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithCancel(context.Background())
		if verbose {
			logger.SetLevel(log.DebugLevel)
			k8s.Verbose = true
		}
		o := sync.Once{}
		// Inject
		deployment := args[0]
		readyChan := make(chan bool, 1)
		_, err := k8s.InjectSidecar(&Namespace, &deployment, &port, ServerImage, readyChan)
		if err != nil {
			log.Fatalf("failed injecting sidecar: %v", err)
		}

		done := make(chan bool, 1)

		// Signal hook to remove sidecar
		sigs := make(chan os.Signal, 1)
		wg := &sync.WaitGroup{}
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGQUIT)
		stopChan := make(chan struct{}, 1)

		go func() {
			o.Do(func() {
				<-sigs
				log.Info("Stopping streams")
				cancel()
				wg.Wait()
				readyChan = make(chan bool, 1)
				ok, err := k8s.RemoveSidecar(&Namespace, &deployment, ServerImage, readyChan)
				if !ok {
					log.Errorf("Failed removing tunnel sidecar; %v", err)
				}
				<-readyChan
				log.Info("Finished, exiting")
				done <- true

			})
		}()

		log.Info("Waiting for deployment to be ready")
		<-readyChan

		// port-Forward
		strPort := strconv.FormatInt(int64(port), 10)
		// Create a tunnel client for each replica
		sourcePorts, err := k8s.PortForward(&Namespace, &deployment, strPort, wg, stopChan)
		if err != nil {
			log.Fatalf("Failed to run port forwarding: %v", err)
			os.Exit(1)
		}
		log.Info("Waiting for port forward to finish")
		wg.Wait()
		for _, srcPort := range *sourcePorts {
			go func() {
				p, err := strconv.ParseInt(srcPort, 10, 0)
				if err != nil {
					log.Fatalf("Failed to run client: %v", err)
				}
				prt := int(p)
				opts := []client.ClientOption{
					client.WithServer(Host, prt),
					client.WithTunnels(Scheme, args[1:]...),
					client.WithLogger(&logger),
				}
				if tls {
					opts = append(opts, client.WithTLS(CertFile, ServerHostOverride))
				}
				err = client.RunClient(ctx, opts...)
				if err != nil {
					log.Fatalf("Failed to run client: %v", err)
				}
			}()
		}
		_ = <-done
	},
}

func init() {
	injectCmd.Flags().StringVarP(&CaFile, "ca-file", "c", "", "TLS cert auth file")
	injectCmd.Flags().StringVarP(&Scheme, "scheme", "s", "tcp", "Connection scheme")
	injectCmd.Flags().StringVarP(&ServerHostOverride, "server-host-override", "o", "", "Server name use to verify the hostname returned by the TLS handshake")
	injectCmd.Flags().StringVarP(&Namespace, "namespace", "n", "default", "Namespace")
	injectDeploymentCmd.Flags().StringVarP(&CaFile, "ca-file", "c", "", "tls cert auth file")
	injectDeploymentCmd.Flags().StringVarP(&Scheme, "scheme", "s", "tcp", "Connection scheme")
	injectDeploymentCmd.Flags().StringVarP(&ServerHostOverride, "server-host-override", "o", "", "Server name use to verify the hostname returned by the TLS handshake")
	injectDeploymentCmd.Flags().StringVarP(&Namespace, "namespace", "n", "default", "Namespace")
	injectDeploymentCmd.Flags().StringVarP(&ServerImage, "server-image", "i", k8s.Image, "Ktunnel server image to use")
	injectCmd.AddCommand(injectDeploymentCmd)
	rootCmd.AddCommand(injectCmd)
}
