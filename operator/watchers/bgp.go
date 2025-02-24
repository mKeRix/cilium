// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Cilium

package watchers

import (
	"context"

	"github.com/cilium/cilium/pkg/bgp/manager"
	k8sClient "github.com/cilium/cilium/pkg/k8s/client"
	"github.com/cilium/cilium/pkg/k8s/utils"
	"github.com/cilium/cilium/pkg/lock"
)

// StartLBIPAllocator starts the service watcher if it hasn't already and looks
// for service of type LoadBalancer. Once it finds a service of that type, it
// will try to allocate an external IP (LoadBalancerIP) for it.
func StartLBIPAllocator(ctx context.Context, cfg ServiceSyncConfiguration, clientset k8sClient.Clientset) {
	optsModifier, err := utils.GetServiceListOptionsModifier(cfg)
	if err != nil {
		log.WithError(err).Fatal("Error creating service option modifier")
	}

	swgSvcs := lock.NewStoppableWaitGroup()
	swgEps := lock.NewStoppableWaitGroup()
	InitServiceWatcher(cfg, clientset, swgSvcs, swgEps, optsModifier)

	m, err := manager.New(ctx, clientset, serviceIndexer)
	if err != nil {
		log.WithError(err).Fatal("Error creating BGP manager")
	}
	serviceSubscribers.Register(m)

	go func() {
		<-k8sSvcCacheSynced
		m.MarkSynced()
	}()
}
