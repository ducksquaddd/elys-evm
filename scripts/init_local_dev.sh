#!/bin/bash

set -e

echo "Setting up local Elys development chain..."

# Configuration
CHAIN_ID="elys-local-dev"
HOME_DIR="$HOME/.elys"
KEYRING_BACKEND="test"
VALIDATOR_NAME="local-validator"
USER_NAME="local-user"
DENOM="uelys"

# Clean up existing data
echo "Cleaning up existing chain data..."
rm -rf "$HOME_DIR"

# Initialize the chain
echo "Initializing chain..."
elysd init local-node --chain-id="$CHAIN_ID" --home="$HOME_DIR"

# Create validator key (using standard Cosmos secp256k1, not Ethereum keys)
echo "Creating validator key..."
elysd keys add "$VALIDATOR_NAME" --keyring-backend="$KEYRING_BACKEND" --home="$HOME_DIR" --algo secp256k1

# Create a user key for testing (standard Cosmos key)
echo "Creating user key..."
elysd keys add "$USER_NAME" --keyring-backend="$KEYRING_BACKEND" --home="$HOME_DIR" --algo secp256k1

# Get addresses (parse from human-readable output to avoid EVM logging contamination)
VALIDATOR_ADDR=$(elysd keys show "$VALIDATOR_NAME" --keyring-backend="$KEYRING_BACKEND" --home="$HOME_DIR" 2>/dev/null | grep "address:" | awk '{print $3}')
USER_ADDR=$(elysd keys show "$USER_NAME" --keyring-backend="$KEYRING_BACKEND" --home="$HOME_DIR" 2>/dev/null | grep "address:" | awk '{print $3}')

echo "Validator address: $VALIDATOR_ADDR"
echo "User address: $USER_ADDR"

# Add genesis accounts with tokens
echo "Adding genesis accounts with ELYS and USDC..."
elysd add-genesis-account "$VALIDATOR_ADDR" 100000000000${DENOM},1000000000uusdc --home="$HOME_DIR" --keyring-backend="$KEYRING_BACKEND"
elysd add-genesis-account "$USER_ADDR" 10000000${DENOM},100000000uusdc --home="$HOME_DIR" --keyring-backend="$KEYRING_BACKEND"

# Create gentx (genesis transaction for validator)
echo "Creating genesis transaction..."
elysd gentx "$VALIDATOR_NAME" 50000000000${DENOM} \
    --keyring-backend="$KEYRING_BACKEND" \
    --home="$HOME_DIR" \
    --chain-id="$CHAIN_ID" \
    --commission-rate="0.05" \
    --commission-max-rate="0.50" \
    --commission-max-change-rate="0.01" \
    --min-self-delegation="1000000"

# Collect genesis transactions
echo "Collecting genesis transactions..."
elysd collect-gentxs --home="$HOME_DIR"

# Fix denomination references from "stake" to "uelys" and configure stablestake
echo "Updating bond, governance, and stablestake denominations..."
GENESIS_FILE="$HOME_DIR/config/genesis.json"
jq '.app_state.staking.params.bond_denom = "uelys"' "$GENESIS_FILE" > "$HOME_DIR/config/genesis_tmp.json" && mv "$HOME_DIR/config/genesis_tmp.json" "$GENESIS_FILE"
jq '.app_state.crisis.constant_fee.denom = "uelys"' "$GENESIS_FILE" > "$HOME_DIR/config/genesis_tmp.json" && mv "$HOME_DIR/config/genesis_tmp.json" "$GENESIS_FILE"
jq '.app_state.gov.params.min_deposit[0].denom = "uelys"' "$GENESIS_FILE" > "$HOME_DIR/config/genesis_tmp.json" && mv "$HOME_DIR/config/genesis_tmp.json" "$GENESIS_FILE"
jq '.app_state.gov.params.expedited_min_deposit[0].denom = "uelys"' "$GENESIS_FILE" > "$HOME_DIR/config/genesis_tmp.json" && mv "$HOME_DIR/config/genesis_tmp.json" "$GENESIS_FILE"
jq '.app_state.stablestake.params.legacy_deposit_denom = "uelys"' "$GENESIS_FILE" > "$HOME_DIR/config/genesis_tmp.json" && mv "$HOME_DIR/config/genesis_tmp.json" "$GENESIS_FILE"

# Fix EVM denomination to match Cosmos SDK denom for unified balance experience
echo "Setting EVM denomination to uelys..."
jq '.app_state.evm.params.evm_denom = "uelys"' "$GENESIS_FILE" > "$HOME_DIR/config/genesis_tmp.json" && mv "$HOME_DIR/config/genesis_tmp.json" "$GENESIS_FILE"

# Configure active static precompiles (Berlin + custom ELYS precompiles)
echo "Activating EVM precompiles (9 Berlin + 2 ELYS custom)..."
jq '.app_state.evm.params.active_static_precompiles = [
    "0x0000000000000000000000000000000000000001",
    "0x0000000000000000000000000000000000000002", 
    "0x0000000000000000000000000000000000000003",
    "0x0000000000000000000000000000000000000004",
    "0x0000000000000000000000000000000000000005",
    "0x0000000000000000000000000000000000000006",
    "0x0000000000000000000000000000000000000007",
    "0x0000000000000000000000000000000000000008",
    "0x0000000000000000000000000000000000000009",
    "0x0000000000000000000000000000000000000804",
    "0x0000000000000000000000000000000000000806"
]' "$GENESIS_FILE" > "$HOME_DIR/config/genesis_tmp.json" && mv "$HOME_DIR/config/genesis_tmp.json" "$GENESIS_FILE"

