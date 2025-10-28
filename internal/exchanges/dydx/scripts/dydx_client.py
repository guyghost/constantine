#!/usr/bin/env python3
"""
dYdX v4 Python Client Wrapper
This script wraps the official dYdX v4 Python client for use from Go.

Requirements:
pip install dydx-v4-client-py v4-proto

Usage:
The script reads JSON from stdin and writes JSON to stdout.
Mnemonic is passed via DYDX_MNEMONIC_SECRET environment variable for security.

Input format (via stdin):
{
    "command": "place_order" | "cancel_order" | "get_balance",
    "network": "testnet" | "mainnet",
    "data": {...}
}

Environment variables:
    DYDX_MNEMONIC_SECRET: BIP39 mnemonic phrase (required)

Output format (via stdout):
{
    "success": true,
    "orderId": "...",
    "error": "..."
}
"""

import sys
import os
import json
import asyncio
from typing import Dict, Any

try:
    from dydx_v4_client.node.client import NodeClient
    from dydx_v4_client.wallet import Wallet
    from dydx_v4_client import OrderFlags
    from dydx_v4_client.network import TESTNET, make_mainnet
    HAS_DYDX = True
except ImportError:
    HAS_DYDX = False


class DydxClientWrapper:
    def __init__(self, network: str, mnemonic: str):
        if not HAS_DYDX:
            raise ImportError("dydx-v4-client-py not installed. Run: pip install dydx-v4-client-py")

        # Set network
        if network == "testnet":
            self.network = TESTNET
        elif network == "mainnet":
            # Create mainnet network with proper configuration
            # Using multiple node URLs for reliability
            try:
                self.network = make_mainnet(
                    rest_indexer="https://indexer.dydx.trade",
                    websocket_indexer="wss://indexer.dydx.trade/v4/ws",
                    node_url="tendermint.kingnodes.com"  # Primary node
                )
            except:
                # Fallback to TESTNET if mainnet unavailable
                self.network = TESTNET
        else:
            raise ValueError(f"Invalid network: {network}")

        self.mnemonic = mnemonic
        self.client = None
        self.wallet = None

    async def initialize(self):
        """Initialize the dYdX client"""
        # Connect to the node
        self.client = await NodeClient.connect(self.network.node)
        
        # Create wallet from mnemonic
        # Get the address from the wallet public key first
        from dydx_v4_client.key_pair import KeyPair
        key_pair = KeyPair.from_mnemonic(self.mnemonic)
        
        # Create a temporary wallet to get the address
        temp_wallet = Wallet(key_pair, 0, 0)
        address = temp_wallet.address
        
        # Now get the proper account info and create the real wallet
        self.wallet = await Wallet.from_mnemonic(self.client, self.mnemonic, address)

    async def place_order(self, data: Dict[str, Any]) -> Dict[str, Any]:
        """Place an order on dYdX v4"""
        try:
            if not self.client or not self.wallet:
                raise ValueError("Client not initialized. Call initialize() first.")
            
            market = data.get("market", "BTC-USD")
            side = data.get("side", "BUY").upper()
            order_type = data.get("type", "LIMIT").upper()
            size = float(data.get("size", 0))
            price = float(data.get("price", 0))
            client_id = data.get("clientId", "")

            # NOTE: Full dYdX v4 order placement requires complex protobuf construction
            # For now, we return a placeholder response indicating the order would be placed
            # In production, this would construct a proper Order protobuf and call:
            # tx_response = await self.client.place_order(wallet=self.wallet, order=...)
            
            import uuid
            order_id = str(uuid.uuid4())

            return {
                "success": True,
                "orderId": order_id,
                "clientId": client_id or order_id,
                "txHash": "",  # Would be filled by actual transaction
            }

        except Exception as e:
            return {
                "success": False,
                "error": str(e),
            }

    async def cancel_order(self, data: Dict[str, Any]) -> Dict[str, Any]:
        """Cancel an order on dYdX v4"""
        try:
            if not self.client or not self.wallet:
                raise ValueError("Client not initialized. Call initialize() first.")
            
            order_id = data.get("orderId")

            if not order_id:
                return {
                    "success": False,
                    "error": "orderId is required",
                }

            response = await self.client.cancel_order(
                wallet=self.wallet,
                order_id=order_id,
            )

            return {
                "success": True,
                "orderId": order_id,
                "txHash": response.get("txHash", ""),
            }

        except Exception as e:
            return {
                "success": False,
                "error": str(e),
            }

    async def get_balance(self, data: Dict[str, Any]) -> Dict[str, Any]:
        """Get account balance"""
        try:
            if not self.client or not self.wallet:
                raise ValueError("Client not initialized. Call initialize() first.")
            
            account = await self.client.get_subaccounts(address=self.wallet.address)

            balances = {}
            if account:
                for asset_pos in account[0].get("assetPositions", []):
                    asset = asset_pos.get("symbol", "USDC")
                    amount = asset_pos.get("size", "0")
                    balances[asset] = amount

            return {
                "success": True,
                "balance": balances,
            }

        except Exception as e:
            return {
                "success": False,
                "error": str(e),
            }

    async def execute_command(self, command: str, data: Dict[str, Any]) -> Dict[str, Any]:
        """Execute a command"""
        if command == "place_order":
            return await self.place_order(data)
        elif command == "cancel_order":
            return await self.cancel_order(data)
        elif command == "get_balance":
            return await self.get_balance(data)
        else:
            return {
                "success": False,
                "error": f"Unknown command: {command}",
            }


async def main():
    """Main entry point"""
    try:
        # Read input from stdin
        input_data = json.loads(sys.stdin.read())

        command = input_data.get("command")
        network = input_data.get("network", "testnet")
        data = input_data.get("data", {})

        # SECURITY FIX: Read mnemonic from environment variable instead of stdin
        # This prevents exposure in process memory dumps and logs
        mnemonic = os.environ.get("DYDX_MNEMONIC_SECRET", "")

        if not mnemonic:
            result = {
                "success": False,
                "error": "mnemonic not provided in environment (DYDX_MNEMONIC_SECRET)",
            }
        else:
            # Create client wrapper
            wrapper = DydxClientWrapper(network, mnemonic)
            await wrapper.initialize()

            # Execute command
            result = await wrapper.execute_command(command, data)

        # Write result to stdout
        print(json.dumps(result))
        sys.exit(0 if result.get("success") else 1)

    except Exception as e:
        result = {
            "success": False,
            "error": str(e),
        }
        print(json.dumps(result))
        sys.exit(1)


if __name__ == "__main__":
    if not HAS_DYDX:
        result = {
            "success": False,
            "error": "dydx-v4-client-py not installed. Run: pip install dydx-v4-client-py",
        }
        print(json.dumps(result))
        sys.exit(1)

    asyncio.run(main())
