package dispatcher

import (
	"context"
	"pack.ag/amqp"
	"sync"
	"time"

	"github.com/lawrencegripper/ion/internal/app/dispatcher/providers" //TODO couldn't it be moved into internal/pkg ?
	"github.com/lawrencegripper/ion/internal/pkg/messaging"            //TODO couldn't it be moved into internal/pkg ?
	"github.com/lawrencegripper/ion/internal/pkg/servicebus"
	"github.com/lawrencegripper/ion/internal/pkg/types"

	log "github.com/sirupsen/logrus"
)

// Run will start the dispatcher server and wait for new AMQP messages
func Run(cfg *types.Configuration) {
	ctx := context.Background()

	var provider providers.Provider
	amqpConnection := servicebus.NewAmqpConnection(ctx, cfg)
	handlerArgs := providers.GetSharedHandlerArgs(cfg, amqpConnection.AccessKeys)

	if cfg.AzureBatch != nil {
		log.Info("Using Azure batch provider...")
		batchProvider, err := providers.NewAzureBatchProvider(cfg, handlerArgs)
		if err != nil {
			log.WithError(err).Panic("Couldn't create azure batch provider")
		}
		provider = batchProvider
	} else {
		log.Info("Defaulting to using Kubernetes provider...")
		k8sProvider, err := providers.NewKubernetesProvider(cfg, handlerArgs)
		if err != nil {
			log.WithError(err).Panic("Couldn't create kubernetes provider")
		}
		provider = k8sProvider
	}

	var wg sync.WaitGroup

	wg.Add(4)
	go func() {
		defer wg.Done()
		for {
			// Locks are held for 1 mins, renew every 20 sec to keep locks
			time.Sleep(20 * time.Second)
			// allow 20seconds for the renew operation, keeping a 25 second buffer
			timeAllowanceForRenewalRequest := time.Second * 15

			// Renew message locks with ServiceBus
			//https://docs.microsoft.com/en-us/azure/service-bus-messaging/service-bus-amqp-request-response#message-renew-lock
			activeMessages := provider.GetActiveMessages()
			if len(activeMessages) < 1 {
				log.Debug("no active messages, skipping lock renewal")
				continue
			}

			messagesAMQP := make([]*amqp.Message, 0, len(activeMessages))
			for _, m := range activeMessages {
				originalMessage := m.GetAMQPMessage()
				messagesAMQP = append(messagesAMQP, originalMessage)
			}

			renewContextWithDeadline, cancel := context.WithTimeout(ctx, timeAllowanceForRenewalRequest)
			defer cancel()
			err := amqpConnection.RenewLocks(renewContextWithDeadline, messagesAMQP)
			if err != nil {
				// Todo: Additional could be put in here to cleanup operations. See: #171
				// https://github.com/lawrencegripper/ion/issues/171
				log.WithError(err).Panic("Failed to renew locks therefore cannot continue operation as message could be reassigned to another dispatcher.")
			}
		}
	}()
	go func() {
		defer wg.Done()
		for {
			// output queue stats every 30 seconds
			time.Sleep(30 * time.Second)
			queueStats, err := amqpConnection.GetQueueDepth()
			if err != nil {
				log.WithError(err).Error("failed getting queue depth from listener")
			}
			log.WithField("activeMessageCount", queueStats.ActiveMessageCount).WithField("deadLetteredMessageCount", queueStats.DeadLetterMessageCount).Info("listenerStats")
		}
	}()
	go func() {
		defer wg.Done()
		for {
			message, err := amqpConnection.Receiver.Receive(ctx)

			if err != nil {
				// Todo: Investigate the type of error here. If this could be triggered by a poisened message
				// app shouldn't panic.
				log.WithError(err).Panic("Error received dequeuing message")
			}

			if message == nil {
				log.WithError(err).Panic("Error received dequeuing message - nil message")
			}

			wrapper := messaging.NewAmqpMessageWrapper(message)
			contextualLogger := providers.GetLoggerForMessage(wrapper, log.NewEntry(log.StandardLogger()))
			contextualLogger.Debug("message received")

			if wrapper.DeliveryCount() > cfg.Job.RetryCount+1 {
				contextualLogger.Error("message re-received when above retryCount. AMQP provider wrongly redelivered message.")
				err := wrapper.Reject()
				if err != nil {
					contextualLogger.Error("error rejecting message")
				}
			}
			err = provider.Dispatch(wrapper)
			if err != nil {
				contextualLogger.WithError(err).Error("Couldn't dispatch message to kubernetes provider")
			}

			contextualLogger.Debug("message dispatched")
		}
	}()

	go func() {
		defer wg.Done()
		for {
			time.Sleep(time.Second * 15)

			if len(provider.GetActiveMessages()) < 1 {
				log.Debug("no active messages, skipping reconciling...")
				continue
			}

			log.Debug("reconciling...")

			err := provider.Reconcile()
			if err != nil {
				// Todo: Should this panic here? Should we tolerate a few failures (k8s upgade causing masters not to be vailable for example?)
				log.WithError(err).Panic("Failed to reconcile ....")
			}
			log.WithField("inProgress", provider.InProgressCount()).Info("providerStats")
		}
	}()
	wg.Wait()

	//init flaeg
	//flaeg := flaeg.New(rootCmd, os.Args[1:])

	//run test
	//if err := flaeg.Run(); err != nil {
	//	fmt.Printf("Error %s \n", err.Error())
	//}
}
