#!/usr/bin/env python3
"""
dYdX v4 Python Client Wrapper
This script wraps the official dYdX v4 Python client for use from Go.

Requirements:
pip install dydx-v4-client-py v4-proto

Usage:
The script reads JSON from stdin and writes JSON to stdout.

Input format:
{
    "command": "place_order" | "cancel_order" | "get_balance",
    "network": "testnet" | "mainnet",
    "mnemonic": "your mnemonic phrase...",
    "data": {...}
}

Output format:
{
    "success": true,
    "orderId": "...",
    "error": "..."
}
"""

import sys
import json
import asyncio
from typing import Dict, Any

try:
    from v4_client_py import NodeClient, OrderFlags
    from v4_client_py.clients.constants import Network
    HAS_DYDX = True
except ImportError:
    HAS_DYDX = False


class DydxClientWrapper:
    def __init__(self, network: str, mnemonic: str):
        if not HAS_DYDX:
            raise ImportError("dydx-v4-client-py not installed. Run: pip install dydx-v4-client-py")

        # Set network
        if network == "testnet":
            self.network = Network.testnet()
        elif network == "mainnet":
            self.network = Network.mainnet()
        else:
            raise ValueError(f"Invalid network: {network}")

        self.mnemonic = mnemonic
        self.client = None

    async def initialize(self):
        """Initialize the dYdX client"""
        self.client = await NodeClient.connect(self.network.node())
        await self.client.from_mnemonic(self.mnemonic)

    async def place_order(self, data: Dict[str, Any]) -> Dict[str, Any]:
        """Place an order on dYdX v4"""
        try:
            market = data.get("market", "BTC-USD")
            side = data.get("side", "BUY")
            order_type = data.get("type", "LIMIT")
            size = float(data.get("size", 0))
            price = float(data.get("price", 0))
            time_in_force = data.get("timeInForce", "GTT")
            reduce_only = data.get("reduceOnly", False)
            post_only = data.get("postOnly", False)
            client_id = data.get("clientId")

            # Determine order flags
            if order_type == "MARKET":
                order_flags = OrderFlags.SHORT_TERM
            else:
                order_flags = OrderFlags.LONG_TERM

            # Place the order
            response = await self.client.place_order(
                market=market,
                side=side.upper(),
                order_type=order_type.upper(),
                size=size,
                price=price,
                time_in_force=time_in_force,
                reduce_only=reduce_only,
                post_only=post_only,
                client_id=int(client_id) if client_id else None,
                order_flags=order_flags,
            )

            return {
                "success": True,
                "orderId": str(response.get("order", {}).get("id", "")),
                "clientId": str(response.get("order", {}).get("clientId", "")),
                "txHash": response.get("hash", ""),
            }

        except Exception as e:
            return {
                "success": False,
                "error": str(e),
            }

    async def cancel_order(self, data: Dict[str, Any]) -> Dict[str, Any]:
        """Cancel an order on dYdX v4"""
        try:
            order_id = data.get("orderId")

            if not order_id:
                return {
                    "success": False,
                    "error": "orderId is required",
                }

            response = await self.client.cancel_order(order_id)

            return {
                "success": True,
                "orderId": order_id,
                "txHash": response.get("hash", ""),
            }

        except Exception as e:
            return {
                "success": False,
                "error": str(e),
            }

    async def get_balance(self, data: Dict[str, Any]) -> Dict[str, Any]:
        """Get account balance"""
        try:
            account = await self.client.get_account()

            balances = {}
            for balance in account.get("subaccounts", [{}])[0].get("assetPositions", []):
                asset = balance.get("symbol", "USDC")
                amount = balance.get("size", "0")
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
        mnemonic = input_data.get("mnemonic", "")
        data = input_data.get("data", {})

        if not mnemonic:
            result = {
                "success": False,
                "error": "mnemonic is required",
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
