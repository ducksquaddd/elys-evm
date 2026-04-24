package keepers

import (
	"fmt"
	"os"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	evidencekeeper "cosmossdk.io/x/evidence/keeper"
	evidencetypes "cosmossdk.io/x/evidence/types"
	"cosmossdk.io/x/feegrant"
	feegrantkeeper "cosmossdk.io/x/feegrant/keeper"
	upgradekeeper "cosmossdk.io/x/upgrade/keeper"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmTypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/runtime"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authcodec "github.com/cosmos/cosmos-sdk/x/auth/codec"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	consensusparamkeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	consensusparamtypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	crisiskeeper "github.com/cosmos/cosmos-sdk/x/crisis/keeper"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/cosmos/cosmos-sdk/x/group"
	groupkeeper "github.com/cosmos/cosmos-sdk/x/group/keeper"
	"github.com/cosmos/cosmos-sdk/x/params"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	paramproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v10/packetforward"
	packetforwardkeeper "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v10/packetforward/keeper"
	packetforwardtypes "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v10/packetforward/types"
	capabilitykeeper "github.com/cosmos/ibc-go/modules/capability/keeper"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	icacontroller "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/controller"
	icacontrollerkeeper "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/controller/keeper"
	icacontrollertypes "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/controller/types"
	icahost "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/host"
	icahostkeeper "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/host/keeper"
	icahosttypes "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/host/types"
	"github.com/cosmos/ibc-go/v10/modules/apps/transfer"
	ibctransferkeeper "github.com/cosmos/ibc-go/v10/modules/apps/transfer/keeper"
	ibctransfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v10/modules/core/02-client/types"
	ibcconnectiontypes "github.com/cosmos/ibc-go/v10/modules/core/03-connection/types"
	porttypes "github.com/cosmos/ibc-go/v10/modules/core/05-port/types"
	ibcexported "github.com/cosmos/ibc-go/v10/modules/core/exported"
	ibckeeper "github.com/cosmos/ibc-go/v10/modules/core/keeper"
	ccvconsumer "github.com/cosmos/interchain-security/v7/x/ccv/consumer"
	ccvconsumerkeeper "github.com/cosmos/interchain-security/v7/x/ccv/consumer/keeper"
	ccvconsumertypes "github.com/cosmos/interchain-security/v7/x/ccv/consumer/types"
	ccv "github.com/cosmos/interchain-security/v7/x/ccv/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"

	feemarketkeeper "github.com/cosmos/evm/x/feemarket/keeper"
	feemarkettypes "github.com/cosmos/evm/x/feemarket/types"
	evmkeeper "github.com/cosmos/evm/x/vm/keeper"
	evmtypes "github.com/cosmos/evm/x/vm/types"

	"github.com/elys-network/elys/v6/app/precompiles"
	accountedpoolmodulekeeper "github.com/elys-network/elys/v6/x/accountedpool/keeper"
	accountedpoolmoduletypes "github.com/elys-network/elys/v6/x/accountedpool/types"
	ammmodulekeeper "github.com/elys-network/elys/v6/x/amm/keeper"
	ammmoduletypes "github.com/elys-network/elys/v6/x/amm/types"
	assetprofilemodulekeeper "github.com/elys-network/elys/v6/x/assetprofile/keeper"
	assetprofilemoduletypes "github.com/elys-network/elys/v6/x/assetprofile/types"
	burnermodulekeeper "github.com/elys-network/elys/v6/x/burner/keeper"
	burnermoduletypes "github.com/elys-network/elys/v6/x/burner/types"
	commitmentmodulekeeper "github.com/elys-network/elys/v6/x/commitment/keeper"
	commitmentmoduletypes "github.com/elys-network/elys/v6/x/commitment/types"
	epochsmodulekeeper "github.com/elys-network/elys/v6/x/epochs/keeper"
	epochsmoduletypes "github.com/elys-network/elys/v6/x/epochs/types"
	estakingmodulekeeper "github.com/elys-network/elys/v6/x/estaking/keeper"
	estakingmoduletypes "github.com/elys-network/elys/v6/x/estaking/types"
	leveragelpmodulekeeper "github.com/elys-network/elys/v6/x/leveragelp/keeper"
	leveragelpmoduletypes "github.com/elys-network/elys/v6/x/leveragelp/types"
	masterchefmodulekeeper "github.com/elys-network/elys/v6/x/masterchef/keeper"
	masterchefmoduletypes "github.com/elys-network/elys/v6/x/masterchef/types"
	legacyoraclekeeper "github.com/elys-network/elys/v6/x/oracle/keeper"
	legacyoracletypes "github.com/elys-network/elys/v6/x/oracle/types"
	parametermodulekeeper "github.com/elys-network/elys/v6/x/parameter/keeper"
	parametermoduletypes "github.com/elys-network/elys/v6/x/parameter/types"
	perpetualmodulekeeper "github.com/elys-network/elys/v6/x/perpetual/keeper"
	perpetualmoduletypes "github.com/elys-network/elys/v6/x/perpetual/types"
	stablestakekeeper "github.com/elys-network/elys/v6/x/stablestake/keeper"
	stablestaketypes "github.com/elys-network/elys/v6/x/stablestake/types"
	tiermodulekeeper "github.com/elys-network/elys/v6/x/tier/keeper"
	tiermoduletypes "github.com/elys-network/elys/v6/x/tier/types"
	tokenomicsmodulekeeper "github.com/elys-network/elys/v6/x/tokenomics/keeper"
	tokenomicsmoduletypes "github.com/elys-network/elys/v6/x/tokenomics/types"
	tradeshieldmodulekeeper "github.com/elys-network/elys/v6/x/tradeshield/keeper"
	tradeshieldmoduletypes "github.com/elys-network/elys/v6/x/tradeshield/types"
	oraclekeeper "github.com/ojo-network/ojo/x/oracle/keeper"
	oracletypes "github.com/ojo-network/ojo/x/oracle/types"
	"github.com/spf13/cast"
	// this line is used by starport scaffolding # stargate/app/moduleImport
)

