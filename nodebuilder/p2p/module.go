package p2p

import (
	logging "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/metrics"
	"github.com/libp2p/go-libp2p/core/network"
	rcmgr "github.com/libp2p/go-libp2p/p2p/host/resource-manager"
	"go.uber.org/fx"

	"github.com/celestiaorg/celestia-node/nodebuilder/node"
)

var log = logging.Logger("module/p2p")

// ConstructModule collects all the components and services related to p2p.
func ConstructModule(tp node.Type, cfg *Config) fx.Option {
	// sanitize config values before constructing module
	cfgErr := cfg.Validate()

	baseComponents := fx.Options(
		fx.Supply(*cfg),
		fx.Error(cfgErr),
		fx.Provide(Key),
		fx.Provide(ID),
		fx.Provide(PeerStore),
		fx.Provide(ConnectionManager),
		fx.Provide(ConnectionGater),
		fx.Provide(Host),
		fx.Provide(RoutedHost),
		fx.Provide(PubSub),
		fx.Provide(DataExchange),
		fx.Provide(BlockService),
		fx.Provide(PeerRouting),
		fx.Provide(ContentRouting),
		fx.Provide(AddrsFactory(cfg.AnnounceAddresses, cfg.NoAnnounceAddresses)),
		fx.Provide(metrics.NewBandwidthCounter),
		fx.Provide(newModule),
		fx.Invoke(Listen(cfg.ListenAddresses)),
	)

	switch tp {
	case node.Full, node.Bridge:
		return fx.Module(
			"p2p",
			baseComponents,
			fx.Provide(func() (network.ResourceManager, error) {
				return rcmgr.NewResourceManager(rcmgr.NewFixedLimiter(rcmgr.InfiniteLimits))
			}),
		)
	case node.Light:
		return fx.Module(
			"p2p",
			baseComponents,
			fx.Provide(func() (network.ResourceManager, error) {
				limits := rcmgr.DefaultLimits
				libp2p.SetDefaultServiceLimits(&limits)
				return rcmgr.NewResourceManager(rcmgr.NewFixedLimiter(limits.AutoScale()))
			}),
		)
	default:
		panic("invalid node type")
	}
}