# Add missing asset profiles that the masterchef module expects
# Note: We add USDC without wrapper address initially - it will be deployed after chain starts
echo "Adding required asset profiles (wrapper addresses will be set after deployment)..."
jq '.app_state.assetprofile.entry_list += [
  {
    "base_denom": "uusdc", 
    "denom": "uusdc",
    "display_name": "USDC",
    "display_symbol": "USDC",
    "decimals": "6",
    "address": "",
    "external_symbol": "",
    "transfer_limit": "",
    "permissions": [],
    "unit_denom": "",
    "ibc_channel_id": "",
    "ibc_counterparty_denom": "",
    "ibc_counterparty_chain_id": "",
    "network": "",
    "authority": "",
    "commit_enabled": true,
    "withdraw_enabled": true
  }
]' "$GENESIS_FILE" > "$HOME_DIR/config/genesis_tmp.json" && mv "$HOME_DIR/config/genesis_tmp.json" "$GENESIS_FILE"

# Fix slashing module - add the actual validator consensus address
echo "Configuring slashing module with proper validator signing info..."
VALIDATOR_CONSADDR=$(elysd tendermint show-address --home="$HOME_DIR" 2>/dev/null | tail -1)

# Add validator signing info with the correct consensus address
jq --arg consaddr "$VALIDATOR_CONSADDR" '.app_state.slashing.signing_infos += [
  {
    "address": $consaddr,
    "validator_signing_info": {
      "address": $consaddr,
      "start_height": "0",
      "index_offset": "0",
      "jailed_until": "1970-01-01T00:00:00Z",
      "tombstoned": false,
      "missed_blocks_counter": "0"
    }
  }
]' "$GENESIS_FILE" > "$HOME_DIR/config/genesis_tmp.json" && mv "$HOME_DIR/config/genesis_tmp.json" "$GENESIS_FILE"

# Disable slashing penalties for development
jq '.app_state.slashing.params.slash_fraction_double_sign = "0.000000000000000000"' "$GENESIS_FILE" > "$HOME_DIR/config/genesis_tmp.json" && mv "$HOME_DIR/config/genesis_tmp.json" "$GENESIS_FILE"
jq '.app_state.slashing.params.slash_fraction_downtime = "0.000000000000000000"' "$GENESIS_FILE" > "$HOME_DIR/config/genesis_tmp.json" && mv "$HOME_DIR/config/genesis_tmp.json" "$GENESIS_FILE"

# Configure governance for fast local development
echo "Configuring governance parameters for local development..."
jq '.app_state.gov.params.voting_period = "60s"' "$GENESIS_FILE" > "$HOME_DIR/config/genesis_tmp.json" && mv "$HOME_DIR/config/genesis_tmp.json" "$GENESIS_FILE"
jq '.app_state.gov.params.max_deposit_period = "30s"' "$GENESIS_FILE" > "$HOME_DIR/config/genesis_tmp.json" && mv "$HOME_DIR/config/genesis_tmp.json" "$GENESIS_FILE"
jq '.app_state.gov.params.expedited_voting_period = "30s"' "$GENESIS_FILE" > "$HOME_DIR/config/genesis_tmp.json" && mv "$HOME_DIR/config/genesis_tmp.json" "$GENESIS_FILE"

# Configure for development
echo "Configuring for development..."
CONFIG_FILE="$HOME_DIR/config/config.toml"
APP_FILE="$HOME_DIR/config/app.toml"
CLIENT_FILE="$HOME_DIR/config/client.toml"

# Update config.toml
sed -i.bak 's/cors_allowed_origins = \[\]/cors_allowed_origins = ["*"]/' "$CONFIG_FILE"
sed -i.bak 's/timeout_commit = "5s"/timeout_commit = "1s"/' "$CONFIG_FILE"

# Update app.toml
sed -i.bak 's/minimum-gas-prices = ""/minimum-gas-prices = "0.001uelys"/' "$APP_FILE"
sed -i.bak '/\[api\]/,/^enable = .*$/ s/^enable = .*$/enable = true/' "$APP_FILE"
sed -i.bak 's/unsafe-cors = false/unsafe-cors = true/' "$APP_FILE"

# Update client.toml
sed -i.bak "s/chain-id = \"\"/chain-id = \"$CHAIN_ID\"/" "$CLIENT_FILE"
sed -i.bak "s/keyring-backend = \"os\"/keyring-backend = \"$KEYRING_BACKEND\"/" "$CLIENT_FILE"

# Clean up backup files
rm -f "$CONFIG_FILE.bak" "$APP_FILE.bak" "$CLIENT_FILE.bak"

echo ""
echo "Local development chain initialized successfully!"
echo ""
echo "Chain ID: $CHAIN_ID"
echo "Validator: $VALIDATOR_NAME ($VALIDATOR_ADDR)"
echo "User: $USER_NAME ($USER_ADDR)"
echo ""
echo "To start the chain, run:"
echo "  elysd start"
echo ""
echo "To interact with the chain:"
echo "  elysd status"
echo "  elysd q bank balances $VALIDATOR_ADDR"
echo "  elysd q bank balances $USER_ADDR" 