type AppKeepers struct {
	keys    map[string]*storetypes.KVStoreKey
	tkeys   map[string]*storetypes.TransientStoreKey
	memKeys map[string]*storetypes.MemoryStoreKey

	// keepers
	AccountKeeper    authkeeper.AccountKeeper
	BankKeeper       bankkeeper.Keeper
	CapabilityKeeper *capabilitykeeper.Keeper
	StakingKeeper    *stakingkeeper.Keeper
	SlashingKeeper   slashingkeeper.Keeper
	DistrKeeper      distrkeeper.Keeper
	GovKeeper        *govkeeper.Keeper
	CrisisKeeper     *crisiskeeper.Keeper
	UpgradeKeeper    *upgradekeeper.Keeper
	ParamsKeeper     paramskeeper.Keeper

	// IBC Keeper must be a pointer in the app, so we can SetRouter on it correctly
	IBCKeeper             *ibckeeper.Keeper
	ICAHostKeeper         icahostkeeper.Keeper
	ICAControllerKeeper   icacontrollerkeeper.Keeper
	EvidenceKeeper        evidencekeeper.Keeper
	TransferKeeper        *ibctransferkeeper.Keeper
	FeeGrantKeeper        feegrantkeeper.Keeper
	AuthzKeeper           authzkeeper.Keeper
	ConsensusParamsKeeper consensusparamkeeper.Keeper

	// EVM keepers
	FeeMarketKeeper feemarketkeeper.Keeper
	EVMKeeper       *evmkeeper.Keeper
	GroupKeeper     groupkeeper.Keeper

	ConsumerKeeper ccvconsumerkeeper.Keeper
	ConsumerModule ccvconsumer.AppModule // Have to declare this here for IBC router

	WasmKeeper wasmkeeper.Keeper

	// make scoped keepers public for test purposes
	ScopedIBCKeeper           capabilitykeeper.ScopedKeeper
	ScopedTransferKeeper      capabilitykeeper.ScopedKeeper
	ScopedICAHostKeeper       capabilitykeeper.ScopedKeeper
	ScopedICAControllerKeeper capabilitykeeper.ScopedKeeper
	ScopedIBCFeeKeeper        capabilitykeeper.ScopedKeeper
	ScopedOracleKeeper        capabilitykeeper.ScopedKeeper
	ScopedCCVConsumerKeeper   capabilitykeeper.ScopedKeeper
	ScopedWasmKeeper          capabilitykeeper.ScopedKeeper

	EpochsKeeper        *epochsmodulekeeper.Keeper
	AssetprofileKeeper  assetprofilemodulekeeper.Keeper
	LegacyOracleKeepper legacyoraclekeeper.Keeper
	OracleKeeper        oraclekeeper.Keeper
	CommitmentKeeper    *commitmentmodulekeeper.Keeper
	TokenomicsKeeper    tokenomicsmodulekeeper.Keeper
	BurnerKeeper        burnermodulekeeper.Keeper
	AmmKeeper           *ammmodulekeeper.Keeper
	ParameterKeeper     parametermodulekeeper.Keeper
	PerpetualKeeper     *perpetualmodulekeeper.Keeper
	AccountedPoolKeeper accountedpoolmodulekeeper.Keeper
	StablestakeKeeper   *stablestakekeeper.Keeper
	LeveragelpKeeper    *leveragelpmodulekeeper.Keeper
	MasterchefKeeper    masterchefmodulekeeper.Keeper
	EstakingKeeper      *estakingmodulekeeper.Keeper
	TierKeeper          *tiermodulekeeper.Keeper
	TradeshieldKeeper   tradeshieldmodulekeeper.Keeper

	PacketForwardKeeper *packetforwardkeeper.Keeper

	ICSValidatorKeeper ICSValidatorKeeper
}

func (appKeepers AppKeepers) GetKVStoreKeys() map[string]*storetypes.KVStoreKey {
	return appKeepers.keys
}

func (appKeepers AppKeepers) GetTransientStoreKeys() map[string]*storetypes.TransientStoreKey {
	return appKeepers.tkeys
}

func (appKeepers AppKeepers) GetMemKeys() map[string]*storetypes.MemoryStoreKey {
	return appKeepers.memKeys
}

func NewAppKeeper(
	appCodec codec.Codec,
	bApp *baseapp.BaseApp,
	legacyAmino *codec.LegacyAmino,
	maccPerms map[string][]string,
	modAccAddrs map[string]bool,
	blockedAddress map[string]bool,
	skipUpgradeHeights map[int64]bool,
	homePath string,
	invCheckPeriod uint,
	logger log.Logger,
	appOpts servertypes.AppOptions,
	AccountAddressPrefix string,
	wasmOpts []wasmkeeper.Option,
) AppKeepers {
	app := AppKeepers{}

	// Set keys KVStoreKey, TransientStoreKey, MemoryStoreKey
	app.GenerateKeys()

	/*
		configure state listening capabilities using AppOptions
		we are doing nothing with the returned streamingServices and waitGroup in this case
	*/
	// load state streaming if enabled

	if err := bApp.RegisterStreamingServices(appOpts, app.keys); err != nil {
		logger.Error("failed to load state streaming", "err", err)
		os.Exit(1)
	}

	app.ParamsKeeper = initParamsKeeper(
		appCodec,
		legacyAmino,
		app.keys[paramstypes.StoreKey],
		app.tkeys[paramstypes.TStoreKey],
	)
	// set the BaseApp's parameter store
	app.ConsensusParamsKeeper = consensusparamkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(app.keys[consensusparamtypes.StoreKey]),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		runtime.EventService{},
	)
	bApp.SetParamStore(app.ConsensusParamsKeeper.ParamsStore)

	app.ParameterKeeper = *parametermodulekeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(app.keys[parametermoduletypes.StoreKey]),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	// add capability keeper and ScopeToModule for ibc module
	app.CapabilityKeeper = capabilitykeeper.NewKeeper(
		appCodec,
		app.keys[capabilitytypes.StoreKey],
		app.memKeys[capabilitytypes.MemStoreKey],
	)

	app.ScopedIBCKeeper = app.CapabilityKeeper.ScopeToModule(ibcexported.ModuleName)
	app.ScopedICAHostKeeper = app.CapabilityKeeper.ScopeToModule(icahosttypes.SubModuleName)
	app.ScopedICAControllerKeeper = app.CapabilityKeeper.ScopeToModule(icacontrollertypes.SubModuleName)
	app.ScopedTransferKeeper = app.CapabilityKeeper.ScopeToModule(ibctransfertypes.ModuleName)
	app.ScopedOracleKeeper = app.CapabilityKeeper.ScopeToModule(oracletypes.ModuleName)
	app.ScopedCCVConsumerKeeper = app.CapabilityKeeper.ScopeToModule(ccvconsumertypes.ModuleName)
	app.ScopedWasmKeeper = app.CapabilityKeeper.ScopeToModule(wasmTypes.ModuleName)

	// Add normal keepers
	app.AccountKeeper = authkeeper.NewAccountKeeper(
		appCodec,
		runtime.NewKVStoreService(app.keys[authtypes.StoreKey]),
		authtypes.ProtoBaseAccount,
		maccPerms,
		address.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix()),
		sdk.GetConfig().GetBech32AccountAddrPrefix(),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	app.BankKeeper = bankkeeper.NewBaseKeeper(
		appCodec,
		runtime.NewKVStoreService(app.keys[banktypes.StoreKey]),
		app.AccountKeeper,
		blockedAddress,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		logger,
	)

	app.CrisisKeeper = crisiskeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(app.keys[crisistypes.StoreKey]),
		invCheckPeriod,
		app.BankKeeper,
		authtypes.FeeCollectorName,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		app.AccountKeeper.AddressCodec(),
	)

	app.AuthzKeeper = authzkeeper.NewKeeper(
		runtime.NewKVStoreService(app.keys[authzkeeper.StoreKey]),
		appCodec,
		bApp.MsgServiceRouter(),
		app.AccountKeeper,
	)

	app.FeeGrantKeeper = feegrantkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(app.keys[feegrant.StoreKey]),
		app.AccountKeeper,
	)

	app.StakingKeeper = stakingkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(app.keys[stakingtypes.StoreKey]),
		app.AccountKeeper,
		app.BankKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		authcodec.NewBech32Codec(sdk.GetConfig().GetBech32ValidatorAddrPrefix()),
		authcodec.NewBech32Codec(sdk.GetConfig().GetBech32ConsensusAddrPrefix()),
	)

	// Initialize EVM keepers in correct order
	// 1) Fee market for EIP-1559 style fees
	app.FeeMarketKeeper = feemarketkeeper.NewKeeper(
		appCodec,
		authtypes.NewModuleAddress(govtypes.ModuleName),
		app.keys[feemarkettypes.StoreKey],
		app.tkeys[feemarkettypes.TransientKey],
	)

	app.AssetprofileKeeper = *assetprofilemodulekeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(app.keys[assetprofilemoduletypes.StoreKey]),
		app.TransferKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	app.CommitmentKeeper = commitmentmodulekeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(app.keys[commitmentmoduletypes.StoreKey]),
		app.AccountKeeper,
		app.BankKeeper,
		app.StakingKeeper,
		app.AssetprofileKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	app.TokenomicsKeeper = *tokenomicsmodulekeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(app.keys[tokenomicsmoduletypes.StoreKey]),
		app.CommitmentKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	app.EstakingKeeper = estakingmodulekeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(app.keys[estakingmoduletypes.StoreKey]),
		app.ParameterKeeper,
		app.StakingKeeper,
		app.CommitmentKeeper,
		&app.DistrKeeper,
		app.AssetprofileKeeper,
		app.TokenomicsKeeper,
		&app.ConsumerKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	app.SlashingKeeper = slashingkeeper.NewKeeper(
		appCodec,
		legacyAmino,
		runtime.NewKVStoreService(app.keys[slashingtypes.StoreKey]),
		app.StakingKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	groupConfig := group.DefaultConfig()
	/*
		Example of setting group params:
		groupConfig.MaxMetadataLen = 1000
	*/
	app.GroupKeeper = groupkeeper.NewKeeper(
		app.keys[group.StoreKey],
		appCodec,
		bApp.MsgServiceRouter(),
		app.AccountKeeper,
		groupConfig,
	)

	// UpgradeKeeper must be created before IBCKeeper
	app.UpgradeKeeper = upgradekeeper.NewKeeper(
		skipUpgradeHeights,
		runtime.NewKVStoreService(app.keys[upgradetypes.StoreKey]),
		appCodec,
		homePath,
		bApp,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	// ... other modules keepers
	// pre-initialize ConsumerKeeper to satisfy ibckeeper.NewKeeper
	// which would panic on nil or zero keeper
	// ConsumerKeeper implements StakingKeeper but all function calls result in no-ops so this is safe
	// communication over IBC is not affected by these changes
	app.ConsumerKeeper = ccvconsumerkeeper.NewNonZeroKeeper(
		appCodec,
		app.keys[ccvconsumertypes.StoreKey],
	)

	// UpgradeKeeper must be created before IBCKeeper
	app.IBCKeeper = ibckeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(app.keys[ibcexported.StoreKey]),
		app.GetSubspace(ibcexported.ModuleName),
		app.UpgradeKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	// Native IBC v10 callbacks will be configured in the transfer stack below

	transferKeeper := ibctransferkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(app.keys[ibctransfertypes.StoreKey]),
		app.GetSubspace(ibctransfertypes.ModuleName),
		app.IBCKeeper.ChannelKeeper, // ICS4Wrapper will be set later in transfer stack
		app.IBCKeeper.ChannelKeeper,
		bApp.MsgServiceRouter(),
		app.AccountKeeper,
		app.BankKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	app.TransferKeeper = &transferKeeper

	// 2) Create EVM-compatible account keeper wrapper
	evmAccountKeeper := NewEVMAccountKeeper(app.AccountKeeper)

	// 3) Core EVM keeper
	fmt.Printf("🔧 Creating EVM keeper...\n")
	app.EVMKeeper = evmkeeper.NewKeeper(
		appCodec,
		app.keys[evmtypes.StoreKey],
		app.tkeys[evmtypes.TransientKey],
		app.GetKVStoreKeys(),
		authtypes.NewModuleAddress(govtypes.ModuleName),
		evmAccountKeeper,
		app.BankKeeper,
		app.StakingKeeper,
		app.FeeMarketKeeper,
		app.ConsensusParamsKeeper,
		nil, // no erc20 keeper (ELYS-EVM doesn't need it)
		"",  // tracer
	)
	fmt.Printf("✅ EVM keeper created successfully\n")

	// Enable universal bank wrapper precompiles
	// Configure EVM precompiles directly (avoiding import cycle)
	fmt.Printf("🔧 Setting up precompiles...\n")
	// Clone the mapping from the latest EVM fork.
	staticPrecompiles := make(map[common.Address]vm.PrecompiledContract)

	// Copy Berlin precompiles as base
	for addr, contract := range vm.PrecompiledContractsBerlin {
		staticPrecompiles[addr] = contract
		fmt.Printf("   ✅ Added Berlin precompile: %s\n", addr.Hex())
	}

	// Add our custom bank precompile for universal bank operations
	fmt.Printf("🔧 Creating bank precompile...\n")
	bankPrecompile, err := precompiles.NewBankPrecompile(app.BankKeeper)
	if err != nil {
		panic(fmt.Sprintf("failed to create bank precompile: %v", err))
	}
	staticPrecompiles[bankPrecompile.Address()] = bankPrecompile
	fmt.Printf("   ✅ Added bank precompile: %s\n", bankPrecompile.Address().Hex())

	// Add factory precompile for CREATE2 wrapper deployment
	fmt.Printf("🔧 Creating wrapper factory precompile...\n")
	factoryPrecompile := precompiles.NewWrapperFactoryPrecompile(app.AssetprofileKeeper)
	staticPrecompiles[factoryPrecompile.Address()] = factoryPrecompile
	fmt.Printf("   ✅ Added factory precompile: %s\n", factoryPrecompile.Address().Hex())

	// Configure EVM keeper with precompiles
	fmt.Printf("🔧 Registering %d total precompiles with EVM keeper...\n", len(staticPrecompiles))
	app.EVMKeeper.WithStaticPrecompiles(staticPrecompiles)
	fmt.Printf("✅ All precompiles registered successfully\n")

	app.PacketForwardKeeper = packetforwardkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(app.keys[packetforwardtypes.StoreKey]),
		app.TransferKeeper,
		app.IBCKeeper.ChannelKeeper,
		app.BankKeeper,
		// Use the channel keeper directly - callbacks will be handled in the transfer stack
		app.IBCKeeper.ChannelKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	app.ICAHostKeeper = icahostkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(app.keys[icahosttypes.StoreKey]),
		app.GetSubspace(icahosttypes.SubModuleName),
		app.IBCKeeper.ChannelKeeper, // ICS4Wrapper
		app.IBCKeeper.ChannelKeeper,
		app.AccountKeeper,
		bApp.MsgServiceRouter(),
		bApp.GRPCQueryRouter(),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	app.ICAControllerKeeper = icacontrollerkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(app.keys[icacontrollertypes.StoreKey]),
		app.GetSubspace(icacontrollertypes.SubModuleName),
		app.IBCKeeper.ChannelKeeper, // ICS4Wrapper
		app.IBCKeeper.ChannelKeeper,
		bApp.MsgServiceRouter(),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	// wasmDir := filepath.Join(homePath, "wasm")
	// wasmConfig, err := wasm.ReadNodeConfig(appOpts)
	// if err != nil {
	// 	panic(fmt.Sprintf("error while reading wasm config: %s", err))
	// }

	// bankKeeper := app.BankKeeper.(bankkeeper.BaseKeeper)
	// wasmOpts = append(
	// 	wasmbindingsclient.RegisterCustomPlugins(
	// 		app.AmmKeeper,
	// 		&app.AccountKeeper,
	// 		&bankKeeper,
	// 	),
	// 	wasmOpts...,
	// )
	// wasmOpts = append(wasmbindingsclient.RegisterStargateQueries(*bApp.GRPCQueryRouter(), appCodec), wasmOpts...)

	// TODO: Fix WASM keeper interface compatibility with IBC v10
	// WASM integration requires more complex interface bridging for IBC v10
	// Temporarily initialize empty keeper to allow core IBC v10 functionality to work
	app.WasmKeeper = wasmkeeper.Keeper{}

	app.ConsumerKeeper = ccvconsumerkeeper.NewKeeper(
		appCodec,
		app.keys[ccvconsumertypes.StoreKey],
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.ConnectionKeeper,
		app.IBCKeeper.ClientKeeper,
		app.SlashingKeeper,
		app.BankKeeper,
		app.AccountKeeper,
		app.TransferKeeper,
		app.IBCKeeper,
		authtypes.FeeCollectorName,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		address.NewBech32Codec(sdk.GetConfig().GetBech32ValidatorAddrPrefix()),
		address.NewBech32Codec(sdk.GetConfig().GetBech32ConsensusAddrPrefix()),
	)

	app.ConsumerKeeper.SetStandaloneStakingKeeper(app.StakingKeeper)

	// consumer keeper satisfies the staking keeper interface
	// of the slashing module
	app.SlashingKeeper = slashingkeeper.NewKeeper(
		appCodec,
		legacyAmino,
		runtime.NewKVStoreService(app.keys[slashingtypes.StoreKey]),
		&app.ConsumerKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	app.ConsumerKeeper = *app.ConsumerKeeper.SetHooks(app.SlashingKeeper.Hooks())

	app.ICSValidatorKeeper = NewICSValidatorKeeper(app.ConsumerKeeper)

	app.ConsumerModule = ccvconsumer.NewAppModule(app.ConsumerKeeper, app.GetSubspace(ccvconsumertypes.ModuleName))

	evidenceKeeper := evidencekeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(app.keys[evidencetypes.StoreKey]),
		&app.ConsumerKeeper,
		app.SlashingKeeper,
		app.AccountKeeper.AddressCodec(),
		runtime.ProvideCometInfoService(),
	)
	// If evidence needs to be handled for the app, set routes in router here and seal
	app.EvidenceKeeper = *evidenceKeeper

	app.LegacyOracleKeepper = *legacyoraclekeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(app.keys[legacyoracletypes.StoreKey]),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	app.OracleKeeper = oraclekeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(app.keys[oracletypes.StoreKey]),
		app.AccountKeeper,
		app.BankKeeper,
		app.ICSValidatorKeeper,
		distrtypes.ModuleName,
		cast.ToBool(appOpts.Get("telemetry.enabled")),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	app.EpochsKeeper = epochsmodulekeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(app.keys[epochsmoduletypes.StoreKey]),
	)

	app.AccountedPoolKeeper = *accountedpoolmodulekeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(app.keys[accountedpoolmoduletypes.StoreKey]),
		app.BankKeeper,
		app.OracleKeeper,
	)

	app.AmmKeeper = ammmodulekeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(app.keys[ammmoduletypes.StoreKey]),
		app.tkeys[ammmoduletypes.TStoreKey],
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		&app.ParameterKeeper,
		app.BankKeeper,
		app.AccountKeeper,
		app.OracleKeeper,
		app.CommitmentKeeper,
		app.AssetprofileKeeper,
		app.AccountedPoolKeeper,
		app.TierKeeper,
	)

	app.StablestakeKeeper = stablestakekeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(app.keys[stablestaketypes.StoreKey]),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		app.BankKeeper,
		app.CommitmentKeeper,
		app.AssetprofileKeeper,
		app.OracleKeeper,
		app.AmmKeeper,
	)

	app.CommitmentKeeper.SetHooks(
		commitmentmodulekeeper.NewMultiCommitmentHooks(
			app.EstakingKeeper.CommitmentHooks(),
		),
	)

	app.DistrKeeper = distrkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(app.keys[distrtypes.StoreKey]),
		app.AccountKeeper,
		app.CommitmentKeeper,
		app.EstakingKeeper,
		ccvconsumertypes.ConsumerRedistributeName,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	app.PerpetualKeeper = perpetualmodulekeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(app.keys[perpetualmoduletypes.StoreKey]),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		app.AmmKeeper,
		app.BankKeeper,
		app.OracleKeeper,
		app.AssetprofileKeeper,
		&app.ParameterKeeper,
		app.TierKeeper,
	)

	app.MasterchefKeeper = *masterchefmodulekeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(app.keys[masterchefmoduletypes.StoreKey]),
		app.ParameterKeeper,
		app.CommitmentKeeper,
		app.AmmKeeper,
		app.OracleKeeper,
		app.AssetprofileKeeper,
		app.AccountedPoolKeeper,
		app.StablestakeKeeper,
		app.TokenomicsKeeper,
		app.AccountKeeper,
		app.BankKeeper,
		app.EstakingKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	app.BurnerKeeper = *burnermodulekeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(app.keys[burnermoduletypes.StoreKey]),
		app.BankKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	// provider depends on gov, so gov must be registered first
	govConfig := govtypes.DefaultConfig()
	// set the MaxMetadataLen for proposals to the same value as it was pre-sdk v0.47.x
	govConfig.MaxMetadataLen = 10200
	app.GovKeeper = govkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(app.keys[govtypes.StoreKey]),
		app.AccountKeeper,
		app.BankKeeper,
		// No need to send EstakingKeeper here as gov only does sk.IterateBondedValidatorsByPower, no need to give vp to Eden and EdenB
		app.StakingKeeper,
		app.DistrKeeper,
		bApp.MsgServiceRouter(),
		govConfig,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	govRouter := govv1beta1.NewRouter()
	govRouter.
		AddRoute(govtypes.RouterKey, govv1beta1.ProposalHandler).
		AddRoute(paramproposal.RouterKey, params.NewParamChangeProposalHandler(app.ParamsKeeper))

	app.GovKeeper.SetLegacyRouter(govRouter)

	app.LeveragelpKeeper = leveragelpmodulekeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(app.keys[leveragelpmoduletypes.StoreKey]),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		app.AmmKeeper,
		app.BankKeeper,
		app.OracleKeeper,
		app.StablestakeKeeper,
		app.CommitmentKeeper,
		app.MasterchefKeeper,
		app.AccountedPoolKeeper,
	)

	app.StablestakeKeeper.SetLeverageLpKeeper(app.LeveragelpKeeper)

	app.TierKeeper = tiermodulekeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(app.keys[tiermoduletypes.StoreKey]),
		app.BankKeeper,
		app.OracleKeeper,
		app.AssetprofileKeeper,
		app.AmmKeeper,
		app.EstakingKeeper,
		app.MasterchefKeeper,
		app.CommitmentKeeper,
		app.StakingKeeper,
		app.PerpetualKeeper,
		app.LeveragelpKeeper,
		app.StablestakeKeeper,
		app.TradeshieldKeeper,
	)
	app.AmmKeeper.SetTierKeeper(app.TierKeeper)
	app.PerpetualKeeper.SetTierKeeper(app.TierKeeper)

	app.TradeshieldKeeper = *tradeshieldmodulekeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(app.keys[tradeshieldmoduletypes.StoreKey]),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		app.BankKeeper,
		app.AmmKeeper,
		app.PerpetualKeeper,
	)

	app.TierKeeper.SetTradeshieldKeeper(&app.TradeshieldKeeper)

	// this line is used by starport scaffolding # stargate/app/keeperDefinition

	/**** IBC Routing ****/

	// Sealing prevents other modules from creating scoped sub-keepers
	app.CapabilityKeeper.Seal()

	// Create Transfer Stack (from bottom to top of stack)
	// - core IBC
	// - ibcfee
	// - ratelimit
	// - pfm
	// - provider
	// - transfer
	//
	// This is how transfer stack will work in the end:
	// * RecvPacket -> IBC core -> Fee -> RateLimit -> PFM -> Provider -> Transfer (AddRoute)
	// * SendPacket -> Transfer -> Provider -> PFM -> RateLimit -> Fee -> IBC core (ICS4Wrapper)

	var transferStack porttypes.IBCModule
	transferStack = transfer.NewIBCModule(*app.TransferKeeper)
	// Add native IBC v10 callbacks middleware if needed
	// transferStack = ibccallbacks.NewIBCMiddleware(transferStack, app.IBCKeeper.ChannelKeeper, maxCallbackGas)
	transferStack = packetforward.NewIBCMiddleware(
		transferStack,
		app.PacketForwardKeeper,
		0, // retries on timeout
		packetforwardkeeper.DefaultForwardTransferPacketTimeoutTimestamp,
	)
	transferICS4Wrapper := transferStack.(porttypes.ICS4Wrapper)
	app.TransferKeeper.WithICS4Wrapper(transferICS4Wrapper)

	// Create ICAHost Stack
	var icaHostStack porttypes.IBCModule = icahost.NewIBCModule(app.ICAHostKeeper)

	// Create Interchain Accounts Controller Stack
	var icaControllerStack porttypes.IBCModule
	icaControllerStack = icacontroller.NewIBCMiddleware(app.ICAControllerKeeper)
	icaICS4Wrapper := icaControllerStack.(porttypes.ICS4Wrapper)
	// Since the callbacks middleware itself is an ics4wrapper, it needs to be passed to the ica controller keeper
	app.ICAControllerKeeper.WithICS4Wrapper(icaICS4Wrapper)

	// TODO: Re-enable WASM stack when interface compatibility is fixed
	// var wasmStack porttypes.IBCModule
	// wasmStack = wasm.NewIBCHandler(app.WasmKeeper, app.IBCKeeper.ChannelKeeper, app.TransferKeeper, bApp)
	// Create IBC Router & seal
	ibcRouter := porttypes.NewRouter().
		AddRoute(icahosttypes.SubModuleName, icaHostStack).
		AddRoute(icacontrollertypes.SubModuleName, icaControllerStack).
		AddRoute(ibctransfertypes.ModuleName, transferStack).
		// AddRoute(wasmTypes.ModuleName, wasmStack). // TODO: Re-enable when WASM compatibility is fixed
		AddRoute(ccvconsumertypes.ModuleName, app.ConsumerModule)

	app.IBCKeeper.SetRouter(ibcRouter)

	// register hooks after all modules have been initialized

	app.StablestakeKeeper.SetHooks(stablestakekeeper.NewMultiStableStakeHooks(
		app.MasterchefKeeper.StableStakeHooks(),
		app.TierKeeper.StableStakeHooks(),
	))

	app.LeveragelpKeeper.SetHooks(leveragelpmoduletypes.NewMultiLeverageLpHooks(
		// PerpetualKeeper.LeverageLpHooks() calling first because it needs to close all position before removing accounted pool
		app.PerpetualKeeper.LeverageLpHooks(),
		app.AccountedPoolKeeper.LeverageLpHooks(),
		app.TierKeeper.LeverageLpHooks(),
	))

	app.EstakingKeeper.SetHooks(
		stakingtypes.NewMultiStakingHooks(
			// insert staking hooks receivers here
			// Do not use slashing keeper hooks when it's consumer chain
			app.DistrKeeper.Hooks(),
			app.EstakingKeeper.StakingHooks(),
			app.TierKeeper.StakingHooks(),
		),
	)
	app.GovKeeper.SetHooks(
		govtypes.NewMultiGovHooks(
		// register the governance hooks
		),
	)

	app.AmmKeeper.SetHooks(
		ammmoduletypes.NewMultiAmmHooks(
			// insert amm hooks receivers here
			app.AccountedPoolKeeper.AmmHooks(),
			app.PerpetualKeeper.AmmHooks(),
			app.LeveragelpKeeper.AmmHooks(),
			app.MasterchefKeeper.AmmHooks(),
			app.TierKeeper.AmmHooks(),
		),
	)

	app.EpochsKeeper = app.EpochsKeeper.SetHooks(
		epochsmoduletypes.NewMultiEpochHooks(
			// insert epoch hooks receivers here
			app.CommitmentKeeper.Hooks(),
			app.BurnerKeeper.Hooks(),
			app.PerpetualKeeper.EpochHooks(),
			app.EstakingKeeper.EpochHooks(),
		),
	)

	app.PerpetualKeeper.SetHooks(
		perpetualmoduletypes.NewMultiPerpetualHooks(
			// insert perpetual hooks receivers here
			app.AccountedPoolKeeper.PerpetualHooks(),
			app.TierKeeper.PerpetualHooks(),
		),
	)

	return app

}

// GetSubspace returns a param subspace for a given module name.
func (appKeepers *AppKeepers) GetSubspace(moduleName string) paramstypes.Subspace {
	subspace, ok := appKeepers.ParamsKeeper.GetSubspace(moduleName)
	if !ok {
		panic("couldn't load subspace for module: " + moduleName)
	}
	return subspace
}

// initParamsKeeper init params keeper and its subspaces
func initParamsKeeper(appCodec codec.BinaryCodec, legacyAmino *codec.LegacyAmino, key, tkey storetypes.StoreKey) paramskeeper.Keeper {
	paramsKeeper := paramskeeper.NewKeeper(appCodec, legacyAmino, key, tkey)

	keyTable := ibcclienttypes.ParamKeyTable()
	keyTable.RegisterParamSet(&ibcconnectiontypes.Params{})
	paramsKeeper.Subspace(authtypes.ModuleName)
	paramsKeeper.Subspace(banktypes.ModuleName)
	paramsKeeper.Subspace(stakingtypes.ModuleName)
	paramsKeeper.Subspace(distrtypes.ModuleName)
	paramsKeeper.Subspace(slashingtypes.ModuleName)
	paramsKeeper.Subspace(govtypes.ModuleName).WithKeyTable(govv1.ParamKeyTable())
	paramsKeeper.Subspace(crisistypes.ModuleName)
	paramsKeeper.Subspace(ibcexported.ModuleName).WithKeyTable(keyTable)
	paramsKeeper.Subspace(ibctransfertypes.ModuleName).WithKeyTable(ibctransfertypes.ParamKeyTable())
	paramsKeeper.Subspace(icacontrollertypes.SubModuleName).WithKeyTable(icacontrollertypes.ParamKeyTable())
	paramsKeeper.Subspace(icahosttypes.SubModuleName).WithKeyTable(icahosttypes.ParamKeyTable())
	paramsKeeper.Subspace(ccvconsumertypes.ModuleName).WithKeyTable(ccv.ParamKeyTable())

	// EVM modules
	paramsKeeper.Subspace(evmtypes.ModuleName)
	paramsKeeper.Subspace(feemarkettypes.ModuleName)

	// Can be removed as we are not using param subspace anymore anywhere
	paramsKeeper.Subspace(assetprofilemoduletypes.ModuleName)
	//paramsKeeper.Subspace(oracletypes.ModuleName)
	paramsKeeper.Subspace(commitmentmoduletypes.ModuleName)
	paramsKeeper.Subspace(tokenomicsmoduletypes.ModuleName)
	paramsKeeper.Subspace(burnermoduletypes.ModuleName)
	paramsKeeper.Subspace(perpetualmoduletypes.ModuleName)
	paramsKeeper.Subspace(stablestaketypes.ModuleName)
	paramsKeeper.Subspace(leveragelpmoduletypes.ModuleName)
	paramsKeeper.Subspace(masterchefmoduletypes.ModuleName)
	paramsKeeper.Subspace(tiermoduletypes.ModuleName)
	// this line is used by starport scaffolding # stargate/app/paramSubspace

	return paramsKeeper
}
